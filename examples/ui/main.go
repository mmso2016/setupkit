package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mmso2016/setupkit/installer"
	_ "github.com/mmso2016/setupkit/installer/ui" // Register UI factory
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct for Wails bindings
type App struct {
	ctx       context.Context
	installer *installer.Installer
	config    *AppConfig
	running   bool
}

// AppConfig represents the configuration for the frontend
type AppConfig struct {
	AppName     string          `json:"appName"`
	Version     string          `json:"version"`
	Publisher   string          `json:"publisher"`
	Website     string          `json:"website"`
	License     string          `json:"license"`
	Components  []ComponentInfo `json:"components"`
	InstallPath string          `json:"installPath"`
	Theme       string          `json:"theme"`
}

// ComponentInfo represents a component for the frontend
type ComponentInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Selected    bool   `json:"selected"`
	Size        int64  `json:"size"`
}

// InstallationConfig from frontend
type InstallationConfig struct {
	Path       string   `json:"path"`
	Components []string `json:"components"`
}

// Global variables for configuration
var (
	appTitle     string
	appVersion   string
	publisher    string
	configFile   string
	themeName    string
	genConfig    string
	listThemes   bool
	verbose      bool
	mode         string
	responseFile string
	logFile      string
)

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Initialize installer configuration
	a.initializeInstaller()
}

// initializeInstaller sets up the installer configuration
func (a *App) initializeInstaller() {
	// Get default installation directory
	installDir := getDefaultInstallDir(appTitle)
	
	// Build installer options
	opts := []installer.Option{
		installer.WithAppName(appTitle),
		installer.WithVersion(appVersion),
		installer.WithPublisher(publisher),
		installer.WithMode(installer.ModeGUI),
		installer.WithInstallDir(installDir),
		installer.WithVerbose(verbose),
		installer.WithLicense(getLicenseText()),
		installer.WithComponents(
			installer.Component{
				ID:          "core",
				Name:        "Core Files",
				Description: "Required application files",
				Required:    true,
				Selected:    true,
				Size:        10485760, // 10 MB
				Installer:   installCore,
			},
			installer.Component{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and help files",
				Required:    false,
				Selected:    true,
				Size:        5242880, // 5 MB
				Installer:   installDocs,
			},
			installer.Component{
				ID:          "examples",
				Name:        "Examples",
				Description: "Sample projects and templates",
				Required:    false,
				Selected:    false,
				Size:        2097152, // 2 MB
				Installer:   installExamples,
			},
		),
	}
	
	// Add response file support
	if responseFile != "" {
		opts = append(opts, installer.WithResponseFile(responseFile))
	}
	
	// Apply UI configuration if specified
	if configFile != "" {
		opts = append(opts, installer.WithUIConfig(configFile))
		log.Printf("Using UI configuration: %s", configFile)
	} else if themeName != "" && themeName != "default" {
		opts = append(opts, installer.WithTheme(themeName))
		log.Printf("Using theme: %s", themeName)
	}
	
	// Create installer
	inst, err := installer.New(opts...)
	if err != nil {
		log.Printf("Failed to create installer: %v", err)
		return
	}
	
	a.installer = inst
	
	// Build app config from installer config
	installerConfig := inst.GetConfig()
	
	components := make([]ComponentInfo, len(installerConfig.Components))
	for i, comp := range installerConfig.Components {
		components[i] = ComponentInfo{
			ID:          comp.ID,
			Name:        comp.Name,
			Description: comp.Description,
			Required:    comp.Required,
			Selected:    comp.Selected,
			Size:        comp.Size,
		}
	}
	
	a.config = &AppConfig{
		AppName:     installerConfig.AppName,
		Version:     installerConfig.Version,
		Publisher:   installerConfig.Publisher,
		Website:     installerConfig.Website,
		License:     installerConfig.License,
		Components:  components,
		InstallPath: installerConfig.InstallDir,
		Theme:       themeName,
	}
}

// GetConfig returns the installation configuration for the frontend
func (a *App) GetConfig() *AppConfig {
	return a.config
}

// BrowseFolder opens a directory selection dialog
func (a *App) BrowseFolder() string {
	path, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:            "Select Installation Directory",
		DefaultDirectory: a.config.InstallPath,
	})
	
	if err != nil {
		wailsruntime.LogError(a.ctx, fmt.Sprintf("Failed to open directory dialog: %v", err))
		return ""
	}
	
	return path
}

// SetSelectedComponents sets the selected components
func (a *App) SetSelectedComponents(componentIDs []string) {
	// Update configuration
	for i := range a.config.Components {
		a.config.Components[i].Selected = false
		for _, id := range componentIDs {
			if a.config.Components[i].ID == id {
				a.config.Components[i].Selected = true
				break
			}
		}
	}
	
	// Update installer components
	if a.installer != nil {
		installerConfig := a.installer.GetConfig()
		for i := range installerConfig.Components {
			installerConfig.Components[i].Selected = false
			for _, id := range componentIDs {
				if installerConfig.Components[i].ID == id {
					installerConfig.Components[i].Selected = true
					break
				}
			}
		}
	}
}

// SetInstallPath sets the installation path
func (a *App) SetInstallPath(path string) error {
	if path == "" {
		return fmt.Errorf("installation path cannot be empty")
	}
	
	a.config.InstallPath = path
	if a.installer != nil {
		a.installer.SetInstallPath(path)
	}
	return nil
}

// StartInstallation begins the installation process
func (a *App) StartInstallation(config InstallationConfig) error {
	if a.installer == nil {
		return fmt.Errorf("installer not initialized")
	}
	
	if a.running {
		return fmt.Errorf("installation already in progress")
	}
	
	// Update configuration
	a.SetInstallPath(config.Path)
	a.SetSelectedComponents(config.Components)
	
	// Run installation in goroutine to not block UI
	go func() {
		a.running = true
		defer func() { a.running = false }()
		
		// Emit start event
		wailsruntime.EventsEmit(a.ctx, "installation-progress", map[string]interface{}{
			"percent": 0,
			"status":  "Preparing installation...",
		})
		
		// Get selected components
		totalComponents := 0
		for _, comp := range a.config.Components {
			if comp.Selected || comp.Required {
				totalComponents++
			}
		}
		
		if totalComponents == 0 {
			wailsruntime.EventsEmit(a.ctx, "installation-error", map[string]interface{}{
				"message": "No components selected",
			})
			return
		}
		
		// Install components
		installed := 0
		installerConfig := a.installer.GetConfig()
		for _, comp := range installerConfig.Components {
			if !comp.Selected && !comp.Required {
				continue
			}
			
			// Update progress
			percent := (installed * 100) / totalComponents
			wailsruntime.EventsEmit(a.ctx, "installation-progress", map[string]interface{}{
				"percent": percent,
				"status":  fmt.Sprintf("Installing %s...", comp.Name),
			})
			
			// Run component installer
			if comp.Installer != nil {
				ctx := context.Background()
				if err := comp.Installer(ctx); err != nil {
					wailsruntime.EventsEmit(a.ctx, "installation-error", map[string]interface{}{
						"message": fmt.Sprintf("Failed to install %s: %v", comp.Name, err),
					})
					return
				}
			} else {
				// Simulate if no installer provided
				time.Sleep(2 * time.Second)
			}
			
			installed++
		}
		
		// Final progress
		wailsruntime.EventsEmit(a.ctx, "installation-progress", map[string]interface{}{
			"percent": 100,
			"status":  "Installation complete!",
		})
		
		// Build completion info
		var componentNames []string
		for _, comp := range a.config.Components {
			if comp.Selected || comp.Required {
				componentNames = append(componentNames, comp.Name)
			}
		}
		
		// Emit completion
		wailsruntime.EventsEmit(a.ctx, "installation-complete", map[string]interface{}{
			"success":     true,
			"installPath": a.config.InstallPath,
			"components":  componentNames,
			"duration":    "2 minutes", // You can track actual duration
		})
	}()
	
	return nil
}

// FinishInstallation completes the installation and optionally launches the app
func (a *App) FinishInstallation(launchApp bool, viewReadme bool) error {
	if launchApp && a.config != nil {
		wailsruntime.LogInfo(a.ctx, "Would launch application from: "+a.config.InstallPath)
		// TODO: Implement launching the installed application
	}
	
	if viewReadme {
		wailsruntime.LogInfo(a.ctx, "Would open README file")
		// TODO: Implement opening README
	}
	
	// Exit the installer
	wailsruntime.Quit(a.ctx)
	return nil
}

// ExitInstaller exits the application
func (a *App) ExitInstaller() {
	wailsruntime.Quit(a.ctx)
}

// Component installation functions
func installCore(ctx context.Context) error {
	log.Println("Installing core files...")
	time.Sleep(2 * time.Second) // Simulate installation
	return nil
}

func installDocs(ctx context.Context) error {
	log.Println("Installing documentation...")
	time.Sleep(1 * time.Second) // Simulate installation
	return nil
}

func installExamples(ctx context.Context) error {
	log.Println("Installing examples...")
	time.Sleep(1 * time.Second) // Simulate installation
	return nil
}

func getDefaultInstallDir(appName string) string {
	// Remove spaces and special characters for directory name
	dirName := appName
	
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("PROGRAMFILES"), dirName)
	case "darwin":
		return filepath.Join("/Applications", dirName+".app")
	default:
		return filepath.Join("/opt", dirName)
	}
}

func getLicenseText() string {
	return `MIT License

Copyright (c) 2024 Example Publisher

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`
}

func listAvailableThemes() {
	log.Println("Available themes:")
	themes := installer.ListThemes()
	for _, theme := range themes {
		log.Printf("  %-20s - %s", theme.Name, theme.Description)
	}
}

func generateConfig(filename, appName string) {
	log.Printf("Generating default configuration: %s", filename)
	if err := installer.GenerateConfigFile(filename, appName); err != nil {
		log.Fatal("Failed to generate config:", err)
	}
	log.Println("Configuration file generated successfully!")
	log.Println("You can now customize it and use with: -config", filename)
}

func main() {
	// Command line flags
	flag.StringVar(&appTitle, "title", "My Application", "Application title")
	flag.StringVar(&appVersion, "version", "1.0.0", "Application version")
	flag.StringVar(&publisher, "publisher", "Example Publisher", "Publisher name")
	flag.StringVar(&mode, "mode", "gui", "UI mode: gui, cli, silent")
	flag.StringVar(&responseFile, "response", "", "Response file for silent installation")
	flag.StringVar(&logFile, "log", "", "Log file for installation")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.StringVar(&configFile, "config", "", "UI configuration file (YAML)")
	flag.StringVar(&themeName, "theme", "default", "Theme name (default, corporate-blue, medical-green, tech-dark, minimal-light)")
	flag.StringVar(&genConfig, "generate-config", "", "Generate default config file and exit")
	flag.BoolVar(&listThemes, "list-themes", false, "List available themes and exit")
	flag.Parse()
	
	// Handle utility commands
	if listThemes {
		listAvailableThemes()
		return
	}
	
	if genConfig != "" {
		generateConfig(genConfig, appTitle)
		return
	}
	
	// If not GUI mode, fall back to CLI mode
	if mode != "gui" {
		runCLIMode()
		return
	}
	
	// Create an instance of the app structure
	app := NewApp()
	
	// Create application with options
	err := wails.Run(&options.App{
		Title:  appTitle + " Setup",
		Width:  900,
		Height: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})
	
	if err != nil {
		log.Fatal("Error:", err)
	}
}

// runCLIMode runs the installer in CLI mode (fallback)
func runCLIMode() {
	log.Println("Running in CLI mode...")
	
	// Parse UI mode
	var uiMode installer.Mode
	switch mode {
	case "cli":
		uiMode = installer.ModeCLI
	case "silent":
		uiMode = installer.ModeSilent
	default:
		uiMode = installer.ModeCLI
	}
	
	// Get default installation directory
	installDir := getDefaultInstallDir(appTitle)
	
	// Build installer options
	opts := []installer.Option{
		installer.WithAppName(appTitle),
		installer.WithVersion(appVersion),
		installer.WithPublisher(publisher),
		installer.WithMode(uiMode),
		installer.WithInstallDir(installDir),
		installer.WithVerbose(verbose),
		installer.WithLicense(getLicenseText()),
		installer.WithComponents(
			installer.Component{
				ID:          "core",
				Name:        "Core Files",
				Description: "Required application files",
				Required:    true,
				Selected:    true,
				Size:        10485760, // 10 MB
				Installer:   installCore,
			},
			installer.Component{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and help files",
				Required:    false,
				Selected:    true,
				Size:        5242880, // 5 MB
				Installer:   installDocs,
			},
			installer.Component{
				ID:          "examples",
				Name:        "Examples",
				Description: "Sample projects and templates",
				Required:    false,
				Selected:    false,
				Size:        2097152, // 2 MB
				Installer:   installExamples,
			},
		),
	}
	
	// Add response file support
	if responseFile != "" {
		opts = append(opts, installer.WithResponseFile(responseFile))
	}
	
	// Apply UI configuration if specified
	if configFile != "" {
		opts = append(opts, installer.WithUIConfig(configFile))
	} else if themeName != "" {
		opts = append(opts, installer.WithTheme(themeName))
	}
	
	// Create installer
	inst, err := installer.New(opts...)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}
	
	// Run installer
	ctx := context.Background()
	if err := inst.RunWithContext(ctx); err != nil {
		log.Fatalf("Installation failed: %v", err)
	}
	
	fmt.Println("Installation completed successfully!")
}

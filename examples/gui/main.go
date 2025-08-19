package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"path/filepath"
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
	config    *installer.Config
	running   bool
	themeCSS  string
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

// InstallConfig represents the installation configuration for frontend
type InstallConfig struct {
	AppName     string          `json:"appName"`
	Version     string          `json:"version"`
	Publisher   string          `json:"publisher"`
	Website     string          `json:"website"`
	License     string          `json:"license"`
	Components  []ComponentInfo `json:"components"`
	InstallPath string          `json:"installPath"`
	Theme       ThemeInfo       `json:"theme"`
	ThemeCSS    string          `json:"themeCSS"`
}

// ThemeInfo represents theme information for the frontend
type ThemeInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InstallationConfig from frontend
type InstallationConfig struct {
	Path       string   `json:"path"`
	Components []string `json:"components"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Get command line options from global variables
	opts := getInstallerOptions()

	// Create installer with theme support
	if inst, err := installer.New(opts...); err != nil {
		log.Printf("Failed to create installer: %v", err)
		return
	} else {
		a.installer = inst
		a.config = inst.GetConfig()
	}

	// Get theme CSS for the frontend
	a.themeCSS = a.config.GetThemeCSS()
}

// GetInstallConfig returns the installation configuration for the frontend
func (a *App) GetInstallConfig() *InstallConfig {
	if a.config == nil {
		return nil
	}

	components := make([]ComponentInfo, len(a.config.Components))
	for i, comp := range a.config.Components {
		components[i] = ComponentInfo{
			ID:          comp.ID,
			Name:        comp.Name,
			Description: comp.Description,
			Required:    comp.Required,
			Selected:    comp.Selected,
			Size:        comp.Size,
		}
	}

	// Get theme information
	themeInfo := a.config.GetThemeInfo()

	return &InstallConfig{
		AppName:     a.config.AppName,
		Version:     a.config.Version,
		Publisher:   a.config.Publisher,
		Website:     a.config.Website,
		License:     a.config.License,
		Components:  components,
		InstallPath: a.config.InstallDir,
		Theme:       ThemeInfo{Name: themeInfo.Name, Description: themeInfo.Description},
		ThemeCSS:    a.themeCSS,
	}
}

// BrowseFolder opens a directory selection dialog
func (a *App) BrowseFolder() string {
	path, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:            "Select Installation Directory",
		DefaultDirectory: a.config.InstallDir,
	})

	if err != nil {
		wailsruntime.LogError(a.ctx, fmt.Sprintf("Failed to open directory dialog: %v", err))
		return ""
	}

	return path
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
	a.config.InstallDir = config.Path

	// Update selected components
	for i := range a.config.Components {
		a.config.Components[i].Selected = false
		for _, id := range config.Components {
			if a.config.Components[i].ID == id {
				a.config.Components[i].Selected = true
				break
			}
		}
	}

	// Run installation in goroutine to not block UI
	go func() {
		a.running = true
		defer func() { a.running = false }()

		// Emit start event
		wailsruntime.EventsEmit(a.ctx, "installation-progress", map[string]interface{}{
			"percent": 0,
			"status":  "Preparing installation...",
		})

		// Simulate installation by running component installers
		totalComponents := 0
		for _, comp := range a.config.Components {
			if comp.Selected {
				totalComponents++
			}
		}

		if totalComponents == 0 {
			wailsruntime.EventsEmit(a.ctx, "installation-error", map[string]interface{}{
				"message": "No components selected",
			})
			return
		}

		installed := 0
		for _, comp := range a.config.Components {
			if !comp.Selected {
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

		// Emit completion
		wailsruntime.EventsEmit(a.ctx, "installation-complete", nil)
	}()

	return nil
}

// Global variables for CLI flags
var (
	configFile string
	themeName  string
	genConfig  string
	listThemes bool
)

// getInstallerOptions creates installer options based on CLI flags and defaults
func getInstallerOptions() []installer.Option {
	opts := []installer.Option{
		installer.WithAppName("Go SetupKit GUI Example"),
		installer.WithVersion("1.0.0"),
		installer.WithPublisher("Go SetupKit Team"),
		installer.WithWebsite("https://github.com/mmso2016/setupkit"),
		installer.WithMode(installer.ModeGUI),
		installer.WithInstallDir(filepath.Join("C:", "Program Files", "SetupKit-Sample")),
		installer.WithLicense(`MIT License

Copyright (c) 2025 Go SetupKit

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
SOFTWARE.`),
		installer.WithComponents(
			installer.Component{
				ID:          "core",
				Name:        "Core Files",
				Description: "Required application files",
				Required:    true,
				Selected:    true,
				Size:        10485760, // 10 MB
				Installer: func(ctx context.Context) error {
					// Simulate installation with progress events
					time.Sleep(time.Second)
					return nil
				},
			},
			installer.Component{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and API documentation",
				Required:    false,
				Selected:    true,
				Size:        5242880, // 5 MB
				Installer: func(ctx context.Context) error {
					// Simulate installation
					time.Sleep(time.Second)
					return nil
				},
			},
			installer.Component{
				ID:          "examples",
				Name:        "Examples",
				Description: "Sample projects and code",
				Required:    false,
				Selected:    false,
				Size:        2097152, // 2 MB
				Installer: func(ctx context.Context) error {
					// Simulate installation
					time.Sleep(time.Second)
					return nil
				},
			},
		),
		installer.WithVerbose(true),
	}

	// Apply UI configuration if specified
	if configFile != "" {
		opts = append(opts, installer.WithUIConfig(configFile))
		log.Printf("Using UI configuration: %s", configFile)
	} else if themeName != "" {
		// Apply theme directly if no config file
		opts = append(opts, installer.WithTheme(themeName))
		log.Printf("Using theme: %s", themeName)
	} else {
		// Default theme for GUI
		opts = append(opts, installer.WithTheme("default"))
	}

	return opts
}

func listAvailableThemes() {
	log.Println("Available themes:")
	themes := installer.ListThemes()
	for _, theme := range themes {
		log.Printf("  %-20s - %s", theme.Name, theme.Description)
	}
}

func generateConfig(filename string) {
	log.Printf("Generating default configuration: %s", filename)
	if err := installer.GenerateConfigFile(filename, "Go Installer GUI Example"); err != nil {
		log.Fatal("Failed to generate config:", err)
	}
	log.Println("Configuration file generated successfully!")
	log.Println("You can now customize it and use with: -config", filename)
}

// FinishInstallation completes the installation and optionally launches the app
func (a *App) FinishInstallation(launchApp bool) error {
	if launchApp && a.config != nil {
		// TODO: Implement launching the installed application
		wailsruntime.LogInfo(a.ctx, "Would launch application from: "+a.config.InstallDir)
	}

	// Give a moment for any UI updates
	wailsruntime.EventsEmit(a.ctx, "closing", nil)

	// Exit the installer
	wailsruntime.Quit(a.ctx)

	return nil
}

func main() {
	// Command line flags for UI configuration
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
		generateConfig(genConfig)
		return
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Go Installer GUI",
		Width:  800,
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
		fmt.Printf("Error: %v\n", err)
	}
}

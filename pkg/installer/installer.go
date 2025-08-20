package installer

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/mmso2016/setupkit/internal/ui"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Config represents the installer configuration
type Config struct {
	// Basic Info
	AppName   string
	Version   string
	Publisher string
	Website   string
	License   string

	// Installation
	InstallPath string
	Components  []Component

	// UI Customization
	Theme           Theme
	WindowWidth     int
	WindowHeight    int
	ShowThemeSelect bool

	// Callbacks
	OnProgress    func(percent int, status string)
	OnComplete    func(path string)
	OnError       func(error)
	BeforeInstall func() error
	AfterInstall  func() error
}

// Component represents an installable component
type Component struct {
	ID          string
	Name        string
	Description string
	Size        int64 // in bytes
	Required    bool
	Selected    bool
	Files       []string
}

// Theme allows UI customization
type Theme struct {
	Name           string
	PrimaryColor   string
	SecondaryColor string
	Logo           []byte
	CustomCSS      string
}

// Installer is the main installer instance
type Installer struct {
	config *Config
	app    *App
}

// App struct for Wails backend
type App struct {
	ctx    context.Context
	config *Config
}

// NewContext sets the context for the app
func (a *App) NewContext(ctx context.Context) {
	a.ctx = ctx
}

// GetConfig returns the installer configuration to the frontend
func (a *App) GetConfig() map[string]interface{} {
	components := make([]map[string]interface{}, len(a.config.Components))
	for i, c := range a.config.Components {
		components[i] = map[string]interface{}{
			"id":          c.ID,
			"name":        c.Name,
			"description": c.Description,
			"size":        c.Size,
			"required":    c.Required,
			"selected":    c.Selected,
		}
	}

	return map[string]interface{}{
		"appName":     a.config.AppName,
		"version":     a.config.Version,
		"publisher":   a.config.Publisher,
		"website":     a.config.Website,
		"license":     a.config.License,
		"installPath": a.config.InstallPath,
		"components":  components,
	}
}

// StartInstallation begins the installation process
func (a *App) StartInstallation(options map[string]interface{}) error {
	// Run pre-install callback
	if a.config.BeforeInstall != nil {
		if err := a.config.BeforeInstall(); err != nil {
			return err
		}
	}

	// TODO: Actual installation logic
	// This would copy files, create registry entries, etc.

	// Simulate progress
	for i := 0; i <= 100; i += 10 {
		if a.config.OnProgress != nil {
			a.config.OnProgress(i, fmt.Sprintf("Installing... %d%%", i))
		}
	}

	// Run post-install callback
	if a.config.AfterInstall != nil {
		if err := a.config.AfterInstall(); err != nil {
			return err
		}
	}

	// Call completion callback
	if a.config.OnComplete != nil {
		a.config.OnComplete(a.config.InstallPath)
	}

	return nil
}

// FinishInstallation completes the installation and exits the application
func (a *App) FinishInstallation(launchApp, viewReadme bool) error {
	// Handle post-installation actions
	if launchApp {
		// TODO: Launch the installed application if requested
		fmt.Printf("Launching application...\n")
	}

	if viewReadme {
		// TODO: Open README or documentation if requested
		fmt.Printf("Opening documentation...\n")
	}

	// Exit the application
	if a.ctx != nil {
		// Try to exit gracefully through Wails runtime
		select {
		case <-a.ctx.Done():
			return nil
		default:
			// Force exit if context is not cancelled
			os.Exit(0)
		}
	} else {
		os.Exit(0)
	}

	return nil
}

// ExitInstaller exits the installer application immediately
func (a *App) ExitInstaller() error {
	os.Exit(0)
	return nil
}

// New creates a new installer instance
func New(config *Config) *Installer {
	// Set defaults
	if config.WindowWidth == 0 {
		config.WindowWidth = 900
	}
	if config.WindowHeight == 0 {
		config.WindowHeight = 700
	}
	if config.InstallPath == "" {
		config.InstallPath = fmt.Sprintf("C:\\Program Files\\%s", config.AppName)
	}

	return &Installer{
		config: config,
		app: &App{
			config: config,
		},
	}
}

// Run starts the installer UI
func (i *Installer) Run() error {
	// Generate UI HTML with template rendering
	uiConfig := &ui.Config{
		AppName:         i.config.AppName,
		Version:         i.config.Version,
		Publisher:       i.config.Publisher,
		Website:         i.config.Website,
		License:         i.config.License,
		InstallPath:     i.config.InstallPath,
		ShowThemeSelect: i.config.ShowThemeSelect,
	}

	// Convert components
	for _, c := range i.config.Components {
		uiConfig.Components = append(uiConfig.Components, ui.Component{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Size:        c.Size,
			Required:    c.Required,
			Selected:    c.Selected,
		})
	}

	// Apply theme
	if i.config.Theme.CustomCSS != "" {
		uiConfig.Theme.CustomCSS = i.config.Theme.CustomCSS
	}

	html, err := ui.GenerateHTML(uiConfig)
	if err != nil {
		return fmt.Errorf("failed to generate UI: %w", err)
	}

	// Create Wails app with rendered HTML
	return wails.Run(&options.App{
		Title:            i.config.AppName + " Setup",
		Width:            i.config.WindowWidth,
		Height:           i.config.WindowHeight,
		MinWidth:         800,
		MinHeight:        600,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        i.app.NewContext,
		Bind: []interface{}{
			i.app,
		},
		// Serve the rendered HTML with Wails runtime support
		AssetServer: &assetserver.Options{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/" || r.URL.Path == "/index.html" {
					w.Header().Set("Content-Type", "text/html")
					w.Write([]byte(html))
					return
				}
				// Let Wails handle other assets like runtime.js
				http.NotFound(w, r)
			}),
		},
	})
}

// Quick function for simple installations
func Install(config Config) error {
	installer := New(&config)
	return installer.Run()
}

// Run starts the installer with the given configuration
func Run(config *Config) error {
	installer := New(config)
	return installer.Run()
}

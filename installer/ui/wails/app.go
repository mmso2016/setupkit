//go:build wails
// +build wails

package wails

import (
	"context"
	"embed"
	"fmt"
	"sync"

	"github.com/mmso2016/setupkit/installer/core"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// WailsUI implements the GUI using Wails framework
type WailsUI struct {
	ctx       context.Context
	coreCtx   *core.Context
	installer *core.Installer
	app       *App

	// Channels for communication with frontend
	ready    chan bool
	shutdown chan bool
}

// App is the struct bound to Wails frontend
type App struct {
	ctx       context.Context
	installer *core.Installer
	config    *core.Config

	// Current state
	currentStep     string
	selectedComps   []string
	installPath     string
	licenseAccepted bool

	// Channels for async operations
	progressChan chan *core.Progress
	errorChan    chan error
	mu           sync.RWMutex
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
}

// New creates a new Wails UI instance
func New() *WailsUI {
	return &WailsUI{
		ready:    make(chan bool, 1),
		shutdown: make(chan bool, 1),
	}
}

// Initialize prepares the UI
func (w *WailsUI) Initialize(ctx *core.Context) error {
	w.coreCtx = ctx

	// Get installer from context
	if installer, ok := ctx.Metadata["installer"].(*core.Installer); ok {
		w.installer = installer
	} else {
		return fmt.Errorf("installer not found in context")
	}

	// Create app instance
	w.app = &App{
		installer:    w.installer,
		config:       ctx.Config,
		progressChan: make(chan *core.Progress, 100),
		errorChan:    make(chan error, 10),
		currentStep:  "welcome",
		installPath:  ctx.Config.InstallDir,
	}

	// Set default install path if empty
	if w.app.installPath == "" {
		w.app.installPath = fmt.Sprintf("C:\\Program Files\\%s", ctx.Config.AppName)
	}

	return nil
}

// Run starts the Wails application
func (w *WailsUI) Run() error {
	// Create Wails application
	err := wails.Run(&options.App{
		Title:            w.coreCtx.Config.AppName + " Setup",
		Width:            900,
		Height:           650,
		MinWidth:         800,
		MinHeight:        600,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        w.app.startup,
		OnShutdown:       w.app.shutdown,
		Bind: []any{
			w.app,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to run Wails app: %w", err)
	}

	return nil
}

// Shutdown cleans up resources
func (w *WailsUI) Shutdown() error {
	close(w.shutdown)
	return nil
}

// UI Interface implementations

func (w *WailsUI) ShowWelcome() error {
	w.app.setStep("welcome")
	return nil
}

func (w *WailsUI) ShowLicense(license string) (bool, error) {
	w.app.setStep("license")
	// Wait for user response through frontend
	// This will be handled via frontend events
	return w.app.licenseAccepted, nil
}

func (w *WailsUI) SelectComponents(components []core.Component) ([]core.Component, error) {
	w.app.setStep("components")
	// Components selection is handled through frontend
	// Return the selected components based on app state
	return w.app.getSelectedComponents(), nil
}

func (w *WailsUI) SelectInstallPath(defaultPath string) (string, error) {
	w.app.setStep("path")
	if w.app.installPath == "" {
		w.app.installPath = defaultPath
	}
	return w.app.installPath, nil
}

func (w *WailsUI) ShowProgress(progress *core.Progress) error {
	// Send progress to frontend via event
	runtime.EventsEmit(w.app.ctx, "installation:progress", map[string]interface{}{
		"current":   progress.CurrentComponent,
		"total":     progress.TotalComponents,
		"component": progress.ComponentName,
		"progress":  progress.OverallProgress,
		"message":   progress.Message,
		"isError":   progress.IsError,
	})
	return nil
}

func (w *WailsUI) ShowError(err error, canRetry bool) (bool, error) {
	runtime.EventsEmit(w.app.ctx, "installation:error", map[string]interface{}{
		"error":    err.Error(),
		"canRetry": canRetry,
	})
	// TODO: Wait for user response
	return false, nil
}

func (w *WailsUI) ShowSuccess(summary *core.InstallSummary) error {
	w.app.setStep("complete")
	runtime.EventsEmit(w.app.ctx, "installation:complete", map[string]interface{}{
		"success":     summary.Success,
		"duration":    summary.Duration.String(),
		"installPath": summary.InstallPath,
		"components":  summary.ComponentsInstalled,
		"nextSteps":   summary.NextSteps,
	})
	return nil
}

func (w *WailsUI) RequestElevation(reason string) (bool, error) {
	runtime.EventsEmit(w.app.ctx, "elevation:required", map[string]interface{}{
		"reason": reason,
	})
	// In GUI mode, we should request elevation at startup if needed
	return true, nil
}

// App methods (bound to Wails)

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.EventsEmit(ctx, "app:ready", true)
}

func (a *App) shutdown(ctx context.Context) {
	// Cleanup
}

func (a *App) setStep(step string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.currentStep = step
	runtime.EventsEmit(a.ctx, "step:changed", step)
}

// GetConfig returns the installation configuration
func (a *App) GetConfig() InstallConfig {
	components := make([]ComponentInfo, 0, len(a.config.Components))
	for _, c := range a.config.Components {
		components = append(components, ComponentInfo{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Required:    c.Required,
			Selected:    c.Selected || c.Required,
			Size:        c.Size,
		})
	}

	return InstallConfig{
		AppName:     a.config.AppName,
		Version:     a.config.Version,
		Publisher:   a.config.Publisher,
		Website:     a.config.Website,
		License:     a.config.License,
		Components:  components,
		InstallPath: a.installPath,
	}
}

// GetCurrentStep returns the current installation step
func (a *App) GetCurrentStep() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentStep
}

// AcceptLicense accepts the license agreement
func (a *App) AcceptLicense() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.licenseAccepted = true
	runtime.EventsEmit(a.ctx, "license:accepted", true)
}

// RejectLicense rejects the license agreement
func (a *App) RejectLicense() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.licenseAccepted = false
	runtime.EventsEmit(a.ctx, "license:rejected", true)
}

// SetSelectedComponents sets the selected components
func (a *App) SetSelectedComponents(componentIDs []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.selectedComps = componentIDs

	// Update the core components
	for i := range a.config.Components {
		a.config.Components[i].Selected = false
		for _, id := range componentIDs {
			if a.config.Components[i].ID == id {
				a.config.Components[i].Selected = true
				break
			}
		}
	}
}

// SetInstallPath sets the installation path
func (a *App) SetInstallPath(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if path == "" {
		return fmt.Errorf("installation path cannot be empty")
	}

	a.installPath = path
	a.installer.SetInstallPath(path)
	return nil
}

// BrowseForFolder opens a folder browser dialog
func (a *App) BrowseForFolder() string {
	options := runtime.OpenDialogOptions{
		Title:            "Select Installation Directory",
		DefaultDirectory: a.installPath,
	}

	path, err := runtime.OpenDirectoryDialog(a.ctx, options)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to open directory dialog: "+err.Error())
		return a.installPath
	}

	if path != "" {
		a.SetInstallPath(path)
		return path
	}

	return a.installPath
}

// StartInstallation begins the installation process
func (a *App) StartInstallation() error {
	a.setStep("installing")

	// Set selected components
	a.installer.SetSelectedComponents(a.getSelectedComponents())

	// Run installation in background
	go func() {
		err := a.installer.ExecuteInstallation()
		if err != nil {
			runtime.EventsEmit(a.ctx, "installation:failed", err.Error())
			a.errorChan <- err
		} else {
			summary := a.installer.CreateSummary()
			runtime.EventsEmit(a.ctx, "installation:complete", map[string]interface{}{
				"success":     summary.Success,
				"duration":    summary.Duration.String(),
				"installPath": summary.InstallPath,
				"components":  summary.ComponentsInstalled,
				"nextSteps":   summary.NextSteps,
			})
		}
	}()

	return nil
}

// CancelInstallation cancels the installation
func (a *App) CancelInstallation() {
	runtime.EventsEmit(a.ctx, "installation:cancelled", true)
	runtime.Quit(a.ctx)
}

// OpenInstallDirectory opens the installation directory
func (a *App) OpenInstallDirectory() error {
	runtime.BrowserOpenURL(a.ctx, "file:///"+a.installPath)
	return nil
}

// Helper methods

func (a *App) getSelectedComponents() []core.Component {
	var selected []core.Component
	for _, c := range a.config.Components {
		if c.Selected || c.Required {
			selected = append(selected, c)
		}
	}
	return selected
}

// CalculateTotalSize calculates the total size of selected components
func (a *App) CalculateTotalSize() int64 {
	var total int64
	for _, c := range a.config.Components {
		if c.Selected || c.Required {
			total += c.Size
		}
	}
	return total
}

// FormatSize formats bytes to human readable format
func (a *App) FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FinishInstallation completes the installation and optionally launches the app
func (a *App) FinishInstallation(launchApp, viewReadme bool) error {
	// Handle post-installation actions
	if launchApp {
		// Launch the installed application if requested
		// This could be implemented based on the specific application
		runtime.LogInfo(a.ctx, "Application launch requested")
	}

	if viewReadme {
		// Open README or documentation if requested
		runtime.LogInfo(a.ctx, "README view requested")
	}

	// Exit the installer
	runtime.Quit(a.ctx)
	return nil
}

// ExitInstaller exits the installer application
func (a *App) ExitInstaller() error {
	runtime.Quit(a.ctx)
	return nil
}

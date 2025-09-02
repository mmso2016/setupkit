// Package core - Standard State Handlers for Wizard Provider
package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// BaseStateHandler provides common functionality for all state handlers
type BaseStateHandler struct {
	config  *Config
	context *Context
	title   string
	desc    string
}

// GetTitle returns the state title
func (bsh *BaseStateHandler) GetTitle() string {
	return bsh.title
}

// GetDescription returns the state description
func (bsh *BaseStateHandler) GetDescription() string {
	return bsh.desc
}

// OnEnter is called when entering the state (default implementation)
func (bsh *BaseStateHandler) OnEnter(ctx context.Context, data map[string]interface{}) error {
	return nil
}

// OnExit is called when leaving the state (default implementation)
func (bsh *BaseStateHandler) OnExit(ctx context.Context, data map[string]interface{}) error {
	return nil
}

// GetActions returns default actions (can be overridden)
func (bsh *BaseStateHandler) GetActions() []StateAction {
	return []StateAction{
		{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
		{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
		{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
	}
}

// WelcomeStateHandler handles the welcome state
type WelcomeStateHandler struct {
	BaseStateHandler
}

// NewWelcomeStateHandler creates a new welcome state handler
func NewWelcomeStateHandler(config *Config, context *Context) *WelcomeStateHandler {
	return &WelcomeStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   fmt.Sprintf("Welcome to %s Setup", config.AppName),
			desc:    fmt.Sprintf("This will install %s %s on your computer.", config.AppName, config.Version),
		},
	}
}

// Execute performs the welcome state logic
func (wsh *WelcomeStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	// Log installation start
	wsh.context.Logger.Info("Starting installation",
		"app", wsh.config.AppName,
		"version", wsh.config.Version,
		"mode", wsh.config.Mode)
	
	// Store basic app info in data
	data["app_name"] = wsh.config.AppName
	data["app_version"] = wsh.config.Version
	data["publisher"] = wsh.config.Publisher
	
	return nil
}

// Validate validates the welcome state
func (wsh *WelcomeStateHandler) Validate(data map[string]interface{}) error {
	// Welcome state always valid
	return nil
}

// GetActions returns welcome-specific actions
func (wsh *WelcomeStateHandler) GetActions() []StateAction {
	return []StateAction{
		{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
		{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
	}
}

// ModeSelectStateHandler handles the installation mode selection
type ModeSelectStateHandler struct {
	BaseStateHandler
}

// NewModeSelectStateHandler creates a new mode selection state handler
func NewModeSelectStateHandler(config *Config, context *Context) *ModeSelectStateHandler {
	return &ModeSelectStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   "Installation Mode",
			desc:    "Choose how you want to install the application.",
		},
	}
}

// Execute performs the mode selection logic
func (msh *ModeSelectStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	// Set default mode if not already set
	if _, exists := data["mode"]; !exists {
		data["mode"] = "custom" // Default to custom mode
	}
	
	return nil
}

// Validate validates the mode selection
func (msh *ModeSelectStateHandler) Validate(data map[string]interface{}) error {
	mode, ok := data["mode"]
	if !ok || mode == "" {
		return fmt.Errorf("installation mode must be selected")
	}
	
	validModes := map[string]bool{
		"express": true,
		"custom":  true,
	}
	
	if !validModes[mode.(string)] {
		return fmt.Errorf("invalid installation mode: %v", mode)
	}
	
	return nil
}

// LicenseStateHandler handles the license agreement
type LicenseStateHandler struct {
	BaseStateHandler
}

// NewLicenseStateHandler creates a new license state handler
func NewLicenseStateHandler(config *Config, context *Context) *LicenseStateHandler {
	return &LicenseStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   "License Agreement",
			desc:    "Please read and accept the license agreement to continue.",
		},
	}
}

// Execute performs the license state logic
func (lsh *LicenseStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	// Set license text in data
	data["license_text"] = lsh.config.License
	
	// Auto-accept in silent mode
	if lsh.config.AcceptLicense {
		data["accept_license"] = true
		lsh.context.Logger.Info("License accepted (automatic)")
	}
	
	return nil
}

// Validate validates the license acceptance
func (lsh *LicenseStateHandler) Validate(data map[string]interface{}) error {
	if accepted, ok := data["accept_license"]; !ok || accepted != true {
		return fmt.Errorf("license must be accepted to continue")
	}
	return nil
}

// ComponentsStateHandler handles component selection
type ComponentsStateHandler struct {
	BaseStateHandler
}

// NewComponentsStateHandler creates a new components state handler
func NewComponentsStateHandler(config *Config, context *Context) *ComponentsStateHandler {
	return &ComponentsStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   "Select Components",
			desc:    "Choose which components to install.",
		},
	}
}

// Execute performs the components state logic
func (csh *ComponentsStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	// Set available components
	data["available_components"] = csh.config.Components
	
	// Set default selected components
	var selectedIDs []string
	for _, comp := range csh.config.Components {
		if comp.Selected || comp.Required {
			selectedIDs = append(selectedIDs, comp.ID)
		}
	}
	
	if _, exists := data["selected_components"]; !exists {
		data["selected_components"] = selectedIDs
	}
	
	return nil
}

// Validate validates the component selection
func (csh *ComponentsStateHandler) Validate(data map[string]interface{}) error {
	selectedComponents, ok := data["selected_components"]
	if !ok {
		return fmt.Errorf("no components selected")
	}
	
	selectedIDs, ok := selectedComponents.([]string)
	if !ok {
		return fmt.Errorf("invalid component selection format")
	}
	
	// Check that at least one component is selected
	if len(selectedIDs) == 0 {
		return fmt.Errorf("at least one component must be selected")
	}
	
	// Ensure all required components are selected
	componentMap := make(map[string]Component)
	for _, comp := range csh.config.Components {
		componentMap[comp.ID] = comp
	}
	
	selectedMap := make(map[string]bool)
	for _, id := range selectedIDs {
		selectedMap[id] = true
	}
	
	for _, comp := range csh.config.Components {
		if comp.Required && !selectedMap[comp.ID] {
			return fmt.Errorf("required component '%s' must be selected", comp.Name)
		}
	}
	
	return nil
}

// LocationStateHandler handles installation location selection
type LocationStateHandler struct {
	BaseStateHandler
}

// NewLocationStateHandler creates a new location state handler
func NewLocationStateHandler(config *Config, context *Context) *LocationStateHandler {
	return &LocationStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   "Installation Location",
			desc:    "Choose where to install the application.",
		},
	}
}

// Execute performs the location state logic
func (lsh *LocationStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	// Set default installation path if not already set
	if _, exists := data["install_path"]; !exists {
		data["install_path"] = lsh.config.InstallDir
	}
	
	return nil
}

// Validate validates the installation location
func (lsh *LocationStateHandler) Validate(data map[string]interface{}) error {
	path, ok := data["install_path"]
	if !ok || path == "" {
		return fmt.Errorf("installation path is required")
	}
	
	installPath := path.(string)
	
	// Check if path is absolute
	if !filepath.IsAbs(installPath) {
		return fmt.Errorf("installation path must be absolute")
	}
	
	// Check if parent directory exists or can be created
	parentDir := filepath.Dir(installPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		// Try to create parent directory
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("cannot create installation directory: %w", err)
		}
	}
	
	// Check disk space if required
	if lsh.config.RequiredSpace > 0 {
		if err := CheckDiskSpace(installPath, lsh.config.RequiredSpace); err != nil {
			return fmt.Errorf("insufficient disk space: %w", err)
		}
	}
	
	return nil
}

// ReadyStateHandler handles the ready-to-install state
type ReadyStateHandler struct {
	BaseStateHandler
}

// NewReadyStateHandler creates a new ready state handler
func NewReadyStateHandler(config *Config, context *Context) *ReadyStateHandler {
	return &ReadyStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   "Ready to Install",
			desc:    "Review your installation choices and click Install to begin.",
		},
	}
}

// Execute performs the ready state logic
func (rsh *ReadyStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	// Generate installation summary
	summary := map[string]interface{}{
		"app_name":            data["app_name"],
		"app_version":         data["app_version"],
		"install_path":        data["install_path"],
		"selected_components": data["selected_components"],
		"license_accepted":    data["accept_license"],
	}
	
	data["install_summary"] = summary
	
	rsh.context.Logger.Info("Ready to install",
		"path", data["install_path"],
		"components", data["selected_components"])
	
	return nil
}

// Validate validates the ready state
func (rsh *ReadyStateHandler) Validate(data map[string]interface{}) error {
	// All validation should have been done in previous states
	return nil
}

// GetActions returns ready-specific actions
func (rsh *ReadyStateHandler) GetActions() []StateAction {
	return []StateAction{
		{ID: "install", Label: "Install", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
		{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
		{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
	}
}

// InstallingStateHandler handles the installation process
type InstallingStateHandler struct {
	BaseStateHandler
}

// NewInstallingStateHandler creates a new installing state handler
func NewInstallingStateHandler(config *Config, context *Context) *InstallingStateHandler {
	return &InstallingStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   "Installing",
			desc:    "Please wait while the application is being installed.",
		},
	}
}

// Execute performs the installation
func (ish *InstallingStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	ish.context.Logger.Info("Starting installation process")
	
	// Set installation progress
	data["progress"] = 0
	data["current_step"] = "Preparing installation"
	
	// This would typically call the actual installer
	// For now, we'll simulate the process
	if !ish.config.DryRun {
		// Actual installation logic would go here
		// This is where the installer would:
		// 1. Create directories
		// 2. Copy files
		// 3. Create shortcuts
		// 4. Update registry/PATH
		// 5. etc.
		
		// For simulation, just mark as in progress
		data["installation_started"] = true
	} else {
		// Dry run mode
		ish.context.Logger.Info("Dry run mode - skipping actual installation")
		data["installation_started"] = true
		data["progress"] = 100
		data["current_step"] = "Installation complete (dry run)"
	}
	
	return nil
}

// Validate validates the installation state
func (ish *InstallingStateHandler) Validate(data map[string]interface{}) error {
	// Installation state doesn't need validation
	return nil
}

// GetActions returns installing-specific actions (none during install)
func (ish *InstallingStateHandler) GetActions() []StateAction {
	return []StateAction{} // No actions during installation
}

// CompleteStateHandler handles the installation completion
type CompleteStateHandler struct {
	BaseStateHandler
}

// NewCompleteStateHandler creates a new complete state handler
func NewCompleteStateHandler(config *Config, context *Context) *CompleteStateHandler {
	return &CompleteStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   "Installation Complete",
			desc:    fmt.Sprintf("%s has been successfully installed.", config.AppName),
		},
	}
}

// Execute performs the completion logic
func (csh *CompleteStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	csh.context.Logger.Info("Installation completed successfully",
		"app", csh.config.AppName,
		"path", data["install_path"])
	
	// Set completion data
	data["installation_complete"] = true
	data["completion_time"] = csh.context.StartTime.Format("2006-01-02 15:04:05")
	
	// Generate next steps
	nextSteps := []string{
		fmt.Sprintf("Run '%s' to start the application", csh.config.AppName),
	}
	
	if installPath, ok := data["install_path"]; ok {
		nextSteps = append(nextSteps, fmt.Sprintf("Application installed to: %s", installPath))
	}
	
	data["next_steps"] = nextSteps
	
	return nil
}

// Validate validates the completion state
func (csh *CompleteStateHandler) Validate(data map[string]interface{}) error {
	// Completion state doesn't need validation
	return nil
}

// GetActions returns completion-specific actions
func (csh *CompleteStateHandler) GetActions() []StateAction {
	return []StateAction{
		{ID: "finish", Label: "Finish", Type: ActionTypeFinish, Primary: true, Enabled: true, Visible: true},
	}
}
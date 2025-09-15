// Package installer provides DFA-based installer functionality
// This bridges the hierarchical wizard system with installer-specific logic
package installer

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// DFAInstaller wraps the hierarchical DFA for installer-specific workflows
type DFAInstaller struct {
	wizard   *wizard.HierarchicalDFA
	config   *core.Config
	ui       core.UI
	context  *core.Context
	
	// Installer-specific state
	selectedComponents []core.Component
	installPath        string
	totalSize          int64
}

// NewDFAInstaller creates a new DFA-based installer
func NewDFAInstaller(config *core.Config) *DFAInstaller {
	installer := &DFAInstaller{
		wizard:   wizard.NewHierarchical(),
		config:   config,
		selectedComponents: make([]core.Component, 0),
	}
	
	// Set default install path if not specified
	if config.InstallDir == "" {
		if runtime.GOOS == "windows" {
			config.InstallDir = filepath.Join("C:", "Program Files", config.AppName)
		} else {
			config.InstallDir = filepath.Join("/usr/local", config.AppName)
		}
	}
	installer.installPath = config.InstallDir
	
	// Initialize installer-specific wizard states
	installer.setupInstallerStates()
	
	return installer
}

// setupInstallerStates configures the standard installer workflow using hierarchical DFA
func (d *DFAInstaller) setupInstallerStates() {
	// 1. Welcome State
	d.wizard.AddMainState("welcome", &wizard.MainStateConfig{
		StateConfig: &wizard.StateConfig{
			Name:        "Welcome",
			Description: fmt.Sprintf("Welcome to %s Setup", d.config.AppName),
			CanGoNext:   true,
			CanGoBack:   false,
			CanCancel:   true,
			OnEnter:     d.onEnterWelcome,
		},
		SubStates: make(map[wizard.SubState]*wizard.SubStateConfig),
	})

	// 2. License State with sub-states
	d.wizard.AddMainState("license", &wizard.MainStateConfig{
		StateConfig: &wizard.StateConfig{
			Name:        "License Agreement",
			Description: "Please review and accept the license agreement",
			CanGoNext:   true,
			CanGoBack:   true,
			CanCancel:   true,
			OnEnter:     d.onEnterLicense,
			ValidateFunc: d.validateLicense,
		},
		SubStates:               make(map[wizard.SubState]*wizard.SubStateConfig),
		InitialSubState:         "reading",
		RequireSubStateCompletion: true,
	})

	// License sub-states
	d.wizard.AddSubState("license", "reading", &wizard.SubStateConfig{
		Name:        "Reading License",
		Description: "User is reviewing the license text",
		AllowedActions: map[wizard.SubAction]bool{
			wizard.SubActionScroll: true,
		},
		OnEnter: d.onEnterLicenseReading,
	})

	d.wizard.AddSubState("license", "accepting", &wizard.SubStateConfig{
		Name:        "Accepting License",
		Description: "User can accept or decline the license",
		AllowedActions: map[wizard.SubAction]bool{
			wizard.SubActionSelect: true,
		},
		OnEnter: d.onEnterLicenseAccepting,
		CanComplete: func(data map[string]interface{}) bool {
			accepted, ok := data["license_accepted"]
			return ok && accepted.(bool)
		},
	})

	// 3. Components State with sub-states
	d.wizard.AddMainState("components", &wizard.MainStateConfig{
		StateConfig: &wizard.StateConfig{
			Name:        "Component Selection",
			Description: "Choose which components to install",
			CanGoNext:   true,
			CanGoBack:   true,
			CanCancel:   true,
			OnEnter:     d.onEnterComponents,
			ValidateFunc: d.validateComponents,
		},
		SubStates:               make(map[wizard.SubState]*wizard.SubStateConfig),
		InitialSubState:         "browsing",
		RequireSubStateCompletion: false,
	})

	// Components sub-states
	d.wizard.AddSubState("components", "browsing", &wizard.SubStateConfig{
		Name:        "Browsing Components",
		Description: "User is viewing available components",
		AllowedActions: map[wizard.SubAction]bool{
			wizard.SubActionSelect:   true,
			wizard.SubActionDeselect: true,
			wizard.SubActionRefresh:  true,
		},
		OnEnter: d.onEnterComponentsBrowsing,
		AutoTransitionFunc: func(data map[string]interface{}) wizard.SubState {
			// Auto-transition to calculating after component selection changes
			if data["components_changed"] == true {
				return "calculating"
			}
			return ""
		},
	})

	d.wizard.AddSubState("components", "calculating", &wizard.SubStateConfig{
		Name:        "Calculating Size",
		Description: "Calculating total installation size",
		AllowedActions: map[wizard.SubAction]bool{},
		OnEnter: d.onEnterComponentsCalculating,
		AutoTransitionTo: "browsing", // Return to browsing after calculation
	})

	// 4. Location State with sub-states
	d.wizard.AddMainState("location", &wizard.MainStateConfig{
		StateConfig: &wizard.StateConfig{
			Name:        "Installation Location",
			Description: "Choose where to install the application",
			CanGoNext:   true,
			CanGoBack:   true,
			CanCancel:   true,
			OnEnter:     d.onEnterLocation,
			ValidateFunc: d.validateLocation,
		},
		SubStates:               make(map[wizard.SubState]*wizard.SubStateConfig),
		InitialSubState:         "browsing",
		RequireSubStateCompletion: true,
	})

	// Location sub-states
	d.wizard.AddSubState("location", "browsing", &wizard.SubStateConfig{
		Name:        "Browsing Directories",
		Description: "User is browsing for installation directory",
		AllowedActions: map[wizard.SubAction]bool{
			wizard.SubActionInput:   true,
			wizard.SubActionSelect:  true,
		},
		OnEnter: d.onEnterLocationBrowsing,
	})

	d.wizard.AddSubState("location", "validating", &wizard.SubStateConfig{
		Name:        "Validating Path",
		Description: "Checking if installation path is valid",
		AllowedActions: map[wizard.SubAction]bool{},
		OnEnter: d.onEnterLocationValidating,
		CanComplete: func(data map[string]interface{}) bool {
			return data["path_valid"] == true
		},
	})

	// 5. Summary State
	d.wizard.AddMainState("summary", &wizard.MainStateConfig{
		StateConfig: &wizard.StateConfig{
			Name:        "Installation Summary",
			Description: "Review installation settings",
			CanGoNext:   true,
			CanGoBack:   true,
			CanCancel:   true,
			OnEnter:     d.onEnterSummary,
		},
		SubStates: make(map[wizard.SubState]*wizard.SubStateConfig),
	})

	// 6. Installing State
	d.wizard.AddMainState("installing", &wizard.MainStateConfig{
		StateConfig: &wizard.StateConfig{
			Name:        "Installing",
			Description: "Installation in progress",
			CanGoNext:   false,
			CanGoBack:   false,
			CanCancel:   false,
			OnEnter:     d.onEnterInstalling,
		},
		SubStates: make(map[wizard.SubState]*wizard.SubStateConfig),
	})

	// 7. Complete State
	d.wizard.AddMainState("complete", &wizard.MainStateConfig{
		StateConfig: &wizard.StateConfig{
			Name:        "Installation Complete",
			Description: "Installation finished successfully",
			CanGoNext:   false,
			CanGoBack:   false,
			CanCancel:   false,
			OnEnter:     d.onEnterComplete,
		},
		SubStates: make(map[wizard.SubState]*wizard.SubStateConfig),
	})

	// Set callbacks for installer-specific functionality
	d.wizard.SetCallbacks(&wizard.Callbacks{
		OnTransition: d.onTransition,
		OnDataChange: d.onDataChange,
	})
}

// State callback implementations
func (d *DFAInstaller) onEnterWelcome(data map[string]interface{}) error {
	data["app_name"] = d.config.AppName
	data["version"] = d.config.Version
	data["publisher"] = d.config.Publisher
	return nil
}

func (d *DFAInstaller) onEnterLicense(data map[string]interface{}) error {
	data["license_text"] = d.config.License
	data["license_accepted"] = false
	return nil
}

func (d *DFAInstaller) onEnterLicenseReading(data map[string]interface{}) error {
	// User started reading license
	return nil
}

func (d *DFAInstaller) onEnterLicenseAccepting(data map[string]interface{}) error {
	// User can now accept/decline
	return nil
}

func (d *DFAInstaller) validateLicense(data map[string]interface{}) error {
	accepted, ok := data["license_accepted"]
	if !ok || !accepted.(bool) {
		return fmt.Errorf("you must accept the license agreement to continue")
	}
	return nil
}

func (d *DFAInstaller) onEnterComponents(data map[string]interface{}) error {
	data["available_components"] = d.config.Components
	data["selected_components"] = d.selectedComponents
	return nil
}

func (d *DFAInstaller) onEnterComponentsBrowsing(data map[string]interface{}) error {
	// Initialize with required components selected
	if len(d.selectedComponents) == 0 {
		for _, comp := range d.config.Components {
			if comp.Required || comp.Selected {
				d.selectedComponents = append(d.selectedComponents, comp)
			}
		}
		data["selected_components"] = d.selectedComponents
	}
	return nil
}

func (d *DFAInstaller) onEnterComponentsCalculating(data map[string]interface{}) error {
	// Calculate total size
	d.totalSize = 0
	for _, comp := range d.selectedComponents {
		d.totalSize += comp.Size
	}
	data["total_size"] = d.totalSize
	data["components_changed"] = false // Reset flag
	return nil
}

func (d *DFAInstaller) validateComponents(data map[string]interface{}) error {
	if len(d.selectedComponents) == 0 {
		return fmt.Errorf("at least one component must be selected")
	}
	
	// Validate that all required components are selected
	for _, comp := range d.config.Components {
		if comp.Required {
			found := false
			for _, selected := range d.selectedComponents {
				if selected.ID == comp.ID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("required component '%s' must be selected", comp.Name)
			}
		}
	}
	return nil
}

func (d *DFAInstaller) onEnterLocation(data map[string]interface{}) error {
	data["install_path"] = d.installPath
	data["path_valid"] = false
	return nil
}

func (d *DFAInstaller) onEnterLocationBrowsing(data map[string]interface{}) error {
	// User can browse for directory
	return nil
}

func (d *DFAInstaller) onEnterLocationValidating(data map[string]interface{}) error {
	// Validate the selected path
	if path, ok := data["install_path"].(string); ok && path != "" {
		d.installPath = path
		// TODO: Add actual path validation (writable, space, etc.)
		data["path_valid"] = true
	}
	return nil
}

func (d *DFAInstaller) validateLocation(data map[string]interface{}) error {
	if d.installPath == "" {
		return fmt.Errorf("installation path must be specified")
	}
	// TODO: Add more validation (path exists, writable, sufficient space)
	return nil
}

func (d *DFAInstaller) onEnterSummary(data map[string]interface{}) error {
	data["summary"] = map[string]interface{}{
		"app_name":    d.config.AppName,
		"version":     d.config.Version,
		"install_path": d.installPath,
		"components":  d.selectedComponents,
		"total_size":  d.totalSize,
	}
	return nil
}

func (d *DFAInstaller) onEnterInstalling(data map[string]interface{}) error {
	// Start actual installation
	go d.performInstallation()
	return nil
}

func (d *DFAInstaller) onEnterComplete(data map[string]interface{}) error {
	data["success"] = true
	data["install_path"] = d.installPath
	return nil
}

func (d *DFAInstaller) onTransition(from, to wizard.State, action wizard.Action) error {
	if d.context != nil && d.context.Logger != nil {
		d.context.Logger.Info("Installer state transition",
			"from", from,
			"to", to,
			"action", action)
	}
	return nil
}

func (d *DFAInstaller) onDataChange(state wizard.State, key string, oldValue, newValue interface{}) error {
	// Handle installer-specific data changes
	switch key {
	case "license_accepted":
		if newValue.(bool) {
			// License accepted, can move to accepting sub-state
			d.wizard.NavigateToSubState("accepting")
		}
	case "selected_components":
		if comps, ok := newValue.([]core.Component); ok {
			d.selectedComponents = comps
			d.wizard.SetData("components_changed", true)
		}
	case "install_path":
		if path, ok := newValue.(string); ok {
			d.installPath = path
			// Trigger path validation
			d.wizard.NavigateToSubState("validating")
		}
	}
	return nil
}

func (d *DFAInstaller) performInstallation() {
	// TODO: Implement actual installation logic
	// This would copy files, create registry entries, etc.
	
	// Simulate installation progress
	for i := 0; i <= 100; i += 10 {
		d.wizard.SetData("install_progress", i)
		// time.Sleep(100 * time.Millisecond) // Simulate work
	}
	
	// Move to complete state
	d.wizard.NavigateToMainState("complete")
}

// Public API methods

// SetUI sets the user interface for the installer
func (d *DFAInstaller) SetUI(ui core.UI) {
	d.ui = ui
}

// SetContext sets the installer context
func (d *DFAInstaller) SetContext(ctx *core.Context) {
	d.context = ctx
}

// Run starts the installer wizard
func (d *DFAInstaller) Run() error {
	if d.ui != nil {
		if err := d.ui.Initialize(d.context); err != nil {
			return fmt.Errorf("failed to initialize UI: %w", err)
		}
		defer d.ui.Shutdown()
	}
	
	// Start the wizard from welcome state
	return d.wizard.NavigateToMainState("welcome")
}

// GetCurrentState returns the current wizard state
func (d *DFAInstaller) GetCurrentState() wizard.CompositeState {
	return d.wizard.GetCurrentState()
}

// HandleAction processes user actions
func (d *DFAInstaller) HandleAction(action wizard.Action) error {
	switch action {
	case wizard.ActionNext:
		return d.handleNext()
	case wizard.ActionBack:
		return d.wizard.GoBack()
	case wizard.ActionCancel:
		return d.handleCancel()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// HandleSubAction processes sub-actions within current state
func (d *DFAInstaller) HandleSubAction(action wizard.SubAction) error {
	return d.wizard.HandleSubAction(action)
}

func (d *DFAInstaller) handleNext() error {
	current := d.wizard.GetCurrentState()
	
	// Determine next main state based on current state
	var nextState wizard.MainState
	switch current.Main {
	case "welcome":
		nextState = "license"
	case "license":
		nextState = "components"
	case "components":
		nextState = "location"
	case "location":
		nextState = "summary"
	case "summary":
		nextState = "installing"
	case "installing":
		nextState = "complete"
	default:
		return fmt.Errorf("no next state defined for %s", current.Main)
	}
	
	return d.wizard.NavigateToMainState(nextState)
}

func (d *DFAInstaller) handleCancel() error {
	// TODO: Implement cleanup logic
	return fmt.Errorf("installation cancelled by user")
}

// GetInstallationSummary returns summary of what will be installed
func (d *DFAInstaller) GetInstallationSummary() map[string]interface{} {
	return map[string]interface{}{
		"app_name":         d.config.AppName,
		"version":          d.config.Version,
		"install_path":     d.installPath,
		"selected_components": d.selectedComponents,
		"total_size":       d.totalSize,
	}
}
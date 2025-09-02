// Package core - Standard Wizard Provider implementation
package core

import (
	"fmt"

	"github.com/mmso2016/setupkit/pkg/wizard"
)

// StandardWizardProvider provides the built-in DFA configurations
type StandardWizardProvider struct {
	mode       InstallMode
	config     *Config
	context    *Context
	dfa        *wizard.DFA
	handlers   map[wizard.State]WizardStateHandler
	uiMappings map[wizard.State]UIStateConfig
}

// NewStandardWizardProvider creates a new standard provider
func NewStandardWizardProvider(mode InstallMode) *StandardWizardProvider {
	return &StandardWizardProvider{
		mode:       mode,
		handlers:   make(map[wizard.State]WizardStateHandler),
		uiMappings: make(map[wizard.State]UIStateConfig),
	}
}

// Initialize sets up the provider with installer context
func (sp *StandardWizardProvider) Initialize(config *Config, context *Context) error {
	sp.config = config
	sp.context = context
	
	// Build DFA based on mode
	if err := sp.buildDFA(); err != nil {
		return fmt.Errorf("failed to build DFA: %w", err)
	}
	
	// Register handlers
	sp.registerHandlers()
	
	// Setup UI mappings
	sp.setupUIMappings()
	
	return nil
}

// GetDFA returns the configured DFA
func (sp *StandardWizardProvider) GetDFA() (*wizard.DFA, error) {
	if sp.dfa == nil {
		return nil, fmt.Errorf("DFA not initialized - call Initialize() first")
	}
	return sp.dfa, nil
}

// GetStateHandler returns the handler for a specific state
func (sp *StandardWizardProvider) GetStateHandler(state wizard.State) WizardStateHandler {
	return sp.handlers[state]
}

// GetUIMapping returns UI configuration for a state
func (sp *StandardWizardProvider) GetUIMapping(state wizard.State) UIStateConfig {
	return sp.uiMappings[state]
}

// ValidateConfiguration checks if the DFA is properly configured
func (sp *StandardWizardProvider) ValidateConfiguration() error {
	if sp.dfa == nil {
		return fmt.Errorf("DFA not initialized")
	}
	return sp.dfa.Validate()
}

// GetMode returns the installation mode
func (sp *StandardWizardProvider) GetMode() InstallMode {
	return sp.mode
}

// buildDFA creates the DFA based on the installation mode
func (sp *StandardWizardProvider) buildDFA() error {
	sp.dfa = wizard.New()
	
	// Set dry-run mode if configured
	if sp.config.DryRun {
		sp.dfa.SetDryRun(true)
	}
	
	switch sp.mode {
	case ModeExpress:
		return sp.buildExpressDFA()
	case ModeCustom:
		return sp.buildCustomDFA()
	case ModeAdvanced:
		return sp.buildAdvancedDFA()
	default:
		return fmt.Errorf("unsupported install mode: %v", sp.mode)
	}
}

// buildExpressDFA creates a simplified flow for express installation
func (sp *StandardWizardProvider) buildExpressDFA() error {
	// Express flow: Welcome -> License -> Installing -> Complete
	states := []struct {
		state  wizard.State
		config *wizard.StateConfig
	}{
		{
			StateWelcome,
			&wizard.StateConfig{
				Name:      "Welcome",
				CanGoNext: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateLicense,
				},
			},
		},
		{
			StateLicense,
			&wizard.StateConfig{
				Name:      "License Agreement",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateInstalling,
				},
				ValidateFunc: sp.validateLicense,
			},
		},
		{
			StateInstalling,
			&wizard.StateConfig{
				Name:      "Installing",
				CanCancel: false,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateComplete,
				},
			},
		},
		{
			StateComplete,
			&wizard.StateConfig{
				Name: "Installation Complete",
			},
		},
	}
	
	// Add states to DFA
	for _, s := range states {
		if err := sp.dfa.AddState(s.state, s.config); err != nil {
			return fmt.Errorf("failed to add state %s: %w", s.state, err)
		}
	}
	
	// Mark final state
	return sp.dfa.AddFinalState(StateComplete)
}

// buildCustomDFA creates a flow with component selection
func (sp *StandardWizardProvider) buildCustomDFA() error {
	// Custom flow: Welcome -> License -> Components -> Location -> Ready -> Installing -> Complete
	states := []struct {
		state  wizard.State
		config *wizard.StateConfig
	}{
		{
			StateWelcome,
			&wizard.StateConfig{
				Name:      "Welcome",
				CanGoNext: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateLicense,
				},
			},
		},
		{
			StateLicense,
			&wizard.StateConfig{
				Name:      "License Agreement",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateComponents,
				},
				ValidateFunc: sp.validateLicense,
			},
		},
		{
			StateComponents,
			&wizard.StateConfig{
				Name:      "Select Components",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateLocation,
				},
				ValidateFunc: sp.validateComponents,
			},
		},
		{
			StateLocation,
			&wizard.StateConfig{
				Name:      "Installation Location",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateReady,
				},
				ValidateFunc: sp.validateLocation,
			},
		},
		{
			StateReady,
			&wizard.StateConfig{
				Name:      "Ready to Install",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateInstalling,
				},
			},
		},
		{
			StateInstalling,
			&wizard.StateConfig{
				Name:      "Installing",
				CanCancel: false,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateComplete,
				},
			},
		},
		{
			StateComplete,
			&wizard.StateConfig{
				Name: "Installation Complete",
			},
		},
	}
	
	// Add states to DFA
	for _, s := range states {
		if err := sp.dfa.AddState(s.state, s.config); err != nil {
			return fmt.Errorf("failed to add state %s: %w", s.state, err)
		}
	}
	
	// Mark final state
	return sp.dfa.AddFinalState(StateComplete)
}

// buildAdvancedDFA creates a flow with mode selection
func (sp *StandardWizardProvider) buildAdvancedDFA() error {
	// Advanced flow: Welcome -> Mode Select -> License -> Components -> Location -> Ready -> Installing -> Complete
	states := []struct {
		state  wizard.State
		config *wizard.StateConfig
	}{
		{
			StateWelcome,
			&wizard.StateConfig{
				Name:      "Welcome",
				CanGoNext: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateModeSelect,
				},
			},
		},
		{
			StateModeSelect,
			&wizard.StateConfig{
				Name:      "Installation Mode",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				NextStateFunc: sp.getModeBasedNextState,
				ValidateFunc:  sp.validateModeSelection,
			},
		},
		{
			StateLicense,
			&wizard.StateConfig{
				Name:      "License Agreement",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateComponents,
				},
				ValidateFunc: sp.validateLicense,
			},
		},
		{
			StateComponents,
			&wizard.StateConfig{
				Name:      "Select Components",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateLocation,
				},
				ValidateFunc: sp.validateComponents,
			},
		},
		{
			StateLocation,
			&wizard.StateConfig{
				Name:      "Installation Location",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateReady,
				},
				ValidateFunc: sp.validateLocation,
			},
		},
		{
			StateReady,
			&wizard.StateConfig{
				Name:      "Ready to Install",
				CanGoNext: true,
				CanGoBack: true,
				CanCancel: true,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateInstalling,
				},
			},
		},
		{
			StateInstalling,
			&wizard.StateConfig{
				Name:      "Installing",
				CanCancel: false,
				Transitions: map[wizard.Action]wizard.State{
					wizard.ActionNext: StateComplete,
				},
			},
		},
		{
			StateComplete,
			&wizard.StateConfig{
				Name: "Installation Complete",
			},
		},
	}
	
	// Add states to DFA
	for _, s := range states {
		if err := sp.dfa.AddState(s.state, s.config); err != nil {
			return fmt.Errorf("failed to add state %s: %w", s.state, err)
		}
	}
	
	// Mark final state
	return sp.dfa.AddFinalState(StateComplete)
}

// registerHandlers registers state handlers
func (sp *StandardWizardProvider) registerHandlers() {
	sp.handlers[StateWelcome] = NewWelcomeStateHandler(sp.config, sp.context)
	sp.handlers[StateModeSelect] = NewModeSelectStateHandler(sp.config, sp.context)
	sp.handlers[StateLicense] = NewLicenseStateHandler(sp.config, sp.context)
	sp.handlers[StateComponents] = NewComponentsStateHandler(sp.config, sp.context)
	sp.handlers[StateLocation] = NewLocationStateHandler(sp.config, sp.context)
	sp.handlers[StateReady] = NewReadyStateHandler(sp.config, sp.context)
	sp.handlers[StateInstalling] = NewInstallingStateHandler(sp.config, sp.context)
	sp.handlers[StateComplete] = NewCompleteStateHandler(sp.config, sp.context)
}

// setupUIMappings sets up UI configurations for each state
func (sp *StandardWizardProvider) setupUIMappings() {
	sp.uiMappings[StateWelcome] = UIStateConfig{
		Title:       fmt.Sprintf("Welcome to %s Setup", sp.config.AppName),
		Description: fmt.Sprintf("This will install %s %s on your computer.", sp.config.AppName, sp.config.Version),
		Type:        UIStateTypeWelcome,
		Layout:      LayoutTypeDefault,
		Actions: []StateAction{
			{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
			{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
		},
	}
	
	sp.uiMappings[StateModeSelect] = UIStateConfig{
		Title:       "Installation Mode",
		Description: "Choose how you want to install the application.",
		Type:        UIStateTypeSelection,
		Layout:      LayoutTypeDefault,
		Fields: []UIField{
			{
				ID:       "mode",
				Label:    "Installation Mode",
				Type:     FieldTypeRadio,
				Required: true,
				Options: []FieldOption{
					{ID: "express", Label: "Express (Recommended)", Value: "express"},
					{ID: "custom", Label: "Custom", Value: "custom"},
				},
			},
		},
		Actions: []StateAction{
			{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
			{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
			{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
		},
	}
	
	sp.uiMappings[StateLicense] = UIStateConfig{
		Title:       "License Agreement",
		Description: "Please read and accept the license agreement to continue.",
		Type:        UIStateTypeLicense,
		Layout:      LayoutTypeDefault,
		Fields: []UIField{
			{
				ID:       "license_text",
				Type:     FieldTypeTextArea,
				Value:    sp.config.License,
				Required: false,
			},
			{
				ID:       "accept_license",
				Label:    "I accept the terms of the license agreement",
				Type:     FieldTypeCheckbox,
				Required: true,
			},
		},
		Actions: []StateAction{
			{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
			{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
			{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
		},
	}
	
	// Add other UI mappings...
	sp.setupComponentsMapping()
	sp.setupLocationMapping()
	sp.setupReadyMapping()
	sp.setupInstallingMapping()
	sp.setupCompleteMapping()
}

// setupComponentsMapping sets up the components selection UI
func (sp *StandardWizardProvider) setupComponentsMapping() {
	options := make([]FieldOption, len(sp.config.Components))
	for i, comp := range sp.config.Components {
		options[i] = FieldOption{
			ID:       comp.ID,
			Label:    comp.Name,
			Value:    comp.Selected,
			Disabled: comp.Required,
		}
	}
	
	sp.uiMappings[StateComponents] = UIStateConfig{
		Title:       "Select Components",
		Description: "Choose which components to install.",
		Type:        UIStateTypeSelection,
		Layout:      LayoutTypeDefault,
		Fields: []UIField{
			{
				ID:      "components",
				Label:   "Components",
				Type:    FieldTypeList,
				Options: options,
			},
		},
		Actions: []StateAction{
			{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
			{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
			{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
		},
	}
}

// setupLocationMapping sets up the installation location UI
func (sp *StandardWizardProvider) setupLocationMapping() {
	sp.uiMappings[StateLocation] = UIStateConfig{
		Title:       "Installation Location",
		Description: "Choose where to install the application.",
		Type:        UIStateTypeInput,
		Layout:      LayoutTypeDefault,
		Fields: []UIField{
			{
				ID:          "install_path",
				Label:       "Installation Directory",
				Type:        FieldTypePath,
				Value:       sp.config.InstallDir,
				Placeholder: "Choose installation directory",
				Required:    true,
			},
		},
		Actions: []StateAction{
			{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
			{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
			{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
		},
	}
}

// setupReadyMapping sets up the ready to install UI
func (sp *StandardWizardProvider) setupReadyMapping() {
	sp.uiMappings[StateReady] = UIStateConfig{
		Title:       "Ready to Install",
		Description: "Review your installation choices and click Install to begin.",
		Type:        UIStateTypeSummary,
		Layout:      LayoutTypeDefault,
		Actions: []StateAction{
			{ID: "install", Label: "Install", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
			{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
			{ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true},
		},
	}
}

// setupInstallingMapping sets up the installation progress UI
func (sp *StandardWizardProvider) setupInstallingMapping() {
	sp.uiMappings[StateInstalling] = UIStateConfig{
		Title:       "Installing",
		Description: "Please wait while the application is being installed.",
		Type:        UIStateTypeProgress,
		Layout:      LayoutTypeDefault,
		Actions:     []StateAction{}, // No actions during installation
	}
}

// setupCompleteMapping sets up the completion UI
func (sp *StandardWizardProvider) setupCompleteMapping() {
	sp.uiMappings[StateComplete] = UIStateConfig{
		Title:       "Installation Complete",
		Description: fmt.Sprintf("%s has been successfully installed.", sp.config.AppName),
		Type:        UIStateTypeSummary,
		Layout:      LayoutTypeDefault,
		Actions: []StateAction{
			{ID: "finish", Label: "Finish", Type: ActionTypeFinish, Primary: true, Enabled: true, Visible: true},
		},
	}
}

// Validation functions
func (sp *StandardWizardProvider) validateLicense(data map[string]interface{}) error {
	if accepted, ok := data["accept_license"]; !ok || accepted != true {
		return fmt.Errorf("license must be accepted to continue")
	}
	return nil
}

func (sp *StandardWizardProvider) validateComponents(data map[string]interface{}) error {
	components, ok := data["components"]
	if !ok {
		return fmt.Errorf("no components selected")
	}
	
	// At least one component must be selected
	if selectedComponents, ok := components.([]string); ok && len(selectedComponents) == 0 {
		return fmt.Errorf("at least one component must be selected")
	}
	
	return nil
}

func (sp *StandardWizardProvider) validateLocation(data map[string]interface{}) error {
	path, ok := data["install_path"]
	if !ok || path == "" {
		return fmt.Errorf("installation path is required")
	}
	
	// Additional path validation could be added here
	
	return nil
}

func (sp *StandardWizardProvider) validateModeSelection(data map[string]interface{}) error {
	mode, ok := data["mode"]
	if !ok || mode == "" {
		return fmt.Errorf("installation mode must be selected")
	}
	return nil
}

// getModeBasedNextState determines the next state based on mode selection
func (sp *StandardWizardProvider) getModeBasedNextState(data map[string]interface{}) (wizard.State, error) {
	mode, ok := data["mode"]
	if !ok {
		return "", fmt.Errorf("mode not selected")
	}
	
	switch mode {
	case "express":
		return StateInstalling, nil // Skip license, components, location for express
	case "custom":
		return StateLicense, nil // Go through full flow
	default:
		return "", fmt.Errorf("unknown mode: %v", mode)
	}
}
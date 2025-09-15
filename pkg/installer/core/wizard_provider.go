// Package core - DFA-based Wizard Provider System
package core

import (
	"context"
	"fmt"

	"github.com/mmso2016/setupkit/pkg/wizard"
)

// WizardProvider defines the interface for wizard providers
// This bridges between the DFA system and the installer framework
type WizardProvider interface {
	// GetDFA returns the configured DFA for the installation
	GetDFA() (*wizard.DFA, error)
	
	// GetStateHandler returns the handler for a specific state
	GetStateHandler(state wizard.State) WizardStateHandler
	
	// GetUIMapping returns UI configuration for a state
	GetUIMapping(state wizard.State) UIStateConfig
	
	// ValidateConfiguration checks if the DFA is properly configured
	ValidateConfiguration() error
	
	// GetMode returns the installation mode
	GetMode() InstallMode
	
	// Initialize sets up the provider with installer context
	Initialize(config *Config, context *Context) error
}

// WizardStateHandler handles the logic for each wizard state
type WizardStateHandler interface {
	// OnEnter is called when entering the state
	OnEnter(ctx context.Context, data map[string]interface{}) error
	
	// OnExit is called when leaving the state
	OnExit(ctx context.Context, data map[string]interface{}) error
	
	// Execute performs the main action for this state
	Execute(ctx context.Context, data map[string]interface{}) error
	
	// Validate checks if the state can be completed
	Validate(data map[string]interface{}) error
	
	// GetActions returns available actions for this state
	GetActions() []StateAction
	
	// GetTitle returns the state title for UI display
	GetTitle() string
	
	// GetDescription returns the state description
	GetDescription() string
}

// StateAction represents an action available in a state
type StateAction struct {
	ID       string
	Label    string
	Type     ActionType
	Primary  bool
	Enabled  bool
	Visible  bool
	Shortcut string
}

// ActionType defines the type of action
type ActionType string

const (
	ActionTypeNext     ActionType = "next"
	ActionTypeBack     ActionType = "back"
	ActionTypeSkip     ActionType = "skip"
	ActionTypeCancel   ActionType = "cancel"
	ActionTypeCustom   ActionType = "custom"
	ActionTypeFinish   ActionType = "finish"
)

// UIStateConfig defines UI configuration for a state
type UIStateConfig struct {
	Title       string
	Description string
	Icon        string
	Type        UIStateType
	Fields      []UIField
	Actions     []StateAction
	Layout      LayoutType
	Validation  UIValidation
	Help        string
	Template    string // Template name for rendering
}

// UIStateType defines the type of UI state
type UIStateType string

const (
	UIStateTypeWelcome    UIStateType = "welcome"
	UIStateTypeLicense    UIStateType = "license"
	UIStateTypeInput      UIStateType = "input"
	UIStateTypeSelection  UIStateType = "selection"
	UIStateTypeProgress   UIStateType = "progress"
	UIStateTypeSummary    UIStateType = "summary"
	UIStateTypeError      UIStateType = "error"
	UIStateTypeCustom     UIStateType = "custom"
)

// UIField represents a field in the UI
type UIField struct {
	ID          string
	Label       string
	Type        FieldType
	Value       interface{}
	Placeholder string
	Required    bool
	Validation  string // Regex or validation rule
	Options     []FieldOption
	Help        string
}

// FieldType defines the type of UI field
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypePassword FieldType = "password"
	FieldTypePath     FieldType = "path"
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeRadio    FieldType = "radio"
	FieldTypeDropdown FieldType = "dropdown"
	FieldTypeList     FieldType = "list"
	FieldTypeTextArea FieldType = "textarea"
	FieldTypeTheme    FieldType = "theme"
)

// FieldOption represents an option for selection fields
type FieldOption struct {
	ID       string
	Label    string
	Value    interface{}
	Disabled bool
	Icon     string
}

// LayoutType defines the layout type for the UI
type LayoutType string

const (
	LayoutTypeDefault   LayoutType = "default"
	LayoutTypeSplit     LayoutType = "split"
	LayoutTypeFullWidth LayoutType = "fullwidth"
	LayoutTypeCompact   LayoutType = "compact"
)

// UIValidation defines validation rules for UI
type UIValidation struct {
	Required  bool
	MinLength int
	MaxLength int
	Pattern   string
	Custom    func(value interface{}) error
}

// InstallMode represents the installation mode
type InstallMode string

const (
	ModeExpress     InstallMode = "express"
	ModeCustom      InstallMode = "custom"
	ModeAdvanced    InstallMode = "advanced"
	ModeRepair      InstallMode = "repair"
	ModeUninstall   InstallMode = "uninstall"
	ModeUserDefined InstallMode = "user"
)

// WizardProviderRegistry manages wizard providers
type WizardProviderRegistry struct {
	providers map[string]WizardProvider
	default_  string
}

var providerRegistry = &WizardProviderRegistry{
	providers: make(map[string]WizardProvider),
}

// RegisterWizardProvider registers a wizard provider
func RegisterWizardProvider(name string, provider WizardProvider) error {
	if _, exists := providerRegistry.providers[name]; exists {
		return fmt.Errorf("wizard provider %s already registered", name)
	}
	providerRegistry.providers[name] = provider
	return nil
}

// SetDefaultWizardProvider sets the default provider
func SetDefaultWizardProvider(name string) error {
	if _, exists := providerRegistry.providers[name]; !exists {
		return fmt.Errorf("wizard provider %s not found", name)
	}
	providerRegistry.default_ = name
	return nil
}

// GetWizardProvider returns a provider by name
func GetWizardProvider(name string) (WizardProvider, error) {
	if name == "" {
		name = providerRegistry.default_
	}
	provider, exists := providerRegistry.providers[name]
	if !exists {
		return nil, fmt.Errorf("wizard provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultWizardProvider returns the default provider
func GetDefaultWizardProvider() (WizardProvider, error) {
	if providerRegistry.default_ == "" {
		// Return built-in standard provider if no default set
		return NewStandardWizardProvider(ModeExpress), nil
	}
	return GetWizardProvider(providerRegistry.default_)
}

// Standard wizard states used by built-in providers
const (
	StateWelcome     wizard.State = "welcome"
	StateModeSelect  wizard.State = "mode_select"
	StateLicense     wizard.State = "license"
	StateComponents  wizard.State = "components"
	StateLocation    wizard.State = "location"
	StateReady       wizard.State = "ready"
	StateInstalling  wizard.State = "installing"
	StateComplete    wizard.State = "complete"
	StateError       wizard.State = "error"
	StateRollback    wizard.State = "rollback"
)
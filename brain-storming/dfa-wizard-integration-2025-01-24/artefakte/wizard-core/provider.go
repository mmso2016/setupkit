// Package wizard provides DFA-based installation flow control
package wizard

import (
	"context"
	"github.com/mmso2016/setupkit/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// Provider defines the interface for wizard providers
type Provider interface {
	// GetDFA returns the configured DFA for the installation
	GetDFA() (*wizard.DFA, error)
	
	// GetStateHandler returns the handler for a specific state
	GetStateHandler(state wizard.State) StateHandler
	
	// GetUIMapping returns UI configuration for a state
	GetUIMapping(state wizard.State) UIStateConfig
	
	// ValidateConfiguration checks if the DFA is properly configured
	ValidateConfiguration() error
	
	// GetMode returns the installation mode
	GetMode() InstallMode
}

// StateHandler handles the logic for each wizard state
type StateHandler interface {
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
	LayoutTypeDefault  LayoutType = "default"
	LayoutTypeSplit    LayoutType = "split"
	LayoutTypeFullWidth LayoutType = "fullwidth"
	LayoutTypeCompact  LayoutType = "compact"
)

// UIValidation defines validation rules for UI
type UIValidation struct {
	Required bool
	MinLength int
	MaxLength int
	Pattern  string
	Custom   func(value interface{}) error
}

// InstallMode represents the installation mode
type InstallMode string

const (
	ModeExpress   InstallMode = "express"
	ModeCustom    InstallMode = "custom"
	ModeAdvanced  InstallMode = "advanced"
	ModeRepair    InstallMode = "repair"
	ModeUninstall InstallMode = "uninstall"
	ModeUserDefined InstallMode = "user"
)

// ProviderRegistry manages wizard providers
type ProviderRegistry struct {
	providers map[string]Provider
	default_  string
}

var registry = &ProviderRegistry{
	providers: make(map[string]Provider),
}

// Register registers a wizard provider
func Register(name string, provider Provider) error {
	if _, exists := registry.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}
	registry.providers[name] = provider
	return nil
}

// SetDefault sets the default provider
func SetDefault(name string) error {
	if _, exists := registry.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}
	registry.default_ = name
	return nil
}

// GetProvider returns a provider by name
func GetProvider(name string) (Provider, error) {
	if name == "" {
		name = registry.default_
	}
	provider, exists := registry.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefault returns the default provider
func GetDefault() (Provider, error) {
	if registry.default_ == "" {
		// Return built-in standard provider if no default set
		return NewStandardProvider(ModeExpress)
	}
	return GetProvider(registry.default_)
}

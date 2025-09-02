// Package core - Extended Wizard Provider for Theme Selection and Custom States
package core

import (
	"context"
	"fmt"

	"github.com/mmso2016/setupkit/installer/themes"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// ExtendedWizardProvider extends the standard provider with additional states
type ExtendedWizardProvider struct {
	*StandardWizardProvider
	
	// Extended configuration
	insertions  []StateInsertion
	themeConfig *ThemeSelectionConfig
}

// StateInsertion defines how to insert a new state into the DFA
type StateInsertion struct {
	NewState    wizard.State
	AfterState  wizard.State
	Handler     WizardStateHandler
	UIConfig    UIStateConfig
	StateConfig *wizard.StateConfig
}

// ThemeSelectionConfig configures theme selection functionality
type ThemeSelectionConfig struct {
	Enabled        bool
	DefaultTheme   string
	AvailableThemes []string
	ShowPreview    bool
	AllowCustom    bool
}

// Standard extended states
const (
	StateThemeSelection wizard.State = "theme_selection"
)

// NewExtendedWizardProvider creates a new extended wizard provider
func NewExtendedWizardProvider(mode InstallMode) *ExtendedWizardProvider {
	standardProvider := NewStandardWizardProvider(mode)
	
	return &ExtendedWizardProvider{
		StandardWizardProvider: standardProvider,
		insertions:             make([]StateInsertion, 0),
		themeConfig: &ThemeSelectionConfig{
			Enabled:      false,
			DefaultTheme: "default",
			ShowPreview:  true,
			AllowCustom:  false,
		},
	}
}

// EnableThemeSelection enables theme selection with the given configuration
func (ewp *ExtendedWizardProvider) EnableThemeSelection(config *ThemeSelectionConfig) {
	ewp.themeConfig = config
	ewp.themeConfig.Enabled = true
	
	// Add theme selection state insertion
	insertion := StateInsertion{
		NewState:   StateThemeSelection,
		AfterState: StateLicense, // Insert after license, before components
		Handler:    NewThemeSelectionStateHandler(ewp.config, ewp.context, config),
		UIConfig:   ewp.createThemeSelectionUIConfig(config),
		StateConfig: &wizard.StateConfig{
			Name:      "Theme Selection",
			CanGoNext: true,
			CanGoBack: true,
			CanCancel: true,
			CanSkip:   !config.Enabled, // Can skip if not mandatory
			Transitions: map[wizard.Action]wizard.State{
				wizard.ActionNext: StateComponents,
				wizard.ActionSkip: StateComponents,
			},
			ValidateFunc: ewp.validateThemeSelection,
		},
	}
	
	ewp.insertions = append(ewp.insertions, insertion)
}

// InsertCustomState inserts a custom state into the wizard flow
func (ewp *ExtendedWizardProvider) InsertCustomState(insertion StateInsertion) {
	ewp.insertions = append(ewp.insertions, insertion)
}

// Initialize initializes the extended provider
func (ewp *ExtendedWizardProvider) Initialize(config *Config, context *Context) error {
	// Initialize the base provider first
	if err := ewp.StandardWizardProvider.Initialize(config, context); err != nil {
		return fmt.Errorf("failed to initialize standard provider: %w", err)
	}
	
	// Apply state insertions
	if err := ewp.applyStateInsertions(); err != nil {
		return fmt.Errorf("failed to apply state insertions: %w", err)
	}
	
	return nil
}

// applyStateInsertions applies all configured state insertions to the DFA
func (ewp *ExtendedWizardProvider) applyStateInsertions() error {
	if len(ewp.insertions) == 0 {
		return nil
	}
	
	// Apply insertions in order
	for _, insertion := range ewp.insertions {
		if err := ewp.insertState(insertion); err != nil {
			return fmt.Errorf("failed to insert state %s: %w", insertion.NewState, err)
		}
	}
	
	return nil
}

// insertState inserts a single state into the DFA
func (ewp *ExtendedWizardProvider) insertState(insertion StateInsertion) error {
	// Add the new state to the DFA
	if err := ewp.dfa.AddState(insertion.NewState, insertion.StateConfig); err != nil {
		return fmt.Errorf("failed to add state to DFA: %w", err)
	}
	
	// Update the "after" state to transition to the new state instead of its original target
	afterStateConfig, err := ewp.dfa.GetStateConfig(insertion.AfterState)
	if err != nil {
		return fmt.Errorf("failed to get after state config: %w", err)
	}
	
	// Find the original next state and update transitions
	var originalNext wizard.State
	if afterStateConfig.NextStateFunc != nil {
		// This is more complex - we'd need to wrap the function
		return fmt.Errorf("states with NextStateFunc are not yet supported for insertion")
	} else if nextState, exists := afterStateConfig.Transitions[wizard.ActionNext]; exists {
		originalNext = nextState
		
		// Update the after state to point to the new state
		afterStateConfig.Transitions[wizard.ActionNext] = insertion.NewState
		
		// Update the new state to point to the original next state
		if insertion.StateConfig.Transitions == nil {
			insertion.StateConfig.Transitions = make(map[wizard.Action]wizard.State)
		}
		insertion.StateConfig.Transitions[wizard.ActionNext] = originalNext
		insertion.StateConfig.Transitions[wizard.ActionSkip] = originalNext
	} else {
		return fmt.Errorf("after state %s has no next transition", insertion.AfterState)
	}
	
	// Register the handler and UI mapping
	ewp.handlers[insertion.NewState] = insertion.Handler
	ewp.uiMappings[insertion.NewState] = insertion.UIConfig
	
	ewp.context.Logger.Info("Inserted wizard state",
		"state", insertion.NewState,
		"after", insertion.AfterState,
		"original_next", originalNext)
	
	return nil
}

// createThemeSelectionUIConfig creates UI configuration for theme selection
func (ewp *ExtendedWizardProvider) createThemeSelectionUIConfig(config *ThemeSelectionConfig) UIStateConfig {
	// Get available themes
	availableThemes := config.AvailableThemes
	if len(availableThemes) == 0 {
		// Use all available themes
		builtinThemes := themes.GetBuiltinThemes()
		for name := range builtinThemes {
			availableThemes = append(availableThemes, name)
		}
	}
	
	// Create theme options
	themeOptions := make([]FieldOption, len(availableThemes))
	for i, themeName := range availableThemes {
		themeOptions[i] = FieldOption{
			ID:    themeName,
			Label: themeName,
			Value: themeName,
		}
	}
	
	fields := []UIField{
		{
			ID:       "selected_theme",
			Label:    "Choose Installation Theme",
			Type:     FieldTypeTheme, // Custom field type for theme selection
			Value:    config.DefaultTheme,
			Required: true,
			Options:  themeOptions,
			Help:     "Select the visual theme for the installer interface.",
		},
	}
	
	// Add preview field if enabled
	if config.ShowPreview {
		fields = append(fields, UIField{
			ID:   "theme_preview",
			Type: "preview", // Custom field type for preview
			Help: "Preview of the selected theme",
		})
	}
	
	actions := []StateAction{
		{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
		{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
	}
	
	// Add skip action if theme selection is not mandatory
	if !config.Enabled {
		actions = append(actions, StateAction{
			ID: "skip", Label: "Skip", Type: ActionTypeSkip, Primary: false, Enabled: true, Visible: true,
		})
	}
	
	actions = append(actions, StateAction{
		ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true,
	})
	
	return UIStateConfig{
		Title:       "Choose Theme",
		Description: "Select the visual theme for the installation interface.",
		Type:        UIStateTypeSelection,
		Layout:      LayoutTypeDefault,
		Fields:      fields,
		Actions:     actions,
		Template:    "theme_selection", // Custom template for theme selection
	}
}

// validateThemeSelection validates theme selection
func (ewp *ExtendedWizardProvider) validateThemeSelection(data map[string]interface{}) error {
	if !ewp.themeConfig.Enabled {
		return nil // Validation not required if theme selection is optional
	}
	
	selectedTheme, ok := data["selected_theme"]
	if !ok || selectedTheme == "" {
		return fmt.Errorf("theme must be selected")
	}
	
	// Validate theme exists
	themeName := selectedTheme.(string)
	if _, err := themes.GetTheme(themeName); err != nil {
		return fmt.Errorf("invalid theme selected: %s", themeName)
	}
	
	return nil
}

// ThemeSelectionStateHandler handles theme selection logic
type ThemeSelectionStateHandler struct {
	BaseStateHandler
	themeConfig *ThemeSelectionConfig
}

// NewThemeSelectionStateHandler creates a new theme selection state handler
func NewThemeSelectionStateHandler(config *Config, context *Context, themeConfig *ThemeSelectionConfig) *ThemeSelectionStateHandler {
	return &ThemeSelectionStateHandler{
		BaseStateHandler: BaseStateHandler{
			config:  config,
			context: context,
			title:   "Choose Theme",
			desc:    "Select the visual theme for the installation interface.",
		},
		themeConfig: themeConfig,
	}
}

// Execute performs theme selection logic
func (tsh *ThemeSelectionStateHandler) Execute(ctx context.Context, data map[string]interface{}) error {
	// Set available themes
	availableThemes := tsh.themeConfig.AvailableThemes
	if len(availableThemes) == 0 {
		builtinThemes := themes.GetBuiltinThemes()
		for name := range builtinThemes {
			availableThemes = append(availableThemes, name)
		}
	}
	data["available_themes"] = availableThemes
	
	// Set default theme if not already selected
	if _, exists := data["selected_theme"]; !exists {
		data["selected_theme"] = tsh.themeConfig.DefaultTheme
	}
	
	// Apply theme if one is selected
	if selectedTheme, ok := data["selected_theme"]; ok && selectedTheme != "" {
		if err := tsh.applyTheme(selectedTheme.(string)); err != nil {
			tsh.context.Logger.Warn("Failed to apply theme", "theme", selectedTheme, "error", err)
		}
	}
	
	return nil
}

// applyTheme applies the selected theme to the installer configuration
func (tsh *ThemeSelectionStateHandler) applyTheme(themeName string) error {
	theme, err := themes.GetTheme(themeName)
	if err != nil {
		return fmt.Errorf("failed to get theme: %w", err)
	}
	
	// Apply theme to config
	tsh.config.Theme = theme
	
	// Apply theme using the existing ApplyTheme method
	if err := tsh.config.ApplyTheme(); err != nil {
		return fmt.Errorf("failed to apply theme: %w", err)
	}
	
	tsh.context.Logger.Info("Applied theme", "theme", themeName)
	return nil
}

// Validate validates theme selection
func (tsh *ThemeSelectionStateHandler) Validate(data map[string]interface{}) error {
	if !tsh.themeConfig.Enabled {
		return nil // Not required
	}
	
	selectedTheme, ok := data["selected_theme"]
	if !ok || selectedTheme == "" {
		return fmt.Errorf("theme must be selected")
	}
	
	// Validate theme exists
	themeName := selectedTheme.(string)
	if _, err := themes.GetTheme(themeName); err != nil {
		return fmt.Errorf("invalid theme selected: %s", themeName)
	}
	
	return nil
}

// GetActions returns theme selection specific actions
func (tsh *ThemeSelectionStateHandler) GetActions() []StateAction {
	actions := []StateAction{
		{ID: "next", Label: "Next", Type: ActionTypeNext, Primary: true, Enabled: true, Visible: true},
		{ID: "back", Label: "Back", Type: ActionTypeBack, Primary: false, Enabled: true, Visible: true},
	}
	
	// Add skip if not mandatory
	if !tsh.themeConfig.Enabled {
		actions = append(actions, StateAction{
			ID: "skip", Label: "Skip", Type: ActionTypeSkip, Primary: false, Enabled: true, Visible: true,
		})
	}
	
	actions = append(actions, StateAction{
		ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, Primary: false, Enabled: true, Visible: true,
	})
	
	return actions
}

// CreateExtendedProviderWithThemes creates an extended provider with theme selection enabled
func CreateExtendedProviderWithThemes(mode InstallMode, themes []string, defaultTheme string) *ExtendedWizardProvider {
	provider := NewExtendedWizardProvider(mode)
	
	themeConfig := &ThemeSelectionConfig{
		Enabled:         true,
		DefaultTheme:    defaultTheme,
		AvailableThemes: themes,
		ShowPreview:     true,
		AllowCustom:     false,
	}
	
	provider.EnableThemeSelection(themeConfig)
	return provider
}

// GetInsertedStates returns all inserted states
func (ewp *ExtendedWizardProvider) GetInsertedStates() []wizard.State {
	states := make([]wizard.State, len(ewp.insertions))
	for i, insertion := range ewp.insertions {
		states[i] = insertion.NewState
	}
	return states
}

// IsExtendedState checks if a state is an extended (inserted) state
func (ewp *ExtendedWizardProvider) IsExtendedState(state wizard.State) bool {
	for _, insertion := range ewp.insertions {
		if insertion.NewState == state {
			return true
		}
	}
	return false
}
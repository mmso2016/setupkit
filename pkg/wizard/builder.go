// Package wizard - Builder for creating DFA wizards
package wizard

import (
	"fmt"
)

// Builder provides a fluent interface for building DFA wizards
type Builder struct {
	dfa           *DFA
	currentState  State
	lastError     error
	stateDefaults *StateConfig
}

// NewBuilder creates a new wizard builder
func NewBuilder() *Builder {
	return &Builder{
		dfa: New(),
		stateDefaults: &StateConfig{
			CanGoBack: true,
			CanCancel: true,
		},
	}
}

// WithDefaults sets default configuration for all states
func (b *Builder) WithDefaults(defaults *StateConfig) *Builder {
	if b.lastError != nil {
		return b
	}
	b.stateDefaults = defaults
	return b
}

// WithMaxHistory sets the maximum history size
func (b *Builder) WithMaxHistory(max int) *Builder {
	if b.lastError != nil {
		return b
	}
	b.dfa.SetMaxHistory(max)
	return b
}

// WithStrictMode enables or disables strict validation
func (b *Builder) WithStrictMode(strict bool) *Builder {
	if b.lastError != nil {
		return b
	}
	b.dfa.SetStrictMode(strict)
	return b
}

// WithCallbacks sets global callbacks
func (b *Builder) WithCallbacks(callbacks *Callbacks) *Builder {
	if b.lastError != nil {
		return b
	}
	b.dfa.SetCallbacks(callbacks)
	return b
}

// AddState adds a new state with configuration
func (b *Builder) AddState(state State, config *StateConfig) *Builder {
	if b.lastError != nil {
		return b
	}
	
	// Merge with defaults
	mergedConfig := b.mergeWithDefaults(config)
	
	err := b.dfa.AddState(state, mergedConfig)
	if err != nil {
		b.lastError = err
		return b
	}
	
	b.currentState = state
	return b
}

// State starts configuration for a new state
func (b *Builder) State(name State) *StateBuilder {
	return &StateBuilder{
		builder: b,
		state:   name,
		config:  b.mergeWithDefaults(&StateConfig{Name: string(name)}),
	}
}

// Initial sets the initial state
func (b *Builder) Initial(state State) *Builder {
	if b.lastError != nil {
		return b
	}
	
	err := b.dfa.SetInitialState(state)
	if err != nil {
		b.lastError = err
	}
	return b
}

// Final marks a state as final
func (b *Builder) Final(state State) *Builder {
	if b.lastError != nil {
		return b
	}
	
	err := b.dfa.AddFinalState(state)
	if err != nil {
		b.lastError = err
	}
	return b
}

// Transition adds a global transition rule
func (b *Builder) Transition(from, to State, action Action) *Builder {
	if b.lastError != nil {
		return b
	}
	
	err := b.dfa.AddTransition(TransitionRule{
		From:   from,
		To:     to,
		Action: action,
	})
	if err != nil {
		b.lastError = err
	}
	return b
}

// ConditionalTransition adds a conditional transition rule
func (b *Builder) ConditionalTransition(from, to State, action Action, condition func(map[string]interface{}) bool) *Builder {
	if b.lastError != nil {
		return b
	}
	
	err := b.dfa.AddTransition(TransitionRule{
		From:      from,
		To:        to,
		Action:    action,
		Condition: condition,
	})
	if err != nil {
		b.lastError = err
	}
	return b
}

// Build creates the configured DFA
func (b *Builder) Build() (*DFA, error) {
	if b.lastError != nil {
		return nil, b.lastError
	}
	
	// Validate the DFA
	if err := b.dfa.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	return b.dfa, nil
}

// mergeWithDefaults merges a config with the default config
func (b *Builder) mergeWithDefaults(config *StateConfig) *StateConfig {
	if config == nil {
		config = &StateConfig{}
	}
	
	// Create a copy of the config
	merged := *config
	
	// Apply defaults if not explicitly set
	if b.stateDefaults != nil {
		if !merged.CanGoBack && b.stateDefaults.CanGoBack {
			merged.CanGoBack = true
		}
		if !merged.CanGoNext && b.stateDefaults.CanGoNext {
			merged.CanGoNext = true
		}
		if !merged.CanCancel && b.stateDefaults.CanCancel {
			merged.CanCancel = true
		}
		if !merged.CanSkip && b.stateDefaults.CanSkip {
			merged.CanSkip = true
		}
	}
	
	if merged.Transitions == nil {
		merged.Transitions = make(map[Action]State)
	}
	
	return &merged
}

// StateBuilder provides a fluent interface for building a single state
type StateBuilder struct {
	builder *Builder
	state   State
	config  *StateConfig
}

// Named sets the display name
func (s *StateBuilder) Named(name string) *StateBuilder {
	s.config.Name = name
	return s
}

// Description sets the description
func (s *StateBuilder) Description(desc string) *StateBuilder {
	s.config.Description = desc
	return s
}

// CanGoNext enables/disables next navigation
func (s *StateBuilder) CanGoNext(can bool) *StateBuilder {
	s.config.CanGoNext = can
	return s
}

// CanGoBack enables/disables back navigation
func (s *StateBuilder) CanGoBack(can bool) *StateBuilder {
	s.config.CanGoBack = can
	return s
}

// CanSkip enables/disables skip functionality
func (s *StateBuilder) CanSkip(can bool) *StateBuilder {
	s.config.CanSkip = can
	return s
}

// CanCancel enables/disables cancel functionality
func (s *StateBuilder) CanCancel(can bool) *StateBuilder {
	s.config.CanCancel = can
	return s
}

// OnEnter sets the OnEnter callback
func (s *StateBuilder) OnEnter(fn func(data map[string]interface{}) error) *StateBuilder {
	s.config.OnEnter = fn
	return s
}

// OnExit sets the OnExit callback
func (s *StateBuilder) OnExit(fn func(data map[string]interface{}) error) *StateBuilder {
	s.config.OnExit = fn
	return s
}

// Validate sets the validation function
func (s *StateBuilder) Validate(fn func(data map[string]interface{}) error) *StateBuilder {
	s.config.ValidateFunc = fn
	return s
}

// ValidateOnEntry sets entry validation
func (s *StateBuilder) ValidateOnEntry(fn func(data map[string]interface{}) error) *StateBuilder {
	s.config.ValidateOnEntry = fn
	return s
}

// ValidateOnExit sets exit validation
func (s *StateBuilder) ValidateOnExit(fn func(data map[string]interface{}) error) *StateBuilder {
	s.config.ValidateOnExit = fn
	return s
}

// CanEnter sets the entry condition
func (s *StateBuilder) CanEnter(fn func(data map[string]interface{}) bool) *StateBuilder {
	s.config.CanEnterFunc = fn
	return s
}

// NextState sets dynamic next state determination
func (s *StateBuilder) NextState(fn func(data map[string]interface{}) (State, error)) *StateBuilder {
	s.config.NextStateFunc = fn
	return s
}

// TransitionTo adds a transition to another state
func (s *StateBuilder) TransitionTo(action Action, target State) *StateBuilder {
	if s.config.Transitions == nil {
		s.config.Transitions = make(map[Action]State)
	}
	s.config.Transitions[action] = target
	return s
}

// Next sets the next state
func (s *StateBuilder) Next(target State) *StateBuilder {
	return s.TransitionTo(ActionNext, target)
}

// Skip sets the skip target
func (s *StateBuilder) Skip(target State) *StateBuilder {
	return s.TransitionTo(ActionSkip, target)
}

// Add adds the state to the DFA
func (s *StateBuilder) Add() *Builder {
	return s.builder.AddState(s.state, s.config)
}

// And is an alias for Add for better readability
func (s *StateBuilder) And() *Builder {
	return s.Add()
}

// WizardBuilder provides templates for common wizard patterns
type WizardBuilder struct {
	builder *Builder
}

// NewWizardBuilder creates a new wizard builder with templates
func NewWizardBuilder() *WizardBuilder {
	return &WizardBuilder{
		builder: NewBuilder(),
	}
}

// SimpleInstaller creates a simple installation wizard
func (w *WizardBuilder) SimpleInstaller() *Builder {
	b := w.builder
	
	// Welcome
	b.State("welcome").
		Named("Welcome").
		CanGoNext(true).
		CanGoBack(false).
		Next("license").
		Add()
	
	// License
	b.State("license").
		Named("License Agreement").
		CanGoNext(true).
		CanSkip(false).
		Validate(func(data map[string]interface{}) error {
			if data["license_accepted"] != true {
				return fmt.Errorf("license must be accepted")
			}
			return nil
		}).
		Next("location").
		Add()
	
	// Install Location
	b.State("location").
		Named("Installation Location").
		CanGoNext(true).
		Validate(func(data map[string]interface{}) error {
			if data["install_path"] == nil || data["install_path"] == "" {
				return fmt.Errorf("installation path required")
			}
			return nil
		}).
		Next("confirm").
		Add()
	
	// Confirmation
	b.State("confirm").
		Named("Ready to Install").
		CanGoNext(true).
		Next("installing").
		Add()
	
	// Installing
	b.State("installing").
		Named("Installing").
		CanGoNext(true).
		CanGoBack(false).
		CanCancel(false).
		Next("complete").
		Add()
	
	// Complete
	b.State("complete").
		Named("Installation Complete").
		CanGoNext(false).
		CanGoBack(false).
		Add()
	
	// Set initial and final states
	b.Initial("welcome")
	b.Final("complete")
	
	return b
}

// ConfigurationWizard creates a configuration wizard
func (w *WizardBuilder) ConfigurationWizard() *Builder {
	b := w.builder
	
	// Start
	b.State("start").
		Named("Configuration Wizard").
		CanGoNext(true).
		CanGoBack(false).
		Next("database").
		Add()
	
	// Database Configuration
	b.State("database").
		Named("Database Settings").
		CanGoNext(true).
		Validate(func(data map[string]interface{}) error {
			if data["db_host"] == nil || data["db_host"] == "" {
				return fmt.Errorf("database host required")
			}
			if data["db_name"] == nil || data["db_name"] == "" {
				return fmt.Errorf("database name required")
			}
			return nil
		}).
		Next("network").
		Add()
	
	// Network Configuration
	b.State("network").
		Named("Network Settings").
		CanGoNext(true).
		CanSkip(true).
		Next("security").
		Skip("security").
		Add()
	
	// Security Settings
	b.State("security").
		Named("Security Settings").
		CanGoNext(true).
		Next("review").
		Add()
	
	// Review
	b.State("review").
		Named("Review Configuration").
		CanGoNext(true).
		Next("apply").
		Add()
	
	// Apply Configuration
	b.State("apply").
		Named("Applying Configuration").
		CanGoNext(true).
		CanGoBack(false).
		CanCancel(false).
		Next("complete").
		Add()
	
	// Complete
	b.State("complete").
		Named("Configuration Complete").
		CanGoNext(false).
		CanGoBack(false).
		Add()
	
	b.Initial("start")
	b.Final("complete")
	
	return b
}

// MultiPathWizard creates a wizard with multiple paths
func (w *WizardBuilder) MultiPathWizard() *Builder {
	b := w.builder
	
	// Start
	b.State("start").
		Named("Choose Setup Type").
		CanGoNext(true).
		CanGoBack(false).
		NextState(func(data map[string]interface{}) (State, error) {
			setupType := data["setup_type"]
			switch setupType {
			case "express":
				return "express_setup", nil
			case "custom":
				return "custom_options", nil
			case "advanced":
				return "advanced_warning", nil
			default:
				return "", fmt.Errorf("please select a setup type")
			}
		}).
		Add()
	
	// Express Path
	b.State("express_setup").
		Named("Express Setup").
		CanGoNext(true).
		Next("express_confirm").
		Add()
	
	b.State("express_confirm").
		Named("Confirm Express Settings").
		CanGoNext(true).
		Next("installing").
		Add()
	
	// Custom Path
	b.State("custom_options").
		Named("Custom Options").
		CanGoNext(true).
		Next("custom_components").
		Add()
	
	b.State("custom_components").
		Named("Select Components").
		CanGoNext(true).
		Next("custom_confirm").
		Add()
	
	b.State("custom_confirm").
		Named("Confirm Custom Settings").
		CanGoNext(true).
		Next("installing").
		Add()
	
	// Advanced Path
	b.State("advanced_warning").
		Named("Advanced Mode Warning").
		CanGoNext(true).
		Next("advanced_system").
		Add()
	
	b.State("advanced_system").
		Named("System Configuration").
		CanGoNext(true).
		Next("advanced_network").
		Add()
	
	b.State("advanced_network").
		Named("Network Configuration").
		CanGoNext(true).
		Next("advanced_security").
		Add()
	
	b.State("advanced_security").
		Named("Security Configuration").
		CanGoNext(true).
		Next("advanced_confirm").
		Add()
	
	b.State("advanced_confirm").
		Named("Confirm Advanced Settings").
		CanGoNext(true).
		Next("installing").
		Add()
	
	// Common Final States
	b.State("installing").
		Named("Installing").
		CanGoNext(true).
		CanGoBack(false).
		CanCancel(false).
		Next("complete").
		Add()
	
	b.State("complete").
		Named("Setup Complete").
		CanGoNext(false).
		CanGoBack(false).
		Add()
	
	b.Initial("start")
	b.Final("complete")
	
	return b
}

// Build finalizes the wizard
func (w *WizardBuilder) Build() (*DFA, error) {
	return w.builder.Build()
}

// QuickWizard creates a simple wizard with minimal configuration
func QuickWizard(states ...State) (*DFA, error) {
	if len(states) < 2 {
		return nil, fmt.Errorf("wizard requires at least 2 states")
	}
	
	b := NewBuilder()
	
	for i, state := range states {
		config := &StateConfig{
			Name:      string(state),
			CanGoNext: i < len(states)-1,
			CanGoBack: i > 0,
			CanCancel: true,
		}
		
		// Add next transition
		if i < len(states)-1 {
			config.Transitions = map[Action]State{
				ActionNext: states[i+1],
			}
		}
		
		b.AddState(state, config)
	}
	
	b.Initial(states[0])
	b.Final(states[len(states)-1])
	
	return b.Build()
}

// Package core - UI Adapter for DFA-based Wizard System
package core

import (
	"context"
	"fmt"

	"github.com/mmso2016/setupkit/pkg/wizard"
)

// WizardUIAdapter adapts the DFA-based wizard system to the existing UI interface
type WizardUIAdapter struct {
	provider WizardProvider
	dfa      *wizard.DFA
	context  *Context
	config   *Config
	
	// Current state tracking
	currentState wizard.State
	wizardData   map[string]interface{}
}

// NewWizardUIAdapter creates a new wizard UI adapter
func NewWizardUIAdapter(provider WizardProvider) *WizardUIAdapter {
	return &WizardUIAdapter{
		provider:   provider,
		wizardData: make(map[string]interface{}),
	}
}

// Initialize initializes the wizard UI adapter
func (wua *WizardUIAdapter) Initialize(context *Context) error {
	wua.context = context
	wua.config = context.Config
	
	// Initialize the provider
	if err := wua.provider.Initialize(wua.config, context); err != nil {
		return fmt.Errorf("failed to initialize wizard provider: %w", err)
	}
	
	// Get the DFA from the provider
	dfa, err := wua.provider.GetDFA()
	if err != nil {
		return fmt.Errorf("failed to get DFA from provider: %w", err)
	}
	wua.dfa = dfa
	
	// Set up DFA callbacks
	wua.setupDFACallbacks()
	
	// Get initial state
	wua.currentState = wua.dfa.CurrentState()
	
	context.Logger.Info("Wizard UI adapter initialized",
		"provider", fmt.Sprintf("%T", wua.provider),
		"initial_state", wua.currentState)
	
	return nil
}

// setupDFACallbacks sets up callbacks for DFA state changes
func (wua *WizardUIAdapter) setupDFACallbacks() {
	callbacks := &wizard.Callbacks{
		OnEnter: func(state wizard.State, data map[string]interface{}) error {
			wua.currentState = state
			
			// Execute state handler
			if handler := wua.provider.GetStateHandler(state); handler != nil {
				ctx := context.Background()
				if err := handler.OnEnter(ctx, wua.wizardData); err != nil {
					return fmt.Errorf("state handler OnEnter failed for %s: %w", state, err)
				}
				
				if err := handler.Execute(ctx, wua.wizardData); err != nil {
					return fmt.Errorf("state handler Execute failed for %s: %w", state, err)
				}
			}
			
			wua.context.Logger.Debug("Entered wizard state", "state", state)
			return nil
		},
		
		OnLeave: func(state wizard.State, data map[string]interface{}) error {
			// Execute state handler OnExit
			if handler := wua.provider.GetStateHandler(state); handler != nil {
				ctx := context.Background()
				if err := handler.OnExit(ctx, wua.wizardData); err != nil {
					return fmt.Errorf("state handler OnExit failed for %s: %w", state, err)
				}
			}
			
			wua.context.Logger.Debug("Left wizard state", "state", state)
			return nil
		},
		
		OnTransition: func(from, to wizard.State, action wizard.Action) error {
			wua.context.Logger.Info("Wizard transition", 
				"from", from, 
				"to", to, 
				"action", action)
			return nil
		},
		
		OnDataChange: func(state wizard.State, key string, oldValue, newValue interface{}) error {
			wua.context.Logger.Debug("Wizard data changed",
				"state", state,
				"key", key,
				"old", oldValue,
				"new", newValue)
			return nil
		},
		
		OnValidationError: func(state wizard.State, err error) {
			wua.context.Logger.Warn("Wizard validation error",
				"state", state,
				"error", err.Error())
		},
	}
	
	wua.dfa.SetCallbacks(callbacks)
}

// GetCurrentState returns the current wizard state
func (wua *WizardUIAdapter) GetCurrentState() wizard.State {
	return wua.currentState
}

// GetCurrentStateConfig returns the UI configuration for the current state
func (wua *WizardUIAdapter) GetCurrentStateConfig() UIStateConfig {
	return wua.provider.GetUIMapping(wua.currentState)
}

// GetCurrentStateHandler returns the handler for the current state
func (wua *WizardUIAdapter) GetCurrentStateHandler() WizardStateHandler {
	return wua.provider.GetStateHandler(wua.currentState)
}

// GetWizardData returns the current wizard data
func (wua *WizardUIAdapter) GetWizardData() map[string]interface{} {
	// Merge DFA data with wizard data
	dfaData := wua.dfa.GetAllData()
	result := make(map[string]interface{})
	
	// Copy DFA data
	for k, v := range dfaData {
		result[k] = v
	}
	
	// Copy wizard-specific data
	for k, v := range wua.wizardData {
		result[k] = v
	}
	
	return result
}

// SetWizardData sets data in the wizard
func (wua *WizardUIAdapter) SetWizardData(key string, value interface{}) error {
	// Store in both wizard data and DFA data
	wua.wizardData[key] = value
	return wua.dfa.SetData(key, value)
}

// CanPerformAction checks if an action can be performed in the current state
func (wua *WizardUIAdapter) CanPerformAction(action ActionType) bool {
	// Map UI action types to DFA actions
	dfaAction := wua.mapActionTypeToDFAAction(action)
	return wua.dfa.CanTransition(dfaAction)
}

// PerformAction performs an action in the current state
func (wua *WizardUIAdapter) PerformAction(action ActionType, data map[string]interface{}) error {
	// Update wizard data with provided data
	for k, v := range data {
		if err := wua.SetWizardData(k, v); err != nil {
			return fmt.Errorf("failed to set wizard data: %w", err)
		}
	}
	
	// Validate current state if not a back action
	if action != ActionTypeBack {
		if handler := wua.provider.GetStateHandler(wua.currentState); handler != nil {
			if err := handler.Validate(wua.GetWizardData()); err != nil {
				return fmt.Errorf("state validation failed: %w", err)
			}
		}
	}
	
	// Perform the transition
	dfaAction := wua.mapActionTypeToDFAAction(action)
	
	switch dfaAction {
	case wizard.ActionNext:
		return wua.dfa.Next()
	case wizard.ActionBack:
		return wua.dfa.Back()
	case wizard.ActionSkip:
		return wua.dfa.Skip()
	case wizard.ActionCancel:
		return wua.dfa.Cancel()
	default:
		return wua.dfa.Transition(dfaAction)
	}
}

// IsInFinalState checks if the wizard is in a final state
func (wua *WizardUIAdapter) IsInFinalState() bool {
	return wua.dfa.IsInFinalState()
}

// GetAvailableActions returns the available actions for the current state
func (wua *WizardUIAdapter) GetAvailableActions() []StateAction {
	if handler := wua.provider.GetStateHandler(wua.currentState); handler != nil {
		return handler.GetActions()
	}
	
	// Fallback: generate actions based on DFA capabilities
	var actions []StateAction
	
	if wua.dfa.CanTransition(wizard.ActionNext) {
		actions = append(actions, StateAction{
			ID: "next", Label: "Next", Type: ActionTypeNext, 
			Primary: true, Enabled: true, Visible: true,
		})
	}
	
	if wua.dfa.CanTransition(wizard.ActionBack) {
		actions = append(actions, StateAction{
			ID: "back", Label: "Back", Type: ActionTypeBack, 
			Primary: false, Enabled: true, Visible: true,
		})
	}
	
	if wua.dfa.CanTransition(wizard.ActionCancel) {
		actions = append(actions, StateAction{
			ID: "cancel", Label: "Cancel", Type: ActionTypeCancel, 
			Primary: false, Enabled: true, Visible: true,
		})
	}
	
	return actions
}

// GetStateHistory returns the state history
func (wua *WizardUIAdapter) GetStateHistory() []wizard.State {
	return wua.dfa.GetHistory()
}

// Reset resets the wizard to its initial state
func (wua *WizardUIAdapter) Reset() error {
	wua.dfa.Reset()
	wua.wizardData = make(map[string]interface{})
	wua.currentState = wua.dfa.CurrentState()
	return nil
}

// ValidateCurrentState validates the current state
func (wua *WizardUIAdapter) ValidateCurrentState() error {
	if handler := wua.provider.GetStateHandler(wua.currentState); handler != nil {
		return handler.Validate(wua.GetWizardData())
	}
	return nil
}

// mapActionTypeToDFAAction maps UI action types to DFA actions
func (wua *WizardUIAdapter) mapActionTypeToDFAAction(actionType ActionType) wizard.Action {
	switch actionType {
	case ActionTypeNext:
		return wizard.ActionNext
	case ActionTypeBack:
		return wizard.ActionBack
	case ActionTypeSkip:
		return wizard.ActionSkip
	case ActionTypeCancel:
		return wizard.ActionCancel
	case ActionTypeFinish:
		return wizard.ActionNext // Finish is typically just a next action
	default:
		return wizard.Action(string(actionType))
	}
}

// GetDryRunLog returns the DFA dry-run log (if in dry-run mode)
func (wua *WizardUIAdapter) GetDryRunLog() []string {
	return wua.dfa.GetDryRunLog()
}

// WizardUIInterface defines the interface that UIs can use to interact with the wizard
type WizardUIInterface interface {
	// State management
	GetCurrentState() wizard.State
	GetCurrentStateConfig() UIStateConfig
	GetCurrentStateHandler() WizardStateHandler
	IsInFinalState() bool
	
	// Data management
	GetWizardData() map[string]interface{}
	SetWizardData(key string, value interface{}) error
	
	// Action management
	CanPerformAction(action ActionType) bool
	PerformAction(action ActionType, data map[string]interface{}) error
	GetAvailableActions() []StateAction
	
	// State validation
	ValidateCurrentState() error
	
	// Utility methods
	GetStateHistory() []wizard.State
	Reset() error
	GetDryRunLog() []string
}

// Ensure WizardUIAdapter implements WizardUIInterface
var _ WizardUIInterface = (*WizardUIAdapter)(nil)
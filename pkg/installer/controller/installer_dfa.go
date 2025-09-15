// Package controller provides DFA-based installation flow control
package controller

import (
	"fmt"
	
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// InstallerStates defines all states in the installation flow
const (
	StateWelcome     wizard.State = "welcome"
	StateLicense     wizard.State = "license"
	StateComponents  wizard.State = "components"
	StateInstallPath wizard.State = "install-path"
	StateSummary     wizard.State = "summary"
	StateProgress    wizard.State = "progress"
	StateComplete    wizard.State = "complete"
	StateCancelled   wizard.State = "cancelled"
)

// InstallerController manages the installation flow using DFA
type InstallerController struct {
	dfa        *wizard.DFA
	config     *core.Config
	installer  *core.Installer

	// View interface - both CLI and GUI implement this
	view       InstallerView

	// Custom state support
	customStates *CustomStateRegistry
	stateData    map[string]interface{}
}

// InstallerView interface that both CLI and GUI must implement
type InstallerView interface {
	// State display methods
	ShowWelcome() error
	ShowLicense(license string) (accepted bool, err error)
	ShowComponents(components []core.Component) (selected []core.Component, err error)
	ShowInstallPath(defaultPath string) (path string, err error)
	ShowSummary(config *core.Config, selectedComponents []core.Component, installPath string) (proceed bool, err error)
	ShowProgress(progress *core.Progress) error
	ShowComplete(summary *core.InstallSummary) error
	ShowErrorMessage(err error) error
	
	// State change notification
	OnStateChanged(oldState, newState wizard.State) error
}

// NewInstallerController creates a new DFA-based installer controller
func NewInstallerController(config *core.Config, installer *core.Installer) *InstallerController {
	controller := &InstallerController{
		dfa:          wizard.New(),
		config:       config,
		installer:    installer,
		customStates: NewCustomStateRegistry(),
		stateData:    make(map[string]interface{}),
	}

	controller.setupDFA()
	return controller
}

// SetView sets the view implementation (CLI or GUI)
func (ic *InstallerController) SetView(view InstallerView) {
	ic.view = view
}

// RegisterCustomState adds a custom state to the installation flow
func (ic *InstallerController) RegisterCustomState(handler CustomStateHandler) error {
	if err := ic.customStates.Register(handler); err != nil {
		return err
	}

	// Rebuild DFA with custom states included
	ic.setupDFA()
	return nil
}

// GetCustomStates returns all registered custom states
func (ic *InstallerController) GetCustomStates() []CustomStateHandler {
	return ic.customStates.GetAll()
}

// GetStateData returns the current state data
func (ic *InstallerController) GetStateData() map[string]interface{} {
	return ic.stateData
}

// setupDFA configures the DFA states and transitions
func (ic *InstallerController) setupDFA() {
	// Clear existing DFA and create a new one to avoid duplicate states
	ic.dfa = wizard.New()

	// Configure DFA callbacks
	callbacks := &wizard.Callbacks{
		OnEnter: func(state wizard.State, data map[string]interface{}) error {
			return ic.handleStateEnter(state, data)
		},
		OnLeave: func(state wizard.State, data map[string]interface{}) error {
			return ic.handleStateLeave(state, data)
		},
		OnTransition: func(from, to wizard.State, action wizard.Action) error {
			if ic.view != nil {
				return ic.view.OnStateChanged(from, to)
			}
			return nil
		},
	}
	ic.dfa.SetCallbacks(callbacks)
	
	// Add states with their configurations
	ic.addState(StateWelcome, &wizard.StateConfig{
		Name:        "Welcome",
		Description: "Welcome screen",
		CanGoNext:   true,
		CanCancel:   true,
		Transitions: map[wizard.Action]wizard.State{
			wizard.ActionNext:   ic.getNextStateAfterWelcome(),
			wizard.ActionCancel: StateCancelled,
		},
	})
	
	ic.addState(StateLicense, &wizard.StateConfig{
		Name:        "License Agreement",
		Description: "License acceptance screen",
		CanGoNext:   true,
		CanGoBack:   true,
		CanCancel:   true,
		ValidateFunc: ic.validateLicense,
		Transitions: map[wizard.Action]wizard.State{
			wizard.ActionNext:   StateComponents,
			wizard.ActionBack:   StateWelcome,
			wizard.ActionCancel: StateCancelled,
		},
	})
	
	ic.addState(StateComponents, &wizard.StateConfig{
		Name:        "Component Selection",
		Description: "Select components to install",
		CanGoNext:   true,
		CanGoBack:   true,
		CanCancel:   true,
		ValidateFunc: ic.validateComponents,
		Transitions: map[wizard.Action]wizard.State{
			wizard.ActionNext:   StateInstallPath,
			wizard.ActionBack:   ic.getPrevStateBeforeComponents(),
			wizard.ActionCancel: StateCancelled,
		},
	})
	
	ic.addState(StateInstallPath, &wizard.StateConfig{
		Name:        "Installation Path",
		Description: "Select installation directory",
		CanGoNext:   true,
		CanGoBack:   true,
		CanCancel:   true,
		ValidateFunc: ic.validateInstallPath,
		Transitions: map[wizard.Action]wizard.State{
			wizard.ActionNext:   StateSummary,
			wizard.ActionBack:   StateComponents,
			wizard.ActionCancel: StateCancelled,
		},
	})
	
	ic.addState(StateSummary, &wizard.StateConfig{
		Name:        "Installation Summary",
		Description: "Review installation settings",
		CanGoNext:   true,
		CanGoBack:   true,
		CanCancel:   true,
		Transitions: map[wizard.Action]wizard.State{
			wizard.ActionNext:   StateProgress,
			wizard.ActionBack:   StateInstallPath,
			wizard.ActionCancel: StateCancelled,
		},
	})
	
	ic.addState(StateProgress, &wizard.StateConfig{
		Name:        "Installing",
		Description: "Installation in progress",
		CanGoNext:   false,
		CanGoBack:   false,
		CanCancel:   false,
		Transitions: map[wizard.Action]wizard.State{
			wizard.ActionNext: StateComplete, // Automatic transition after install
		},
	})
	
	ic.addState(StateComplete, &wizard.StateConfig{
		Name:        "Installation Complete",
		Description: "Installation finished successfully",
		CanGoNext:   false,
		CanGoBack:   false,
		CanCancel:   false,
	})
	
	ic.addState(StateCancelled, &wizard.StateConfig{
		Name:        "Installation Cancelled",
		Description: "Installation was cancelled",
		CanGoNext:   false,
		CanGoBack:   false,
		CanCancel:   false,
	})
	
	// Add custom states and rebuild transitions
	ic.integrateCustomStates()

	// Set initial state
	ic.dfa.SetInitialState(StateWelcome)
	ic.dfa.AddFinalState(StateComplete)
	ic.dfa.AddFinalState(StateCancelled)
}

// Helper methods for conditional next states
func (ic *InstallerController) getNextStateAfterWelcome() wizard.State {
	if ic.config.License != "" {
		return StateLicense
	}
	return StateComponents
}

func (ic *InstallerController) getPrevStateBeforeComponents() wizard.State {
	if ic.config.License != "" {
		return StateLicense
	}
	return StateWelcome
}

// integrateCustomStates adds custom states to the DFA and rebuilds transitions
func (ic *InstallerController) integrateCustomStates() {
	customHandlers := ic.customStates.GetAll()
	if len(customHandlers) == 0 {
		return
	}

	// Group custom states by insertion point
	insertionGroups := make(map[wizard.State][]CustomStateHandler)
	for _, handler := range customHandlers {
		insertPoint := handler.GetInsertionPoint()
		insertionGroups[insertPoint.After] = append(insertionGroups[insertPoint.After], handler)
	}

	// Add custom states to DFA
	for _, handler := range customHandlers {
		config := handler.GetConfig()

		// Wrap the validation function to include controller context
		if config.ValidateFunc != nil {
			originalValidate := config.ValidateFunc
			config.ValidateFunc = func(data map[string]interface{}) error {
				return originalValidate(data)
			}
		} else {
			// Use the handler's validation method - merge with persistent state data
			config.ValidateFunc = func(data map[string]interface{}) error {
				// Merge global state data with current data (same as HandleEnter/HandleLeave)
				mergedData := make(map[string]interface{})
				for k, v := range ic.stateData {
					mergedData[k] = v
				}
				for k, v := range data {
					mergedData[k] = v
				}
				return handler.Validate(ic, mergedData)
			}
		}

		ic.addState(handler.GetStateID(), config)
	}

	// Rebuild transitions to include custom states
	ic.rebuildTransitions(insertionGroups)
}

// rebuildTransitions updates state transitions to include custom states
func (ic *InstallerController) rebuildTransitions(insertionGroups map[wizard.State][]CustomStateHandler) {
	// Standard flow: Welcome -> [License] -> Components -> InstallPath -> Summary -> Progress -> Complete

	// Update transitions to chain custom states
	for afterState, handlers := range insertionGroups {
		if len(handlers) == 0 {
			continue
		}

		// Get the original next state for the insertion point
		originalNext := ic.getOriginalNextState(afterState)

		// Chain custom states together
		prevState := afterState
		for i, handler := range handlers {
			stateID := handler.GetStateID()

			// Update previous state to point to this custom state
			ic.updateStateTransition(prevState, wizard.ActionNext, stateID)

			// Set up transitions for this custom state
			var nextState wizard.State
			if i == len(handlers)-1 {
				// Last custom state points to original next
				nextState = originalNext
			} else {
				// Point to next custom state
				nextState = handlers[i+1].GetStateID()
			}

			// Set up back transition
			backState := ic.getOriginalBackState(afterState, stateID)

			ic.updateCustomStateTransitions(stateID, nextState, backState)
			prevState = stateID
		}
	}
}

// getOriginalNextState returns what the next state should be for a given state
func (ic *InstallerController) getOriginalNextState(state wizard.State) wizard.State {
	switch state {
	case StateWelcome:
		return ic.getNextStateAfterWelcome()
	case StateLicense:
		return StateComponents
	case StateComponents:
		return StateInstallPath
	case StateInstallPath:
		return StateSummary
	case StateSummary:
		return StateProgress
	default:
		return StateCancelled // Fallback
	}
}

// getOriginalBackState returns what the back state should be for a custom state
func (ic *InstallerController) getOriginalBackState(afterState wizard.State, customState wizard.State) wizard.State {
	// For now, custom states go back to the state they were inserted after
	return afterState
}

// updateStateTransition updates a state's transition for a specific action
func (ic *InstallerController) updateStateTransition(state wizard.State, action wizard.Action, newTarget wizard.State) {
	if stateConfig, err := ic.dfa.GetStateConfig(state); err == nil {
		if stateConfig.Transitions == nil {
			stateConfig.Transitions = make(map[wizard.Action]wizard.State)
		}
		stateConfig.Transitions[action] = newTarget
	}
}

// updateCustomStateTransitions sets up transitions for a custom state
func (ic *InstallerController) updateCustomStateTransitions(state wizard.State, nextState wizard.State, backState wizard.State) {
	if stateConfig, err := ic.dfa.GetStateConfig(state); err == nil {
		if stateConfig.Transitions == nil {
			stateConfig.Transitions = make(map[wizard.Action]wizard.State)
		}
		stateConfig.Transitions[wizard.ActionNext] = nextState
		stateConfig.Transitions[wizard.ActionBack] = backState
		stateConfig.Transitions[wizard.ActionCancel] = StateCancelled
	}
}

// addState adds a state to the DFA
func (ic *InstallerController) addState(state wizard.State, config *wizard.StateConfig) {
	if err := ic.dfa.AddState(state, config); err != nil {
		panic(fmt.Sprintf("Failed to add state %s: %v", state, err))
	}
}

// Validation functions
func (ic *InstallerController) validateLicense(data map[string]interface{}) error {
	if accepted, ok := data["license_accepted"].(bool); !ok || !accepted {
		return fmt.Errorf("license must be accepted to continue")
	}
	return nil
}

func (ic *InstallerController) validateComponents(data map[string]interface{}) error {
	if components, ok := data["selected_components"].([]core.Component); ok {
		// Ensure at least one required component is selected
		for _, comp := range components {
			if comp.Required && comp.Selected {
				return nil
			}
		}
		return fmt.Errorf("at least one required component must be selected")
	}
	return fmt.Errorf("no components selected")
}

func (ic *InstallerController) validateInstallPath(data map[string]interface{}) error {
	if path, ok := data["install_path"].(string); !ok || path == "" {
		return fmt.Errorf("installation path cannot be empty")
	}
	return nil
}

// State enter handlers
func (ic *InstallerController) handleStateEnter(state wizard.State, data map[string]interface{}) error {
	if ic.view == nil {
		return fmt.Errorf("no view set")
	}
	
	switch state {
	case StateWelcome:
		return ic.view.ShowWelcome()
		
	case StateLicense:
		accepted, err := ic.view.ShowLicense(ic.config.License)
		if err != nil {
			return err
		}
		data["license_accepted"] = accepted
		return nil
		
	case StateComponents:
		selected, err := ic.view.ShowComponents(ic.config.Components)
		if err != nil {
			return err
		}
		data["selected_components"] = selected
		// Update installer with selected components
		ic.installer.SetSelectedComponents(selected)
		return nil
		
	case StateInstallPath:
		defaultPath := ic.config.InstallDir
		if defaultPath == "" {
			// Determine default path based on platform
			defaultPath = "/opt/" + ic.config.AppName // Simplified
		}
		
		path, err := ic.view.ShowInstallPath(defaultPath)
		if err != nil {
			return err
		}
		data["install_path"] = path
		// Update installer with selected path
		ic.installer.SetInstallPath(path)
		return nil
		
	case StateSummary:
		components, _ := data["selected_components"].([]core.Component)
		installPath, _ := data["install_path"].(string)
		
		proceed, err := ic.view.ShowSummary(ic.config, components, installPath)
		if err != nil {
			return err
		}
		if !proceed {
			return fmt.Errorf("installation cancelled by user")
		}
		return nil
		
	case StateProgress:
		// Start installation in background
		go func() {
			ic.installer.SetUI(&controllerUIAdapter{controller: ic})
			err := ic.installer.ExecuteInstallation()
			if err != nil {
				ic.view.ShowErrorMessage(err)
				return
			}
			// Auto-transition to complete when done
			ic.dfa.Next()
		}()
		return nil
		
	case StateComplete:
		summary := ic.installer.CreateSummary()
		return ic.view.ShowComplete(summary)

	default:
		// Check if this is a custom state
		if handler, exists := ic.customStates.GetHandler(state); exists {
			// Merge global state data with current data
			mergedData := make(map[string]interface{})
			for k, v := range ic.stateData {
				mergedData[k] = v
			}
			for k, v := range data {
				mergedData[k] = v
			}

			// Call custom state handler
			if err := handler.HandleEnter(ic, mergedData); err != nil {
				return err
			}

			// Update global state data with results
			for k, v := range mergedData {
				ic.stateData[k] = v
			}

			return nil
		}

		return fmt.Errorf("unknown state: %s", state)
	}
}

// State leave handlers
func (ic *InstallerController) handleStateLeave(state wizard.State, data map[string]interface{}) error {
	// Handle custom states
	if handler, exists := ic.customStates.GetHandler(state); exists {
		// Merge global state data with current data
		mergedData := make(map[string]interface{})
		for k, v := range ic.stateData {
			mergedData[k] = v
		}
		for k, v := range data {
			mergedData[k] = v
		}

		// Call custom state leave handler
		if err := handler.HandleLeave(ic, mergedData); err != nil {
			return err
		}

		// Update global state data
		for k, v := range mergedData {
			ic.stateData[k] = v
		}
	}

	// Standard cleanup logic if needed
	return nil
}

// Public methods for UI to call
func (ic *InstallerController) Start() error {
	return ic.dfa.Start()
}

func (ic *InstallerController) Next() error {
	return ic.dfa.Next()
}

func (ic *InstallerController) Back() error {
	return ic.dfa.Back()
}

func (ic *InstallerController) Cancel() error {
	return ic.dfa.Cancel()
}

func (ic *InstallerController) GetCurrentState() wizard.State {
	return ic.dfa.CurrentState()
}

func (ic *InstallerController) CanGoNext() bool {
	return ic.dfa.CanTransition(wizard.ActionNext)
}

func (ic *InstallerController) CanGoBack() bool {
	return ic.dfa.CanTransition(wizard.ActionBack)
}

func (ic *InstallerController) CanCancel() bool {
	return ic.dfa.CanTransition(wizard.ActionCancel)
}

// controllerUIAdapter adapts the controller to the core.UI interface for installer progress
type controllerUIAdapter struct {
	controller *InstallerController
}

func (a *controllerUIAdapter) Initialize(ctx *core.Context) error { return nil }
func (a *controllerUIAdapter) Run() error { return nil }
func (a *controllerUIAdapter) Shutdown() error { return nil }
func (a *controllerUIAdapter) ShowWelcome() error { return nil }
func (a *controllerUIAdapter) ShowLicense(license string) (bool, error) { return true, nil }
func (a *controllerUIAdapter) SelectComponents(components []core.Component) ([]core.Component, error) {
	return components, nil
}
func (a *controllerUIAdapter) SelectInstallPath(defaultPath string) (string, error) {
	return defaultPath, nil
}
func (a *controllerUIAdapter) ShowProgress(progress *core.Progress) error {
	if a.controller.view != nil {
		return a.controller.view.ShowProgress(progress)
	}
	return nil
}
func (a *controllerUIAdapter) ShowError(err error, canRetry bool) (bool, error) {
	if a.controller.view != nil {
		return false, a.controller.view.ShowErrorMessage(err)
	}
	return false, err
}
func (a *controllerUIAdapter) ShowSuccess(summary *core.InstallSummary) error {
	// This is handled by the StateComplete transition
	return nil
}
func (a *controllerUIAdapter) RequestElevation(reason string) (bool, error) {
	return true, nil // Simplified
}
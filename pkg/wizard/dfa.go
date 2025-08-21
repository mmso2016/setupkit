// Package wizard provides a deterministic finite automaton (DFA) for managing
// wizard-like workflows with states, transitions, and validation.
package wizard

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents a state in the wizard
type State string

// Common states
const (
	StateInitial State = "initial"
	StateError   State = "error"
	StateFinal   State = "final"
)

// Action represents an action that can trigger a transition
type Action string

// Common actions
const (
	ActionNext     Action = "next"
	ActionBack     Action = "back"
	ActionCancel   Action = "cancel"
	ActionSkip     Action = "skip"
	ActionRetry    Action = "retry"
	ActionValidate Action = "validate"
	ActionSave     Action = "save"
)

// TransitionRule defines when and how a transition can occur
type TransitionRule struct {
	From      State
	To        State
	Action    Action
	Condition func(data map[string]interface{}) bool
	Priority  int // Higher priority rules are evaluated first
}

// Callbacks defines all callback functions for the DFA
type Callbacks struct {
	OnEnter           func(state State, data map[string]interface{}) error
	OnLeave           func(state State, data map[string]interface{}) error
	OnTransition      func(from, to State, action Action) error
	OnDataChange      func(state State, key string, oldValue, newValue interface{}) error
	OnValidationError func(state State, err error)
	OnCancel          func(state State, data map[string]interface{})
	BeforeTransition  func(from, to State, action Action) error
	AfterTransition   func(from, to State, action Action) error
}

// StateConfig defines the configuration for a state
type StateConfig struct {
	Name        string
	Description string

	// Capabilities
	CanGoBack bool
	CanGoNext bool
	CanCancel bool
	CanSkip   bool

	// Validation
	ValidateFunc    func(data map[string]interface{}) error
	ValidateOnEntry func(data map[string]interface{}) error
	ValidateOnExit  func(data map[string]interface{}) error

	// Entry control
	CanEnterFunc func(data map[string]interface{}) bool

	// Callbacks
	OnEnter      func(data map[string]interface{}) error
	OnExit       func(data map[string]interface{}) error
	OnDataChange func(key string, oldValue, newValue interface{}) error

	// Dynamic next state determination
	NextStateFunc func(data map[string]interface{}) (State, error)

	// Static transitions for this state
	Transitions map[Action]State
}

// DFA represents the deterministic finite automaton
type DFA struct {
	// Synchronization
	mu sync.RWMutex

	// State management
	states      map[State]*StateConfig
	current     State
	initial     State
	finalStates map[State]bool

	// History for back navigation
	history []State
	future  []State // For redo functionality

	// Data store
	data map[string]interface{}

	// Callbacks
	callbacks *Callbacks

	// Global validation
	GlobalValidator func(state State, data map[string]interface{}) error

	// Transition rules (global)
	transitions []TransitionRule

	// Options
	maxHistory     int
	AllowBackToAny bool // Allow going back to any previous state
	strictMode     bool // Enforce all validations
	DryRun         bool // Dry-run mode for testing

	// Dry-run tracking
	dryRunLog []string
}

// New creates a new DFA instance
func New() *DFA {
	return &DFA{
		states:      make(map[State]*StateConfig),
		finalStates: make(map[State]bool),
		history:     make([]State, 0),
		future:      make([]State, 0),
		data:        make(map[string]interface{}),
		transitions: make([]TransitionRule, 0),
		maxHistory:  100,
		strictMode:  true,
		dryRunLog:   make([]string, 0),
	}
}

// SetMaxHistory sets the maximum history size
func (d *DFA) SetMaxHistory(max int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maxHistory = max
}

// SetCallbacks sets the callbacks for the DFA
func (d *DFA) SetCallbacks(callbacks *Callbacks) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.callbacks = callbacks
}

// SetStrictMode enables or disables strict validation mode
func (d *DFA) SetStrictMode(strict bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.strictMode = strict
}

// SetDryRun enables or disables dry-run mode
func (d *DFA) SetDryRun(dryRun bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.DryRun = dryRun
	if dryRun {
		d.dryRunLog = make([]string, 0)
	}
}

// GetDryRunLog returns the dry-run log
func (d *DFA) GetDryRunLog() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return append([]string{}, d.dryRunLog...)
}

// logDryRun logs a dry-run action
func (d *DFA) logDryRun(format string, args ...interface{}) {
	if d.DryRun {
		msg := fmt.Sprintf(format, args...)
		d.dryRunLog = append(d.dryRunLog, msg)
	}
}

// AddState adds a new state to the DFA
func (d *DFA) AddState(state State, config *StateConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.states[state]; exists {
		return fmt.Errorf("state %s already exists", state)
	}

	if config == nil {
		config = &StateConfig{}
	}

	if config.Name == "" {
		config.Name = string(state)
	}

	if config.Transitions == nil {
		config.Transitions = make(map[Action]State)
	}

	d.states[state] = config

	// First added state becomes initial state if not set
	if d.initial == "" {
		d.initial = state
		d.current = state
		d.history = []State{state}
		d.logDryRun("Initial state set to: %s", state)
	}

	d.logDryRun("Added state: %s", state)
	return nil
}

// SetInitialState sets the initial state
func (d *DFA) SetInitialState(state State) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.states[state]; !exists {
		return fmt.Errorf("state %s does not exist", state)
	}

	d.initial = state
	d.current = state
	d.history = []State{state}
	d.logDryRun("Initial state changed to: %s", state)
	return nil
}

// GetStateConfig retrieves the configuration for a state
func (d *DFA) GetStateConfig(state State) (*StateConfig, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	config, exists := d.states[state]
	if !exists {
		return nil, fmt.Errorf("state %s does not exist", state)
	}

	return config, nil
}

// CurrentState returns the current state
func (d *DFA) CurrentState() State {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.current
}

// GetData returns a value from the data store
func (d *DFA) GetData(key string) (interface{}, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	value, exists := d.data[key]
	return value, exists
}

// GetAllData returns a copy of all data
func (d *DFA) GetAllData() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range d.data {
		result[k] = v
	}
	return result
}

// SetData sets a value in the data store
func (d *DFA) SetData(key string, value interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	oldValue := d.data[key]
	d.data[key] = value

	d.logDryRun("SetData: %s = %v (old: %v)", key, value, oldValue)

	// Skip callbacks in dry-run mode
	if d.DryRun {
		return nil
	}

	// Trigger data change callbacks
	if config, exists := d.states[d.current]; exists && config.OnDataChange != nil {
		if err := config.OnDataChange(key, oldValue, value); err != nil {
			// Rollback on error
			d.data[key] = oldValue
			return err
		}
	}

	if d.callbacks != nil && d.callbacks.OnDataChange != nil {
		if err := d.callbacks.OnDataChange(d.current, key, oldValue, value); err != nil {
			// Rollback on error
			d.data[key] = oldValue
			return err
		}
	}

	return nil
}

// Start initializes the DFA and moves to the initial state
func (d *DFA) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.initial == "" {
		return errors.New("no initial state defined")
	}

	d.current = ""
	d.history = []State{}
	d.future = []State{}

	d.logDryRun("Starting DFA from initial state: %s", d.initial)
	return d.transitionToInternal(d.initial, ActionNext)
}

// Next moves to the next state
func (d *DFA) Next() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logDryRun("Attempting Next from state: %s", d.current)

	config, exists := d.states[d.current]
	if !exists {
		return fmt.Errorf("current state %s does not exist", d.current)
	}

	if !config.CanGoNext {
		return errors.New("cannot go to next state from current state")
	}

	// Validate current state first
	if config.ValidateFunc != nil && d.strictMode && !d.DryRun {
		if err := config.ValidateFunc(d.data); err != nil {
			if d.callbacks != nil && d.callbacks.OnValidationError != nil {
				d.callbacks.OnValidationError(d.current, err)
			}
			return err
		}
	}

	// Determine next state
	var nextState State
	var err error

	if config.NextStateFunc != nil {
		nextState, err = config.NextStateFunc(d.data)
		if err != nil {
			return err
		}
	} else if next, ok := config.Transitions[ActionNext]; ok {
		nextState = next
	} else {
		// Look for global transition rule
		for _, rule := range d.transitions {
			if rule.From == d.current && rule.Action == ActionNext {
				if rule.Condition == nil || rule.Condition(d.data) {
					nextState = rule.To
					break
				}
			}
		}
	}

	if nextState == "" {
		return errors.New("no next state defined")
	}

	return d.transitionToInternal(nextState, ActionNext)
}

// Back moves to the previous state
func (d *DFA) Back() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logDryRun("Attempting Back from state: %s", d.current)

	if len(d.history) <= 1 {
		return errors.New("no previous state in history")
	}

	config, exists := d.states[d.current]
	if !exists {
		return fmt.Errorf("current state %s does not exist", d.current)
	}

	if !config.CanGoBack {
		return errors.New("cannot go back from current state")
	}

	// Get previous state (skip current in history)
	prevState := d.history[len(d.history)-2]

	// Add current state to future for redo
	d.future = append(d.future, d.current)

	// Remove current from history
	d.history = d.history[:len(d.history)-1]

	return d.transitionToInternal(prevState, ActionBack)
}

// Skip skips the current state
func (d *DFA) Skip() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logDryRun("Attempting Skip from state: %s", d.current)

	config, exists := d.states[d.current]
	if !exists {
		return fmt.Errorf("current state %s does not exist", d.current)
	}

	if !config.CanSkip {
		return errors.New("cannot skip current state")
	}

	// Determine skip target
	var skipTo State
	if next, ok := config.Transitions[ActionSkip]; ok {
		skipTo = next
	} else if next, ok := config.Transitions[ActionNext]; ok {
		skipTo = next
	} else {
		return errors.New("no skip target defined")
	}

	// Skip bypasses validation
	return d.transitionToInternal(skipTo, ActionSkip)
}

// Cancel cancels the wizard
func (d *DFA) Cancel() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logDryRun("Attempting Cancel from state: %s", d.current)

	config, exists := d.states[d.current]
	if !exists {
		return fmt.Errorf("current state %s does not exist", d.current)
	}

	if !config.CanCancel {
		return errors.New("cannot cancel from current state")
	}

	if d.callbacks != nil && d.callbacks.OnCancel != nil && !d.DryRun {
		// Create a copy of data for the callback
		dataCopy := make(map[string]interface{})
		for k, v := range d.data {
			dataCopy[k] = v
		}
		d.callbacks.OnCancel(d.current, dataCopy)
	}

	return nil
}

// Transition performs a custom action transition
func (d *DFA) Transition(action Action) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logDryRun("Attempting transition with action %s from state: %s", action, d.current)

	config, exists := d.states[d.current]
	if !exists {
		return fmt.Errorf("current state %s does not exist", d.current)
	}

	// Check state-specific transitions
	if next, ok := config.Transitions[action]; ok {
		return d.transitionToInternal(next, action)
	}

	// Check global transitions
	for _, rule := range d.transitions {
		if rule.From == d.current && rule.Action == action {
			if rule.Condition == nil || rule.Condition(d.data) {
				return d.transitionToInternal(rule.To, action)
			}
		}
	}

	return fmt.Errorf("no transition defined for action %s from state %s", action, d.current)
}

// CanTransition checks if an action is possible from current state
func (d *DFA) CanTransition(action Action) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.canTransitionInternal(action)
}

// canTransitionInternal checks if an action is possible (internal, assumes lock held)
func (d *DFA) canTransitionInternal(action Action) bool {
	config, exists := d.states[d.current]
	if !exists {
		return false
	}

	switch action {
	case ActionNext:
		return config.CanGoNext
	case ActionBack:
		return config.CanGoBack && len(d.history) > 1
	case ActionSkip:
		return config.CanSkip
	case ActionCancel:
		return config.CanCancel
	default:
		// Check if custom action is defined
		if _, ok := config.Transitions[action]; ok {
			return true
		}
		// Check global transitions
		for _, rule := range d.transitions {
			if rule.From == d.current && rule.Action == action {
				if rule.Condition == nil || rule.Condition(d.data) {
					return true
				}
			}
		}
	}

	return false
}

// ValidateCurrentState validates the current state
func (d *DFA) ValidateCurrentState() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Skip validation in dry-run mode
	if d.DryRun {
		return nil
	}

	config, exists := d.states[d.current]
	if !exists {
		return fmt.Errorf("current state %s does not exist", d.current)
	}

	if config.ValidateFunc != nil {
		return config.ValidateFunc(d.data)
	}

	if config.ValidateOnExit != nil {
		return config.ValidateOnExit(d.data)
	}

	if d.GlobalValidator != nil {
		return d.GlobalValidator(d.current, d.data)
	}

	return nil
}

// transitionToInternal performs the actual transition (internal, assumes lock held)
func (d *DFA) transitionToInternal(to State, action Action) error {
	from := d.current

	d.logDryRun("Transition: %s -> %s (action: %s)", from, to, action)

	// Check if target state exists
	toConfig, exists := d.states[to]
	if !exists {
		return fmt.Errorf("target state %s does not exist", to)
	}

	// Check if can enter target state
	if toConfig.CanEnterFunc != nil && !toConfig.CanEnterFunc(d.data) && !d.DryRun {
		return fmt.Errorf("cannot enter state %s", to)
	}

	// Before transition callback
	if d.callbacks != nil && d.callbacks.BeforeTransition != nil && !d.DryRun {
		if err := d.callbacks.BeforeTransition(from, to, action); err != nil {
			return err
		}
	}

	// Exit current state
	if from != "" {
		if config, exists := d.states[from]; exists {
			// Validate on exit
			if config.ValidateOnExit != nil && d.strictMode && !d.DryRun {
				if err := config.ValidateOnExit(d.data); err != nil {
					return err
				}
			}

			// OnLeave callback
			if d.callbacks != nil && d.callbacks.OnLeave != nil && !d.DryRun {
				if err := d.callbacks.OnLeave(from, d.data); err != nil {
					return err
				}
			}

			// State OnExit
			if config.OnExit != nil && !d.DryRun {
				if err := config.OnExit(d.data); err != nil {
					return err
				}
			}
		}
	}

	// OnTransition callback
	if d.callbacks != nil && d.callbacks.OnTransition != nil && !d.DryRun {
		if err := d.callbacks.OnTransition(from, to, action); err != nil {
			return err
		}
	}

	// Update state
	oldCurrent := d.current
	d.current = to

	// Update history (don't add duplicates for back action)
	if action != ActionBack {
		d.history = append(d.history, to)
		// Trim history if needed
		if d.maxHistory > 0 && len(d.history) > d.maxHistory {
			d.history = d.history[len(d.history)-d.maxHistory:]
		}
		// Clear future on forward navigation
		d.future = []State{}
	}

	// Enter new state
	// Validate on entry
	if toConfig.ValidateOnEntry != nil && d.strictMode && !d.DryRun {
		if err := toConfig.ValidateOnEntry(d.data); err != nil {
			// Rollback
			d.current = oldCurrent
			if action != ActionBack {
				d.history = d.history[:len(d.history)-1]
			}
			return err
		}
	}

	// OnEnter callback
	if d.callbacks != nil && d.callbacks.OnEnter != nil && !d.DryRun {
		if err := d.callbacks.OnEnter(to, d.data); err != nil {
			// Rollback
			d.current = oldCurrent
			if action != ActionBack {
				d.history = d.history[:len(d.history)-1]
			}
			return err
		}
	}

	// State OnEnter
	if toConfig.OnEnter != nil && !d.DryRun {
		if err := toConfig.OnEnter(d.data); err != nil {
			// Rollback
			d.current = oldCurrent
			if action != ActionBack {
				d.history = d.history[:len(d.history)-1]
			}
			return err
		}
	}

	// Check if we've entered a final state and set completion time
	if d.finalStates[to] {
		if _, exists := d.data["completed_at"]; !exists {
			if d.DryRun {
				// In dry-run mode, use a fixed timestamp for consistency
				d.data["completed_at"] = "2024-01-01 12:00:00"
			} else {
				d.data["completed_at"] = time.Now().Format(time.RFC3339)
			}
		}
	}

	// After transition callback
	if d.callbacks != nil && d.callbacks.AfterTransition != nil && !d.DryRun {
		d.callbacks.AfterTransition(from, to, action)
	}

	return nil
}

// GetHistory returns the state history
func (d *DFA) GetHistory() []State {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return append([]State{}, d.history...)
}

// Reset resets the DFA to initial state
func (d *DFA) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.current = d.initial
	d.history = []State{d.initial}
	d.future = []State{}
	d.data = make(map[string]interface{})
	d.dryRunLog = []string{}

	d.logDryRun("DFA reset to initial state: %s", d.initial)
}

// IsInFinalState checks if current state is final
func (d *DFA) IsInFinalState() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.finalStates[d.current]
}

// AddFinalState marks a state as final
func (d *DFA) AddFinalState(state State) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.states[state]; !exists {
		return fmt.Errorf("state %s does not exist", state)
	}
	d.finalStates[state] = true

	d.logDryRun("State %s marked as final", state)
	return nil
}

// AddTransition adds a global transition rule
func (d *DFA) AddTransition(rule TransitionRule) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.states[rule.From]; !exists {
		return fmt.Errorf("from state %s does not exist", rule.From)
	}
	if _, exists := d.states[rule.To]; !exists {
		return fmt.Errorf("to state %s does not exist", rule.To)
	}

	d.transitions = append(d.transitions, rule)
	d.logDryRun("Added transition: %s -> %s (action: %s)", rule.From, rule.To, rule.Action)
	return nil
}

// GetAvailableActions returns available actions for current state
func (d *DFA) GetAvailableActions() []Action {
	d.mu.RLock()
	defer d.mu.RUnlock()

	actions := []Action{}

	if d.canTransitionInternal(ActionNext) {
		actions = append(actions, ActionNext)
	}
	if d.canTransitionInternal(ActionBack) {
		actions = append(actions, ActionBack)
	}
	if d.canTransitionInternal(ActionSkip) {
		actions = append(actions, ActionSkip)
	}
	if d.canTransitionInternal(ActionCancel) {
		actions = append(actions, ActionCancel)
	}

	// Add custom actions
	if config, exists := d.states[d.current]; exists {
		for action := range config.Transitions {
			if action != ActionNext && action != ActionBack &&
				action != ActionSkip && action != ActionCancel {
				actions = append(actions, action)
			}
		}
	}

	return actions
}

// Validate performs a complete validation of the DFA configuration
func (d *DFA) Validate() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.states) == 0 {
		return errors.New("no states defined")
	}

	if d.initial == "" {
		return errors.New("no initial state defined")
	}

	if _, exists := d.states[d.initial]; !exists {
		return fmt.Errorf("initial state %s does not exist", d.initial)
	}

	// Check for unreachable states (optional, could be expensive)
	// Check for invalid transitions
	for _, rule := range d.transitions {
		if _, exists := d.states[rule.From]; !exists {
			return fmt.Errorf("transition rule references non-existent from state: %s", rule.From)
		}
		if _, exists := d.states[rule.To]; !exists {
			return fmt.Errorf("transition rule references non-existent to state: %s", rule.To)
		}
	}

	// Check state transitions
	for state, config := range d.states {
		for _, target := range config.Transitions {
			if _, exists := d.states[target]; !exists {
				return fmt.Errorf("state %s has transition to non-existent state: %s", state, target)
			}
		}
	}

	return nil
}

// Clone creates a deep copy of the DFA
func (d *DFA) Clone() *DFA {
	d.mu.RLock()
	defer d.mu.RUnlock()

	clone := New()

	// Copy states
	for state, config := range d.states {
		configCopy := *config
		if config.Transitions != nil {
			configCopy.Transitions = make(map[Action]State)
			for k, v := range config.Transitions {
				configCopy.Transitions[k] = v
			}
		}
		clone.states[state] = &configCopy
	}

	// Copy other fields
	clone.current = d.current
	clone.initial = d.initial
	clone.history = append([]State{}, d.history...)
	clone.future = append([]State{}, d.future...)

	// Copy data
	for k, v := range d.data {
		clone.data[k] = v
	}

	// Copy final states
	for state := range d.finalStates {
		clone.finalStates[state] = true
	}

	// Copy transitions
	clone.transitions = append([]TransitionRule{}, d.transitions...)

	// Copy options
	clone.maxHistory = d.maxHistory
	clone.AllowBackToAny = d.AllowBackToAny
	clone.strictMode = d.strictMode
	clone.DryRun = d.DryRun

	return clone
}

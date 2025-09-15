// Package wizard provides a deterministic finite automaton (DFA) for managing
// wizard-like workflows with states, transitions, and validation.
package wizard

import (
	"errors"
	"fmt"
	"strings"
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

// MainState represents a top-level screen state  
type MainState string

// SubState represents a sub-state within a main state
type SubState string

// CompositeState represents the current hierarchical state
type CompositeState struct {
	Main MainState `json:"main"`
	Sub  SubState  `json:"sub"`
}

// String returns string representation of composite state
func (cs CompositeState) String() string {
	if cs.Sub == "" {
		return string(cs.Main)
	}
	return fmt.Sprintf("%s.%s", cs.Main, cs.Sub)
}

// ToState converts CompositeState to legacy State for compatibility
func (cs CompositeState) ToState() State {
	return State(cs.String())
}

// ParseCompositeState parses a string into CompositeState
func ParseCompositeState(s string) CompositeState {
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		return CompositeState{Main: MainState(parts[0]), Sub: ""}
	}
	return CompositeState{Main: MainState(parts[0]), Sub: SubState(parts[1])}
}

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

// SubAction represents actions within a sub-state
type SubAction string

// Common sub-actions  
const (
	SubActionSelect   SubAction = "select"
	SubActionDeselect SubAction = "deselect"
	SubActionInput    SubAction = "input"
	SubActionFocus    SubAction = "focus"
	SubActionBlur     SubAction = "blur"
	SubActionScroll   SubAction = "scroll"
	SubActionResize   SubAction = "resize"
	SubActionRefresh  SubAction = "refresh"
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

// SubStateConfig defines configuration for a sub-state
type SubStateConfig struct {
	Name        string
	Description string
	
	// Allowed sub-actions  
	AllowedActions map[SubAction]bool
	
	// Validation function for sub-state
	ValidateFunc func(data map[string]interface{}) error
	
	// Callbacks for sub-state
	OnEnter func(data map[string]interface{}) error
	OnExit  func(data map[string]interface{}) error
	
	// Auto-transition conditions
	AutoTransitionTo   SubState
	AutoTransitionFunc func(data map[string]interface{}) SubState
	
	// Completion condition - when this returns true, sub-state can exit
	CanComplete func(data map[string]interface{}) bool
}

// MainStateConfig defines configuration for a main state with sub-states
type MainStateConfig struct {
	*StateConfig // Embed original config
	
	// Sub-states for this main state
	SubStates map[SubState]*SubStateConfig
	
	// Initial sub-state when entering this main state
	InitialSubState SubState
	
	// Whether sub-states are required or optional
	RequireSubStateCompletion bool
}

// HierarchicalDFA represents a two-level DFA system
type HierarchicalDFA struct {
	// Synchronization
	mu sync.RWMutex
	
	// Main state management (level 1 - screens)
	mainStates map[MainState]*MainStateConfig
	
	// Current hierarchical state
	currentState CompositeState
	initial      CompositeState
	finalStates  map[MainState]bool
	
	// History for hierarchical navigation
	history []CompositeState
	future  []CompositeState
	
	// Shared data store
	data map[string]interface{}
	
	// Callbacks
	callbacks *Callbacks
	
	// Global validation
	GlobalValidator func(state CompositeState, data map[string]interface{}) error
	
	// Options
	maxHistory     int
	AllowBackToAny bool
	strictMode     bool
	DryRun         bool
	
	// Dry-run tracking
	dryRunLog []string
}

// DFA represents the legacy single-level deterministic finite automaton
// Kept for backward compatibility where needed
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

// NewHierarchical creates a new hierarchical DFA instance
func NewHierarchical() *HierarchicalDFA {
	return &HierarchicalDFA{
		mainStates:  make(map[MainState]*MainStateConfig),
		finalStates: make(map[MainState]bool),
		history:     make([]CompositeState, 0),
		future:      make([]CompositeState, 0),
		data:        make(map[string]interface{}),
		maxHistory:  100,
		strictMode:  true,
		dryRunLog:   make([]string, 0),
	}
}

// New creates a new DFA instance (legacy)
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

// =============================================================================
// HIERARCHICAL DFA METHODS
// =============================================================================

// AddMainState adds a main state with optional sub-states to the hierarchical DFA
func (h *HierarchicalDFA) AddMainState(mainState MainState, config *MainStateConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if config == nil {
		return errors.New("main state config cannot be nil")
	}
	
	// Check if main state already exists
	if _, exists := h.mainStates[mainState]; exists {
		return fmt.Errorf("main state %s already exists", mainState)
	}
	
	// Initialize sub-states map if nil
	if config.SubStates == nil {
		config.SubStates = make(map[SubState]*SubStateConfig)
	}
	
	// Set first main state as initial if none set
	if h.initial.Main == "" {
		h.initial = CompositeState{Main: mainState, Sub: config.InitialSubState}
		h.currentState = h.initial
	}
	
	h.mainStates[mainState] = config
	
	if h.DryRun {
		h.dryRunLog = append(h.dryRunLog, fmt.Sprintf("Added main state: %s", mainState))
	}
	
	return nil
}

// AddSubState adds a sub-state to an existing main state
func (h *HierarchicalDFA) AddSubState(mainState MainState, subState SubState, config *SubStateConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if config == nil {
		return errors.New("sub-state config cannot be nil")
	}
	
	// Check if main state exists
	mainConfig, exists := h.mainStates[mainState]
	if !exists {
		return fmt.Errorf("main state %s does not exist", mainState)
	}
	
	// Check if sub-state already exists
	if _, exists := mainConfig.SubStates[subState]; exists {
		return fmt.Errorf("sub-state %s.%s already exists", mainState, subState)
	}
	
	// Initialize allowed actions if nil
	if config.AllowedActions == nil {
		config.AllowedActions = make(map[SubAction]bool)
	}
	
	mainConfig.SubStates[subState] = config
	
	if h.DryRun {
		h.dryRunLog = append(h.dryRunLog, fmt.Sprintf("Added sub-state: %s.%s", mainState, subState))
	}
	
	return nil
}

// NavigateToMainState transitions to a different main state  
func (h *HierarchicalDFA) NavigateToMainState(mainState MainState) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	return h.navigateToMainState(mainState)
}

// navigateToMainState internal method without locking
func (h *HierarchicalDFA) navigateToMainState(mainState MainState) error {
	// Check if main state exists
	mainConfig, exists := h.mainStates[mainState]
	if !exists {
		return fmt.Errorf("main state %s does not exist", mainState)
	}
	
	oldState := h.currentState
	newState := CompositeState{Main: mainState, Sub: mainConfig.InitialSubState}
	
	// Validate transition if in strict mode
	if h.strictMode && h.GlobalValidator != nil {
		if err := h.GlobalValidator(newState, h.data); err != nil {
			return fmt.Errorf("global validation failed: %w", err)
		}
	}
	
	// Call exit callback for current state
	if oldState.Main != "" {
		if oldMainConfig, ok := h.mainStates[oldState.Main]; ok {
			if oldMainConfig.StateConfig != nil && oldMainConfig.StateConfig.OnExit != nil {
				if err := oldMainConfig.StateConfig.OnExit(h.data); err != nil && h.strictMode {
					return fmt.Errorf("exit callback failed: %w", err)
				}
			}
			
			// Call sub-state exit callback if in sub-state
			if oldState.Sub != "" {
				if subConfig, ok := oldMainConfig.SubStates[oldState.Sub]; ok && subConfig.OnExit != nil {
					if err := subConfig.OnExit(h.data); err != nil && h.strictMode {
						return fmt.Errorf("sub-state exit callback failed: %w", err)
					}
				}
			}
		}
	}
	
	// Add to history
	h.addToHistory(oldState)
	
	// Update current state
	h.currentState = newState
	
	// Call enter callback for new main state
	if mainConfig.StateConfig != nil && mainConfig.StateConfig.OnEnter != nil {
		if err := mainConfig.StateConfig.OnEnter(h.data); err != nil && h.strictMode {
			return fmt.Errorf("enter callback failed: %w", err)
		}
	}
	
	// Call enter callback for initial sub-state if it exists
	if newState.Sub != "" {
		if subConfig, ok := mainConfig.SubStates[newState.Sub]; ok && subConfig.OnEnter != nil {
			if err := subConfig.OnEnter(h.data); err != nil && h.strictMode {
				return fmt.Errorf("sub-state enter callback failed: %w", err)
			}
		}
	}
	
	// Call transition callback
	if h.callbacks != nil && h.callbacks.OnTransition != nil {
		action := Action("navigate") // Default action for main state transitions
		if err := h.callbacks.OnTransition(oldState.ToState(), newState.ToState(), action); err != nil && h.strictMode {
			return fmt.Errorf("transition callback failed: %w", err)
		}
	}
	
	if h.DryRun {
		h.dryRunLog = append(h.dryRunLog, fmt.Sprintf("Navigated: %s → %s", oldState, newState))
	}
	
	return nil
}

// NavigateToSubState transitions to a sub-state within the current main state
func (h *HierarchicalDFA) NavigateToSubState(subState SubState) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	return h.navigateToSubState(subState)
}

// navigateToSubState internal method without locking
func (h *HierarchicalDFA) navigateToSubState(subState SubState) error {
	currentMain := h.currentState.Main
	
	// Check if main state has this sub-state
	mainConfig, exists := h.mainStates[currentMain]
	if !exists {
		return fmt.Errorf("current main state %s does not exist", currentMain)
	}
	
	subConfig, exists := mainConfig.SubStates[subState]
	if !exists {
		return fmt.Errorf("sub-state %s.%s does not exist", currentMain, subState)
	}
	
	oldState := h.currentState
	newState := CompositeState{Main: currentMain, Sub: subState}
	
	// Validate sub-state transition
	if h.strictMode && subConfig.ValidateFunc != nil {
		if err := subConfig.ValidateFunc(h.data); err != nil {
			return fmt.Errorf("sub-state validation failed: %w", err)
		}
	}
	
	// Call exit callback for current sub-state
	if oldState.Sub != "" {
		if oldSubConfig, ok := mainConfig.SubStates[oldState.Sub]; ok && oldSubConfig.OnExit != nil {
			if err := oldSubConfig.OnExit(h.data); err != nil && h.strictMode {
				return fmt.Errorf("sub-state exit callback failed: %w", err)
			}
		}
	}
	
	// Update current state
	h.currentState = newState
	
	// Call enter callback for new sub-state
	if subConfig.OnEnter != nil {
		if err := subConfig.OnEnter(h.data); err != nil && h.strictMode {
			return fmt.Errorf("sub-state enter callback failed: %w", err)
		}
	}
	
	if h.DryRun {
		h.dryRunLog = append(h.dryRunLog, fmt.Sprintf("Sub-state transition: %s → %s", oldState, newState))
	}
	
	return nil
}

// HandleSubAction processes a sub-action within the current sub-state
func (h *HierarchicalDFA) HandleSubAction(action SubAction) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	currentMain := h.currentState.Main
	currentSub := h.currentState.Sub
	
	// Check if we're in a sub-state
	if currentSub == "" {
		return fmt.Errorf("not currently in a sub-state, cannot handle sub-action %s", action)
	}
	
	// Get main and sub state configs
	mainConfig, exists := h.mainStates[currentMain]
	if !exists {
		return fmt.Errorf("current main state %s does not exist", currentMain)
	}
	
	subConfig, exists := mainConfig.SubStates[currentSub]
	if !exists {
		return fmt.Errorf("current sub-state %s.%s does not exist", currentMain, currentSub)
	}
	
	// Check if action is allowed
	if len(subConfig.AllowedActions) > 0 {
		if !subConfig.AllowedActions[action] {
			return fmt.Errorf("action %s not allowed in sub-state %s.%s", action, currentMain, currentSub)
		}
	}
	
	// Check for auto-transition
	if subConfig.AutoTransitionFunc != nil {
		if nextSub := subConfig.AutoTransitionFunc(h.data); nextSub != "" {
			return h.navigateToSubState(nextSub)
		}
	} else if subConfig.AutoTransitionTo != "" {
		return h.navigateToSubState(subConfig.AutoTransitionTo)
	}
	
	if h.DryRun {
		h.dryRunLog = append(h.dryRunLog, fmt.Sprintf("Handled sub-action: %s in %s", action, h.currentState))
	}
	
	return nil
}

// CanCompleteCurrentSubState checks if current sub-state can be completed
func (h *HierarchicalDFA) CanCompleteCurrentSubState() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if h.currentState.Sub == "" {
		return true // Not in sub-state, always can complete
	}
	
	mainConfig, exists := h.mainStates[h.currentState.Main]
	if !exists {
		return false
	}
	
	subConfig, exists := mainConfig.SubStates[h.currentState.Sub]
	if !exists {
		return false
	}
	
	if subConfig.CanComplete != nil {
		return subConfig.CanComplete(h.data)
	}
	
	return true // No completion condition means always can complete
}

// GetCurrentState returns the current composite state
func (h *HierarchicalDFA) GetCurrentState() CompositeState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentState
}

// SetData sets a data value
func (h *HierarchicalDFA) SetData(key string, value interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	oldValue := h.data[key]
	h.data[key] = value
	
	// Call data change callback
	if h.callbacks != nil && h.callbacks.OnDataChange != nil {
		h.callbacks.OnDataChange(h.currentState.ToState(), key, oldValue, value)
	}
}

// GetData gets a data value
func (h *HierarchicalDFA) GetData(key string) (interface{}, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	value, exists := h.data[key]
	return value, exists
}

// GetAllData returns a copy of all data
func (h *HierarchicalDFA) GetAllData() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range h.data {
		result[k] = v
	}
	return result
}

// CanGoBack checks if we can go back in history
func (h *HierarchicalDFA) CanGoBack() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.history) > 0
}

// GoBack goes back to the previous state in history
func (h *HierarchicalDFA) GoBack() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if len(h.history) == 0 {
		return errors.New("no previous state in history")
	}
	
	// Get previous state
	prevState := h.history[len(h.history)-1]
	h.history = h.history[:len(h.history)-1]
	
	// Add current state to future
	h.addToFuture(h.currentState)
	
	// Navigate to previous state
	oldState := h.currentState
	h.currentState = prevState
	
	if h.DryRun {
		h.dryRunLog = append(h.dryRunLog, fmt.Sprintf("Went back: %s → %s", oldState, prevState))
	}
	
	return nil
}

// addToHistory adds a state to history
func (h *HierarchicalDFA) addToHistory(state CompositeState) {
	// Don't add if same as current
	if len(h.history) > 0 && h.history[len(h.history)-1] == state {
		return
	}
	
	h.history = append(h.history, state)
	
	// Trim history if too long
	if len(h.history) > h.maxHistory {
		h.history = h.history[1:]
	}
	
	// Clear future when adding new history
	h.future = h.future[:0]
}

// addToFuture adds a state to future (for redo)
func (h *HierarchicalDFA) addToFuture(state CompositeState) {
	h.future = append(h.future, state)
	
	// Trim future if too long
	if len(h.future) > h.maxHistory {
		h.future = h.future[1:]
	}
}

// IsFinalState checks if the current main state is final
func (h *HierarchicalDFA) IsFinalState() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.finalStates[h.currentState.Main]
}

// SetDryRun enables or disables dry-run mode
func (h *HierarchicalDFA) SetDryRun(dryRun bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.DryRun = dryRun
}

// GetDryRunLog returns the dry-run log
func (h *HierarchicalDFA) GetDryRunLog() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]string, len(h.dryRunLog))
	copy(result, h.dryRunLog)
	return result
}

// SetCallbacks sets the callbacks for the hierarchical DFA
func (h *HierarchicalDFA) SetCallbacks(callbacks *Callbacks) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.callbacks = callbacks
}

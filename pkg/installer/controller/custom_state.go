// Package controller provides custom state support for DFA installer
package controller

import (
	"fmt"

	"github.com/mmso2016/setupkit/pkg/wizard"
)

// CustomStateHandler defines the interface for custom installation states
type CustomStateHandler interface {
	// GetStateID returns the unique identifier for this state
	GetStateID() wizard.State

	// GetConfig returns the DFA state configuration
	GetConfig() *wizard.StateConfig

	// HandleEnter is called when entering this state
	HandleEnter(controller *InstallerController, data map[string]interface{}) error

	// HandleLeave is called when leaving this state
	HandleLeave(controller *InstallerController, data map[string]interface{}) error

	// Validate is called to validate state data before proceeding
	Validate(controller *InstallerController, data map[string]interface{}) error

	// GetInsertionPoint returns where in the flow this state should be inserted
	GetInsertionPoint() InsertionPoint
}

// InsertionPoint defines where a custom state should be inserted in the flow
type InsertionPoint struct {
	After  wizard.State // Insert after this state
	Before wizard.State // Insert before this state (optional)
}

// CustomStateData holds data for custom states
type CustomStateData map[string]interface{}

// Common insertion points
var (
	InsertAfterWelcome     = InsertionPoint{After: StateWelcome, Before: StateLicense}
	InsertAfterLicense     = InsertionPoint{After: StateLicense, Before: StateComponents}
	InsertAfterComponents  = InsertionPoint{After: StateComponents, Before: StateInstallPath}
	InsertAfterInstallPath = InsertionPoint{After: StateInstallPath, Before: StateSummary}
	InsertAfterSummary     = InsertionPoint{After: StateSummary, Before: StateProgress}
)

// CustomStateRegistry manages registered custom states
type CustomStateRegistry struct {
	handlers map[wizard.State]CustomStateHandler
	order    []wizard.State
}

// NewCustomStateRegistry creates a new registry
func NewCustomStateRegistry() *CustomStateRegistry {
	return &CustomStateRegistry{
		handlers: make(map[wizard.State]CustomStateHandler),
		order:    make([]wizard.State, 0),
	}
}

// Register adds a custom state handler to the registry
func (r *CustomStateRegistry) Register(handler CustomStateHandler) error {
	stateID := handler.GetStateID()
	if _, exists := r.handlers[stateID]; exists {
		return fmt.Errorf("custom state already registered: %s", stateID)
	}

	r.handlers[stateID] = handler
	r.order = append(r.order, stateID)
	return nil
}

// GetHandler retrieves a custom state handler by ID
func (r *CustomStateRegistry) GetHandler(stateID wizard.State) (CustomStateHandler, bool) {
	handler, exists := r.handlers[stateID]
	return handler, exists
}

// GetAll returns all registered handlers in registration order
func (r *CustomStateRegistry) GetAll() []CustomStateHandler {
	handlers := make([]CustomStateHandler, 0, len(r.order))
	for _, stateID := range r.order {
		if handler, exists := r.handlers[stateID]; exists {
			handlers = append(handlers, handler)
		}
	}
	return handlers
}

// ExtendedInstallerView extends the base InstallerView with custom state support
type ExtendedInstallerView interface {
	InstallerView

	// ShowCustomState handles display of custom states
	ShowCustomState(stateID wizard.State, data CustomStateData) (result CustomStateData, err error)
}

// BaseCustomStateHandler provides common functionality for custom states
type BaseCustomStateHandler struct {
	StateID       wizard.State
	Name          string
	Description   string
	InsertPoint   InsertionPoint
	ValidateFunc  func(*InstallerController, map[string]interface{}) error
	CanGoNext     bool
	CanGoBack     bool
	CanCancel     bool
}

// GetStateID implements CustomStateHandler
func (b *BaseCustomStateHandler) GetStateID() wizard.State {
	return b.StateID
}

// GetConfig implements CustomStateHandler
func (b *BaseCustomStateHandler) GetConfig() *wizard.StateConfig {
	config := &wizard.StateConfig{
		Name:        b.Name,
		Description: b.Description,
		CanGoNext:   b.CanGoNext,
		CanGoBack:   b.CanGoBack,
		CanCancel:   b.CanCancel,
		Transitions: make(map[wizard.Action]wizard.State),
	}

	if b.ValidateFunc != nil {
		config.ValidateFunc = func(data map[string]interface{}) error {
			// This will be set by the controller when registering
			return nil
		}
	}

	return config
}

// GetInsertionPoint implements CustomStateHandler
func (b *BaseCustomStateHandler) GetInsertionPoint() InsertionPoint {
	return b.InsertPoint
}

// HandleEnter provides default implementation
func (b *BaseCustomStateHandler) HandleEnter(controller *InstallerController, data map[string]interface{}) error {
	return nil
}

// HandleLeave provides default implementation
func (b *BaseCustomStateHandler) HandleLeave(controller *InstallerController, data map[string]interface{}) error {
	return nil
}

// Validate provides default implementation
func (b *BaseCustomStateHandler) Validate(controller *InstallerController, data map[string]interface{}) error {
	if b.ValidateFunc != nil {
		return b.ValidateFunc(controller, data)
	}
	return nil
}
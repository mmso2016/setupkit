// Package webview provides the thin-layer WebView integration for SetupKit
package webview

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"sync"
	
	"github.com/mmso2016/setupkit/pkg/wizard"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ThinLayerApp manages the WebView as a thin rendering layer
type ThinLayerApp struct {
	ctx        context.Context
	dfa        *wizard.DFA
	templates  *TemplateRegistry
	mu         sync.RWMutex
	
	// Current rendered state
	currentHTML string
	
	// Channel for UI events
	events     chan UIEvent
	
	// SSR configuration
	ssrConfig  SSRConfig
}

// SSRConfig defines server-side rendering configuration
type SSRConfig struct {
	TemplateDir    string
	CacheTemplates bool
	ThemeConfig    ThemeConfig
	MinifyHTML     bool
}

// UIEvent represents an event from the WebView
type UIEvent struct {
	Type    string                 `json:"type"`
	Action  string                 `json:"action"`
	Data    map[string]interface{} `json:"data"`
	StateID string                 `json:"stateId"`
}

// StateRenderer handles rendering for a specific DFA state
type StateRenderer interface {
	// RenderState generates HTML for the current state
	RenderState(state wizard.State, data map[string]interface{}) (string, error)
	
	// GetRequiredData returns data keys needed for rendering
	GetRequiredData() []string
	
	// ValidateData validates data before rendering
	ValidateData(data map[string]interface{}) error
}

// NewThinLayerApp creates a new thin-layer WebView application
func NewThinLayerApp(dfa *wizard.DFA, config SSRConfig) *ThinLayerApp {
	return &ThinLayerApp{
		dfa:       dfa,
		templates: NewTemplateRegistry(),
		events:    make(chan UIEvent, 100),
		ssrConfig: config,
	}
}

// Startup is called when the app starts
func (a *ThinLayerApp) Startup(ctx context.Context) {
	a.ctx = ctx
	
	// Initialize templates
	if err := a.initializeTemplates(); err != nil {
		runtime.LogError(a.ctx, "Failed to initialize templates: "+err.Error())
	}
	
	// Start event processor
	go a.processEvents()
	
	// Render initial state
	if err := a.renderCurrentState(); err != nil {
		runtime.LogError(a.ctx, "Failed to render initial state: "+err.Error())
	}
}

// HandleAction processes an action from the WebView
func (a *ThinLayerApp) HandleAction(action string, data map[string]interface{}) (map[string]interface{}, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Log the action for debugging
	runtime.LogDebug(a.ctx, fmt.Sprintf("HandleAction: %s with data: %v", action, data))
	
	// Update DFA data if provided
	if data != nil {
		for k, v := range data {
			if err := a.dfa.SetData(k, v); err != nil {
				return nil, fmt.Errorf("failed to set data: %w", err)
			}
		}
	}
	
	// Execute DFA transition based on action
	var err error
	switch wizard.Action(action) {
	case wizard.ActionNext:
		err = a.dfa.Next()
	case wizard.ActionBack:
		err = a.dfa.Back()
	case wizard.ActionSkip:
		err = a.dfa.Skip()
	case wizard.ActionCancel:
		err = a.dfa.Cancel()
	default:
		err = a.dfa.Transition(wizard.Action(action))
	}
	
	if err != nil {
		return nil, err
	}
	
	// Render new state
	if err := a.renderCurrentState(); err != nil {
		return nil, fmt.Errorf("failed to render state: %w", err)
	}
	
	// Return response with new state info
	return map[string]interface{}{
		"state":          string(a.dfa.CurrentState()),
		"availableActions": a.getAvailableActions(),
		"html":           a.currentHTML,
	}, nil
}

// renderCurrentState renders the current DFA state to HTML
func (a *ThinLayerApp) renderCurrentState() error {
	state := a.dfa.CurrentState()
	data := a.dfa.GetAllData()
	
	// Get template for current state
	tmpl, err := a.templates.GetTemplate(string(state))
	if err != nil {
		// Try default template as fallback
		tmpl, err = a.templates.GetTemplate("default")
		if err != nil {
			return fmt.Errorf("no template for state %s: %w", state, err)
		}
	}
	
	// Prepare template data
	templateData := map[string]interface{}{
		"State":     string(state),
		"Data":      data,
		"Actions":   a.getAvailableActions(),
		"Theme":     a.ssrConfig.ThemeConfig,
		"AppConfig": a.getAppConfig(),
	}
	
	// Render HTML
	html, err := a.executeTemplate(tmpl, templateData)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}
	
	// Optionally minify HTML
	if a.ssrConfig.MinifyHTML {
		html = a.minifyHTML(html)
	}
	
	// Store rendered HTML
	a.currentHTML = html
	
	// Emit to WebView
	runtime.EventsEmit(a.ctx, "state-rendered", map[string]interface{}{
		"html":  html,
		"state": string(state),
	})
	
	return nil
}

// GetCurrentHTML returns the currently rendered HTML
func (a *ThinLayerApp) GetCurrentHTML() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentHTML
}

// processEvents handles events from the UI
func (a *ThinLayerApp) processEvents() {
	for event := range a.events {
		switch event.Type {
		case "action":
			// Process action through DFA
			_, err := a.HandleAction(event.Action, event.Data)
			if err != nil {
				runtime.LogError(a.ctx, "Event processing error: "+err.Error())
				// Emit error to UI
				runtime.EventsEmit(a.ctx, "error", map[string]string{
					"message": err.Error(),
				})
			}
			
		case "validate":
			// Validate current state
			err := a.dfa.ValidateCurrentState()
			runtime.EventsEmit(a.ctx, "validation-result", map[string]interface{}{
				"valid": err == nil,
				"error": err,
			})
			
		case "data-update":
			// Update data without transition
			for k, v := range event.Data {
				a.dfa.SetData(k, v)
			}
		}
	}
}

// Additional helper methods and types continue...
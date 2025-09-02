// Package templates provides Scriggo template engine integration for SetupKit
package templates

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	
	"github.com/open2b/scriggo"
	"github.com/open2b/scriggo/builtin"
	"github.com/open2b/scriggo/native"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// ScriggoRenderer provides server-side rendering using Scriggo template engine
type ScriggoRenderer struct {
	mu        sync.RWMutex
	fsys      fs.FS
	programs  map[string]*scriggo.Program
	globals   native.Declarations
	opts      *scriggo.BuildOptions
	cache     bool
}

// RendererConfig configures the Scriggo renderer
type RendererConfig struct {
	// TemplateFS is the filesystem containing templates
	TemplateFS fs.FS
	
	// CacheTemplates enables template caching
	CacheTemplates bool
	
	// Globals are variables/functions available in all templates
	Globals map[string]interface{}
	
	// Extensions for template files (default: .html)
	Extensions []string
}

// NewScriggoRenderer creates a new Scriggo-based renderer
func NewScriggoRenderer(config RendererConfig) (*ScriggoRenderer, error) {
	renderer := &ScriggoRenderer{
		fsys:     config.TemplateFS,
		programs: make(map[string]*scriggo.Program),
		cache:    config.CacheTemplates,
	}
	
	// Setup globals - these are available in all templates
	renderer.globals = native.Declarations{
		// Helper functions
		"formatSize": native.Function{
			Name: "formatSize",
			Function: func(size int64) string {
				const unit = 1024
				if size < unit {
					return fmt.Sprintf("%d B", size)
				}
				div, exp := int64(unit), 0
				for n := size / unit; n >= unit; n /= unit {
					div *= unit
					exp++
				}
				return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
			},
		},
		
		"isState": native.Function{
			Name: "isState",
			Function: func(current, target string) bool {
				return current == target
			},
		},
		
		"hasAction": native.Function{
			Name: "hasAction",
			Function: func(actions []map[string]interface{}, action string) bool {
				for _, a := range actions {
					if a["id"] == action {
						return true
					}
				}
				return false
			},
		},
	}
	
	// Add user-provided globals
	for name, value := range config.Globals {
		switch v := value.(type) {
		case func(...interface{}) interface{}:
			renderer.globals[name] = native.Function{
				Name:     name,
				Function: v,
			}
		default:
			renderer.globals[name] = native.Var{
				Name:  name,
				Value: v,
			}
		}
	}
	
	// Build options for Scriggo
	renderer.opts = &scriggo.BuildOptions{
		Globals:          renderer.globals,
		MarkdownConverter: nil, // Can add markdown support if needed
	}
	
	return renderer, nil
}

// RenderState renders a specific state template with data
func (r *ScriggoRenderer) RenderState(state wizard.State, data map[string]interface{}) (string, error) {
	templateName := r.getTemplateName(state)
	
	// Get or compile the template
	program, err := r.getProgram(templateName)
	if err != nil {
		// Try default template as fallback
		program, err = r.getProgram("default.html")
		if err != nil {
			return "", fmt.Errorf("no template found for state %s: %w", state, err)
		}
	}
	
	// Run the template with data
	var output strings.Builder
	err = program.Run(context.Background(), &output, data)
	if err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	
	return output.String(), nil
}

// getProgram retrieves or compiles a template program
func (r *ScriggoRenderer) getProgram(name string) (*scriggo.Program, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check cache if enabled
	if r.cache {
		if program, exists := r.programs[name]; exists {
			return program, nil
		}
	}
	
	// Build the template
	program, err := scriggo.BuildTemplate(r.fsys, name, r.opts)
	if err != nil {
		return nil, err
	}
	
	// Cache if enabled
	if r.cache {
		r.programs[name] = program
	}
	
	return program, nil
}

// Example Scriggo template with advanced features
const exampleTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{ $.AppName }} - {{ $.State }}</title>
    {{ render "partials/styles.html" $ }}
</head>
<body>
    <div class="wizard-container">
        {{ render "partials/header.html" $ }}
        
        <div class="wizard-content">
            {{ switch $.State }}
            {{ case "welcome" }}
                {{ render "states/welcome.html" $ }}
            {{ case "license" }}
                {{ render "states/license.html" $ }}
            {{ case "components" }}
                {{ render "states/components.html" $ }}
            {{ default }}
                {{ render "states/default.html" $ }}
            {{ end }}
        </div>
        
        {{ render "partials/footer.html" $ }}
    </div>
    
    {{ render "partials/scripts.html" $ }}
</body>
</html>
`

// StateHandler using Scriggo renderer
type StateHandler struct {
	renderer *ScriggoRenderer
	dfa      *wizard.DFA
}

// HandleStateTransition processes state transition and renders new HTML
func (h *StateHandler) HandleStateTransition(action wizard.Action) (string, error) {
	// Execute DFA transition
	var err error
	switch action {
	case wizard.ActionNext:
		err = h.dfa.Next()
	case wizard.ActionBack:
		err = h.dfa.Back()
	default:
		err = h.dfa.Transition(action)
	}
	
	if err != nil {
		return "", err
	}
	
	// Get current state and data
	currentState := h.dfa.CurrentState()
	data := h.dfa.GetAllData()
	
	// Add metadata for templates
	templateData := map[string]interface{}{
		"State":   string(currentState),
		"Data":    data,
		"Actions": h.getAvailableActions(),
	}
	
	// Render the new state
	return h.renderer.RenderState(currentState, templateData)
}

// Embedded templates for production
//go:embed templates/*.html templates/states/*.html templates/partials/*.html
var templateFS embed.FS

// CreateProductionRenderer creates a renderer with embedded templates
func CreateProductionRenderer() (*ScriggoRenderer, error) {
	return NewScriggoRenderer(RendererConfig{
		TemplateFS:     templateFS,
		CacheTemplates: true, // Always cache in production
		Globals: map[string]interface{}{
			"appVersion": "1.0.0",
			"buildDate":  "2025-01-24",
		},
	})
}
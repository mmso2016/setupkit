// Package views provides a unified view abstraction for both HTML and CLI rendering
// This implements Server-Side Rendering (SSR) patterns to generate output for different UI modes
package views

import (
	"fmt"
	"html/template"
	"strings"
	texttemplate "text/template"

	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// ViewType defines the target rendering format
type ViewType int

const (
	ViewHTML ViewType = iota // WebView/HTML output
	ViewCLI                  // Terminal/CLI output
	ViewJSON                 // JSON API output for custom UIs
)

// ViewData represents the data model for view rendering
type ViewData struct {
	// Common data for all views
	AppName     string
	Version     string
	Publisher   string
	Website     string
	License     string
	
	// Page-specific data
	PageTitle       string
	PageDescription string
	CurrentStep     int
	TotalSteps      int
	
	// Components
	Components         []ComponentViewModel
	SelectedComponents []ComponentViewModel
	TotalSize          string
	
	// Installation
	InstallPath    string
	Progress       int
	ProgressText   string
	IsComplete     bool
	ErrorMessage   string
	
	// Navigation
	CanGoBack   bool
	CanGoNext   bool
	CanCancel   bool
	NextLabel   string
	BackLabel   string
	CancelLabel string
}

// ComponentViewModel represents a component for view rendering
type ComponentViewModel struct {
	ID          string
	Name        string
	Description string
	Size        string
	Required    bool
	Selected    bool
	Index       int
}

// ViewRenderer handles rendering for different output formats
type ViewRenderer struct {
	htmlTemplates *template.Template
	cliTemplates  *texttemplate.Template
}

// NewViewRenderer creates a new view renderer with templates
func NewViewRenderer() *ViewRenderer {
	return &ViewRenderer{
		htmlTemplates: template.Must(template.New("html").Funcs(htmlFuncMap()).Parse("<html><body>{{.}}</body></html>")),
		cliTemplates:  texttemplate.Must(texttemplate.New("cli").Funcs(cliFuncMap()).Parse("{{.}}")),
	}
}

// RenderView renders a view for the specified output type
func (vr *ViewRenderer) RenderView(viewName string, viewType ViewType, data *ViewData) (string, error) {
	switch viewType {
	case ViewHTML:
		return vr.renderHTML(viewName, data)
	case ViewCLI:
		return vr.renderCLI(viewName, data)
	case ViewJSON:
		return vr.renderJSON(viewName, data)
	default:
		return "", fmt.Errorf("unsupported view type: %v", viewType)
	}
}

func (vr *ViewRenderer) renderHTML(viewName string, data *ViewData) (string, error) {
	var buf strings.Builder
	
	// Use embedded templates if external templates not found
	tmpl := vr.htmlTemplates.Lookup(viewName + ".html")
	if tmpl == nil {
		tmpl = getEmbeddedHTMLTemplate(viewName)
	}
	
	err := tmpl.Execute(&buf, data)
	return buf.String(), err
}

func (vr *ViewRenderer) renderCLI(viewName string, data *ViewData) (string, error) {
	var buf strings.Builder
	
	// Use embedded templates if external templates not found
	tmpl := vr.cliTemplates.Lookup(viewName + ".txt")
	if tmpl == nil {
		tmpl = getEmbeddedCLITemplate(viewName)
	}
	
	err := tmpl.Execute(&buf, data)
	return buf.String(), err
}

func (vr *ViewRenderer) renderJSON(viewName string, data *ViewData) (string, error) {
	// For JSON output, return structured data
	// This could be used by custom UIs or web APIs
	return fmt.Sprintf(`{"view":"%s","data":%+v}`, viewName, data), nil
}

// Template function maps for different output types
func htmlFuncMap() template.FuncMap {
	return template.FuncMap{
		"progress": func(current, total int) string {
			if total == 0 {
				return "0%"
			}
			percent := (current * 100) / total
			return fmt.Sprintf("%d%%", percent)
		},
		"progressBar": func(percent int) string {
			return fmt.Sprintf(`<div class="progress-bar"><div class="progress-fill" style="width:%d%%"></div></div>`, percent)
		},
		"formatSize": formatSizeHelper,
		"selected": func(selected bool) string {
			if selected {
				return "checked"
			}
			return ""
		},
	}
}

func cliFuncMap() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"progress": func(current, total int) string {
			if total == 0 {
				return "0%"
			}
			percent := (current * 100) / total
			return fmt.Sprintf("%d%%", percent)
		},
		"progressBar": func(percent int) string {
			width := 40
			filled := (percent * width) / 100
			empty := width - filled
			return fmt.Sprintf("[%s%s]", strings.Repeat("=", filled), strings.Repeat(" ", empty))
		},
		"formatSize": formatSizeHelper,
		"selected": func(selected bool) string {
			if selected {
				return "X"
			}
			return " "
		},
		"required": func(required bool) string {
			if required {
				return "R"
			}
			return " "
		},
		"separator": func(width int) string {
			return strings.Repeat("=", width)
		},
		"center": func(text string, width int) string {
			if len(text) >= width {
				return text
			}
			padding := (width - len(text)) / 2
			return strings.Repeat(" ", padding) + text
		},
	}
}

func formatSizeHelper(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatSize is the exported version of formatSizeHelper
func FormatSize(bytes int64) string {
	return formatSizeHelper(bytes)
}

// Helper function to convert core.Component to ComponentViewModel
func ComponentToViewModel(comp core.Component, index int) ComponentViewModel {
	return ComponentViewModel{
		ID:          comp.ID,
		Name:        comp.Name,
		Description: comp.Description,
		Size:        formatSizeHelper(comp.Size),
		Required:    comp.Required,
		Selected:    comp.Selected,
		Index:       index,
	}
}

// Helper function to convert core.Config to ViewData
func ConfigToViewData(config *core.Config, pageTitle string) *ViewData {
	components := make([]ComponentViewModel, len(config.Components))
	var selectedComponents []ComponentViewModel
	var totalSize int64
	
	for i, comp := range config.Components {
		vm := ComponentToViewModel(comp, i+1)
		components[i] = vm
		
		if comp.Selected || comp.Required {
			selectedComponents = append(selectedComponents, vm)
			totalSize += comp.Size
		}
	}
	
	return &ViewData{
		AppName:            config.AppName,
		Version:            config.Version,
		Publisher:          config.Publisher,
		Website:            config.Website,
		License:            config.License,
		PageTitle:          pageTitle,
		Components:         components,
		SelectedComponents: selectedComponents,
		TotalSize:          formatSizeHelper(totalSize),
		InstallPath:        config.InstallDir,
		
		// Default navigation
		CanGoBack:   true,
		CanGoNext:   true,
		CanCancel:   true,
		NextLabel:   "Next",
		BackLabel:   "Back", 
		CancelLabel: "Cancel",
	}
}
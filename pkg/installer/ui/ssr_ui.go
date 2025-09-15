// Package ui provides SSR-based UI factory
package ui

import (
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/installer/ui/ssr"
	"github.com/mmso2016/setupkit/pkg/installer/ui/views"
)

// CreateSSRUI creates a new SSR-based UI for the specified mode
func CreateSSRUI(mode core.Mode) (core.UI, error) {
	var viewType views.ViewType
	
	switch mode {
	case core.ModeGUI:
		viewType = views.ViewHTML
	case core.ModeCLI:
		viewType = views.ViewCLI
	case core.ModeSilent:
		viewType = views.ViewJSON
	default:
		// Auto-detect
		if HasDisplay() {
			viewType = views.ViewHTML
		} else {
			viewType = views.ViewCLI
		}
	}
	
	return ssr.NewSSRController(viewType), nil
}

// SSRUIAdapter adapts the SSR controller to the core.UI interface
type SSRUIAdapter struct {
	*ssr.SSRController
}

// NewSSRUIAdapter creates a new SSR UI adapter
func NewSSRUIAdapter(viewType views.ViewType) core.UI {
	return &SSRUIAdapter{
		SSRController: ssr.NewSSRController(viewType),
	}
}
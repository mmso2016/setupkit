// Package ui provides the user interface abstraction for the installer
package ui

import (
	"fmt"

	"github.com/mmso2016/setupkit/installer/core"
)

// init registers the UI factory with the core package
func init() {
	core.RegisterUIFactory(CreateUI)
}

// Factory creates the appropriate UI based on the mode
func CreateUI(mode core.Mode) (core.UI, error) {
	switch mode {
	case core.ModeGUI:
		return createGUI()
	case core.ModeCLI:
		return createCLI()
	case core.ModeSilent:
		return createSilent()
	case core.ModeAuto:
		return detectBestUI()
	default:
		return nil, fmt.Errorf("unknown mode: %v", mode)
	}
}

// detectBestUI determines the best UI mode for the current environment
func detectBestUI() (core.UI, error) {
	// Check if we're in a GUI environment
	if hasDisplay() {
		// Try to create GUI
		ui, err := createGUI()
		if err == nil {
			return ui, nil
		}
		// Fall back to CLI if GUI fails
	}

	// Default to CLI
	return createCLI()
}

// hasDisplay checks if a display is available
func hasDisplay() bool {
	// This is a simplified check
	// TODO: Implement proper display detection for each platform
	return false // For now, default to CLI
}

// createSilent creates a silent/headless UI
func createSilent() (core.UI, error) {
	return NewSilentUI(), nil
}

// NewSilentUI is defined in silent.go

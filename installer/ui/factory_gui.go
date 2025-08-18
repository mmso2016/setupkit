//go:build !wails && !nogui
// +build !wails,!nogui

package ui

import (
	"github.com/mmso2016/setupkit/installer/core"
)

// createGUI creates a basic GUI-based UI (fallback implementation)
func createGUI() (core.UI, error) {
	// Basic GUI implementation - could use native dialogs or fall back to CLI
	// For now, fall back to CLI since no GUI framework is available
	return createCLI()
}

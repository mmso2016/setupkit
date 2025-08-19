//go:build wails
// +build wails

package ui

import (
	"fmt"
	"github.com/mmso2016/setupkit/installer/core"
	"github.com/mmso2016/setupkit/installer/ui/wails"
)

// createGUI creates a Wails-based GUI
func createGUI() (core.UI, error) {
	// Note: Wails UI requires the main application to be structured as a Wails app
	// If this fails, fall back to CLI
	ui := wails.New()
	if ui == nil {
		return nil, fmt.Errorf("failed to create Wails UI - falling back to CLI")
	}
	return ui, nil
}

//go:build !wails && !nogui
// +build !wails,!nogui

package ui

import (
	"fmt"

	"github.com/mmso2016/setupkit/installer/core"
)

// createGUI creates a basic GUI-based UI (fallback implementation)
func createGUI() (core.UI, error) {
	// GUI support not available without specific build tags
	return nil, fmt.Errorf("GUI support not available - compile with wails build tag or use --mode cli")
}

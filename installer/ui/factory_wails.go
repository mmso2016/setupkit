//go:build wails
// +build wails

package ui

import (
	"github.com/mmso2016/setupkit/installer/core"
	"github.com/mmso2016/setupkit/installer/ui/wails"
)

// createGUI creates a Wails-based GUI
func createGUI() (core.UI, error) {
	return wails.New(), nil
}

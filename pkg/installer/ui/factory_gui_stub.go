//go:build nogui
// +build nogui

package ui

import (
	"fmt"

	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// createGUI returns an error when GUI is not available
func createGUI() (core.UI, error) {
	return nil, fmt.Errorf("GUI support not compiled in")
}

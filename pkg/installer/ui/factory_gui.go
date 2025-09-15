//go:build !nogui
// +build !nogui

package ui

import (
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// createGUI creates a native WebView GUI UI
func createGUI() (core.UI, error) {
	return NewWebViewAdapter(), nil
}

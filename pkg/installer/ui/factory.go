// Package ui provides the user interface abstraction for the installer
package ui

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// init registers the UI factory with the core package
func init() {
	core.RegisterUIFactory(CreateUI)
}

// Factory creates the appropriate UI based on the mode
func CreateUI(mode core.Mode) (core.UI, error) {
	switch mode {
	case core.ModeGUI:
		return createWebViewGUI()  // Lorca WebView GUI
	case core.ModeBrowser:
		return createGUIDFA() // Browser-based GUI
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
	if HasDisplay() {
		// Try to create WebView GUI
		ui, err := createWebViewGUI()
		if err == nil {
			return ui, nil
		}
		// Fall back to CLI if WebView GUI fails
	}

	// Default to CLI
	return createCLI()
}

// HasDisplay checks if a display is available
func HasDisplay() bool {
	// On Windows, we generally have a display if not running as a service
	// On Unix systems, check for DISPLAY environment variable
	// This is a simplified check - improve as needed
	if runtime.GOOS == "windows" {
		return true // Windows usually has display unless running as service
	}
	// Unix/Linux: check for DISPLAY or WAYLAND_DISPLAY
	return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

// createSilent creates a silent/headless UI
func createSilent() (core.UI, error) {
	return NewSilentUIDFA(), nil
}

// createWebViewGUI creates a WebView-based GUI using Lorca
func createWebViewGUI() (core.UI, error) {
	return NewWebViewAdapter(), nil
}

// NewSilentUI is defined in silent.go

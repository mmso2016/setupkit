//go:build !nogui
// +build !nogui

package ui

import (
	"fmt"

	"github.com/mmso2016/setupkit/pkg/installer/controller"
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// webViewAdapter adapts the WebView GUI to core.UI interface
type webViewAdapter struct {
	webview    controller.InstallerView
	controller *controller.InstallerController
}

// NewWebViewAdapter creates a new WebView adapter for core.UI interface
func NewWebViewAdapter() core.UI {
	return &webViewAdapter{
		webview: NewWebViewGUI(),
	}
}

// Initialize implements core.UI interface
func (w *webViewAdapter) Initialize(ctx *core.Context) error {
	// Initialize the WebView GUI
	if initializer, ok := w.webview.(interface{ Initialize(*core.Context) error }); ok {
		if err := initializer.Initialize(ctx); err != nil {
			return err
		}
	}

	// Get installer from context
	installer, ok := ctx.Metadata["installer"].(*core.Installer)
	if !ok {
		return fmt.Errorf("no installer found in context")
	}

	// Create DFA controller
	w.controller = controller.NewInstallerController(ctx.Config, installer)

	// Set controller on WebView
	if setController, ok := w.webview.(interface{ SetController(*controller.InstallerController) }); ok {
		setController.SetController(w.controller)
	}

	// Set WebView as the controller's view
	w.controller.SetView(w.webview)

	return nil
}

// Run implements core.UI interface
func (w *webViewAdapter) Run() error {
	// For WebView GUI, use the UI's Run() method
	if runner, ok := w.webview.(interface{ Run() error }); ok {
		return runner.Run()
	}

	// Fallback: start DFA controller
	if w.controller != nil {
		return w.controller.Start()
	}

	return fmt.Errorf("WebView GUI not properly initialized")
}

// Shutdown implements core.UI interface
func (w *webViewAdapter) Shutdown() error {
	if shutdowner, ok := w.webview.(interface{ Shutdown() error }); ok {
		return shutdowner.Shutdown()
	}
	return nil
}

// ShowWelcome implements core.UI interface
func (w *webViewAdapter) ShowWelcome() error {
	if welcomer, ok := w.webview.(interface{ ShowWelcome() error }); ok {
		return welcomer.ShowWelcome()
	}
	return nil
}

// ShowLicense implements core.UI interface
func (w *webViewAdapter) ShowLicense(license string) (accepted bool, err error) {
	if licenser, ok := w.webview.(interface{ ShowLicense(string) (bool, error) }); ok {
		return licenser.ShowLicense(license)
	}
	return true, nil
}

// SelectComponents implements core.UI interface
func (w *webViewAdapter) SelectComponents(components []core.Component) ([]core.Component, error) {
	if selector, ok := w.webview.(interface{ ShowComponents([]core.Component) ([]core.Component, error) }); ok {
		return selector.ShowComponents(components)
	}
	// Default selection
	var result []core.Component
	for _, comp := range components {
		if comp.Selected || comp.Required {
			result = append(result, comp)
		}
	}
	return result, nil
}

// SelectInstallPath implements core.UI interface
func (w *webViewAdapter) SelectInstallPath(defaultPath string) (string, error) {
	if pathSelector, ok := w.webview.(interface{ ShowInstallPath(string) (string, error) }); ok {
		return pathSelector.ShowInstallPath(defaultPath)
	}
	return defaultPath, nil
}

// ShowProgress implements core.UI interface
func (w *webViewAdapter) ShowProgress(progress *core.Progress) error {
	if progresser, ok := w.webview.(interface{ ShowProgress(*core.Progress) error }); ok {
		return progresser.ShowProgress(progress)
	}
	return nil
}

// ShowError implements core.UI interface
func (w *webViewAdapter) ShowError(err error, canRetry bool) (retry bool, errOut error) {
	if errorShower, ok := w.webview.(interface{ ShowError(error, bool) (bool, error) }); ok {
		return errorShower.ShowError(err, canRetry)
	}
	return false, nil
}

// ShowSuccess implements core.UI interface
func (w *webViewAdapter) ShowSuccess(summary *core.InstallSummary) error {
	if successer, ok := w.webview.(interface{ ShowComplete(*core.InstallSummary) error }); ok {
		return successer.ShowComplete(summary)
	}
	return nil
}

// RequestElevation implements core.UI interface
func (w *webViewAdapter) RequestElevation(reason string) (bool, error) {
	if elevationRequester, ok := w.webview.(interface{ RequestElevation(string) (bool, error) }); ok {
		return elevationRequester.RequestElevation(reason)
	}
	return true, nil
}
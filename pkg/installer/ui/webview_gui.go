//go:build !nogui
// +build !nogui

package ui

import (
	"fmt"
	"log"

	"github.com/jchv/go-webview2"
	"github.com/mmso2016/setupkit/pkg/html"
	"github.com/mmso2016/setupkit/pkg/installer/controller"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// webViewNativeGUI implements the InstallerView interface for WebView2 control
type webViewNativeGUI struct {
	context    *core.Context
	controller *controller.InstallerController
	webview    webview2.WebView
	renderer   *html.SSRRenderer

	// Current state
	currentState wizard.State
	isVisible    bool

	// Control channels
	done     chan struct{}
	finished chan struct{}

	// User input storage
	userInputs map[string]interface{}
}

// NewWebViewGUI creates a new native WebView GUI instance
func NewWebViewGUI() controller.InstallerView {
	return &webViewNativeGUI{}
}

// Initialize sets up the WebView GUI with context
func (w *webViewNativeGUI) Initialize(ctx *core.Context) error {
	w.context = ctx
	w.renderer = html.NewSSRRenderer()
	w.done = make(chan struct{})
	w.finished = make(chan struct{})
	w.userInputs = make(map[string]interface{})

	// Create WebView2 instance
	title := fmt.Sprintf("%s v%s - Installer", ctx.Config.AppName, ctx.Config.Version)
	wv := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     true,
		AutoFocus: true,
		DataPath:  "",
		WindowOptions: webview2.WindowOptions{
			Title:  title,
			Width:  800,
			Height: 600,
			IconId: 2, // Use default icon
		},
	})
	if wv == nil {
		return fmt.Errorf("failed to create webview2 instance")
	}
	w.webview = wv

	// Bind JavaScript functions for installer interaction
	w.setupJavaScriptBindings()

	return nil
}

// SetController assigns the shared DFA controller to this UI
func (w *webViewNativeGUI) SetController(ctrl *controller.InstallerController) {
	w.controller = ctrl
}

// Run starts the native WebView installation flow
func (w *webViewNativeGUI) Run() error {
	if w.controller == nil {
		return fmt.Errorf("no controller assigned - call SetController() first")
	}

	fmt.Println("Starting native WebView GUI installation...")

	// Show initial welcome page
	if err := w.ShowWelcome(); err != nil {
		return fmt.Errorf("failed to show welcome page: %w", err)
	}

	// Start DFA controller in goroutine
	go func() {
		if err := w.controller.Start(); err != nil {
			log.Printf("DFA controller error: %v", err)
			w.done <- struct{}{}
		}
	}()

	// Run webview (blocks until window is closed)
	w.webview.Run()

	return nil
}

func (w *webViewNativeGUI) Shutdown() error {
	if w.webview != nil {
		w.webview.Destroy()
	}
	return nil
}

// ============================================================================
// InstallerView Interface Implementation
// ============================================================================

// ShowWelcome displays the welcome screen
func (w *webViewNativeGUI) ShowWelcome() error {
	w.currentState = controller.StateWelcome
	fmt.Printf("[WebView] Showing welcome page\n")

	// Update WebView content based on current state
	w.updateWebViewContent()

	return nil
}

// ShowLicense displays license and returns acceptance
func (w *webViewNativeGUI) ShowLicense(license string) (accepted bool, err error) {
	w.currentState = controller.StateLicense
	w.userInputs["license_text"] = license
	fmt.Printf("[WebView] Showing license page\n")

	// Update WebView content for license state
	w.updateWebViewContent()

	// In WebView GUI mode, this method sets up the state
	// User interaction happens via JavaScript bindings
	// Return default acceptance for GUI flow (actual interaction is async)
	return true, nil
}

// ShowComponents displays component selection
func (w *webViewNativeGUI) ShowComponents(components []core.Component) (selected []core.Component, err error) {
	w.currentState = controller.StateComponents
	w.userInputs["available_components"] = components
	fmt.Printf("[WebView] Showing components page\n")

	// Update WebView content for components state
	w.updateWebViewContent()

	// Return default selection - should be handled via JavaScript in real usage
	var result []core.Component
	for _, comp := range components {
		if comp.Selected || comp.Required {
			result = append(result, comp)
		}
	}

	return result, nil
}

// ShowInstallPath allows user to select installation path
func (w *webViewNativeGUI) ShowInstallPath(defaultPath string) (path string, err error) {
	w.currentState = controller.StateInstallPath
	w.userInputs["default_path"] = defaultPath
	fmt.Printf("[WebView] Showing install path page\n")

	// Update WebView content for install path state
	w.updateWebViewContent()

	// Return default path - should be handled via JavaScript in real usage
	return defaultPath, nil
}

// ShowSummary displays installation summary and gets confirmation
func (w *webViewNativeGUI) ShowSummary(config *core.Config, selectedComponents []core.Component, installPath string) (proceed bool, err error) {
	w.currentState = controller.StateSummary
	w.userInputs["summary_config"] = config
	w.userInputs["summary_components"] = selectedComponents
	w.userInputs["summary_path"] = installPath
	fmt.Printf("[WebView] Showing summary page\n")

	// Update WebView content for summary state
	w.updateWebViewContent()

	// Auto-proceed for now - should wait for user confirmation via JavaScript
	return true, nil
}

// ShowProgress displays installation progress
func (w *webViewNativeGUI) ShowProgress(progress *core.Progress) error {
	w.currentState = controller.StateProgress
	w.userInputs["progress"] = progress
	fmt.Printf("[WebView] Progress: %.0f%% - %s\n", progress.OverallProgress*100, progress.ComponentName)

	// Update WebView content for progress state
	w.updateWebViewContent()

	return nil
}

// ShowComplete displays installation completion
func (w *webViewNativeGUI) ShowComplete(summary *core.InstallSummary) error {
	w.currentState = controller.StateComplete
	fmt.Printf("[WebView] Installation completed successfully!\n")

	// Update WebView content for completion state
	w.updateWebViewContent()

	// Signal completion
	go func() {
		w.done <- struct{}{}
	}()

	return nil
}

// ShowErrorMessage displays an error message
func (w *webViewNativeGUI) ShowErrorMessage(err error) error {
	fmt.Printf("[WebView] Error: %v\n", err)
	return nil
}

// OnStateChanged handles state change notifications (DFA-compliant)
func (w *webViewNativeGUI) OnStateChanged(oldState, newState wizard.State) error {
	fmt.Printf("[WebView] State transition: %s → %s\n", oldState, newState)
	w.currentState = newState

	// Update WebView content when state changes (follows DFA transitions)
	w.updateWebViewContent()

	return nil
}

// ============================================================================
// core.UI Interface Compatibility Methods
// ============================================================================

// SelectComponents - adapts InstallerView.ShowComponents to core.UI interface
func (w *webViewNativeGUI) SelectComponents(components []core.Component) ([]core.Component, error) {
	return w.ShowComponents(components)
}

// SelectInstallPath - adapts InstallerView.ShowInstallPath to core.UI interface
func (w *webViewNativeGUI) SelectInstallPath(defaultPath string) (string, error) {
	return w.ShowInstallPath(defaultPath)
}

// ShowSuccess - adapts InstallerView.ShowComplete to core.UI interface
func (w *webViewNativeGUI) ShowSuccess(summary *core.InstallSummary) error {
	return w.ShowComplete(summary)
}

// ShowError - core.UI interface method with retry logic
func (w *webViewNativeGUI) ShowError(err error, canRetry bool) (retry bool, errOut error) {
	w.ShowErrorMessage(err)
	return false, nil
}

// RequestElevation requests elevated privileges
func (w *webViewNativeGUI) RequestElevation(reason string) (bool, error) {
	fmt.Printf("[WebView] Administrative privileges required: %s\n", reason)
	// TODO: Show elevation dialog in WebView
	return true, nil
}

// ============================================================================
// ExtendedInstallerView Interface Implementation
// ============================================================================

// ShowCustomState handles custom states in WebView mode
func (w *webViewNativeGUI) ShowCustomState(stateID wizard.State, data controller.CustomStateData) (controller.CustomStateData, error) {
	fmt.Printf("[WebView] Showing custom state: %s\n", stateID)
	w.currentState = stateID
	w.userInputs["custom_state_data"] = data

	switch stateID {
	case controller.StateDBConfig:
		if config, exists := data["config"]; exists {
			fmt.Printf("[WebView] Using provided database configuration\n")
			return controller.CustomStateData{"config": config}, nil
		}

		defaultDB := controller.DefaultDatabaseConfig()
		fmt.Printf("[WebView] Using default database configuration: %s\n", defaultDB.String())
		return controller.CustomStateData{"config": defaultDB}, nil

	default:
		fmt.Printf("[WebView] Unknown custom state: %s\n", stateID)
		return data, nil
	}
}

// ============================================================================
// DFA-Compliant WebView Content Management
// ============================================================================

// updateWebViewContent renders the appropriate page based on current DFA state
func (w *webViewNativeGUI) updateWebViewContent() {
	var doc *html.Document

	switch w.currentState {
	case controller.StateLicense:
		if license, ok := w.userInputs["license_text"].(string); ok {
			doc = w.renderer.RenderLicensePage(w.context.Config, license)
		} else {
			doc = w.renderer.RenderLicensePage(w.context.Config, "No license text provided")
		}
	case controller.StateComponents:
		doc = w.renderer.RenderComponentsPage(w.context.Config)
	case controller.StateInstallPath:
		if defaultPath, ok := w.userInputs["default_path"].(string); ok {
			doc = w.renderer.RenderInstallPathPage(w.context.Config, defaultPath)
		} else {
			doc = w.renderer.RenderInstallPathPage(w.context.Config, "/opt/"+w.context.Config.AppName)
		}
	case controller.StateSummary:
		var selectedComponents []core.Component
		var installPath string

		if components, ok := w.userInputs["summary_components"].([]core.Component); ok {
			selectedComponents = components
		}
		if path, ok := w.userInputs["summary_path"].(string); ok {
			installPath = path
		}

		doc = w.renderer.RenderSummaryPage(w.context.Config, selectedComponents, installPath)
	case controller.StateProgress:
		if progress, ok := w.userInputs["progress"].(*core.Progress); ok {
			percentage := int(progress.OverallProgress * 100)
			doc = w.renderer.RenderProgressPage(w.context.Config, percentage, progress.ComponentName)
		} else {
			doc = w.renderer.RenderProgressPage(w.context.Config, 0, "Starting...")
		}
	case controller.StateComplete:
		doc = w.renderer.RenderCompletionPage(w.context.Config, true)
	default:
		doc = w.renderer.RenderWelcomePage(w.context.Config)
	}

	// Update WebView2 content
	w.webview.SetHtml(doc.Render())
}

// ============================================================================
// JavaScript Bindings for WebView Interaction
// ============================================================================

// setupJavaScriptBindings creates minimal JavaScript functions that follow DFA pattern
func (w *webViewNativeGUI) setupJavaScriptBindings() {
	// DFA-compliant bindings - only delegate to controller, no flow control
	w.webview.Bind("installerNext", func() {
		fmt.Printf("[WebView] Next button clicked from state: %s\n", w.currentState)
		go func() {
			if err := w.controller.Next(); err != nil {
				fmt.Printf("[WebView] Next transition error: %v\n", err)
			}
		}()
	})

	w.webview.Bind("installerBack", func() {
		fmt.Printf("[WebView] Back button clicked from state: %s\n", w.currentState)
		go func() {
			if err := w.controller.Back(); err != nil {
				fmt.Printf("[WebView] Back transition error: %v\n", err)
			}
		}()
	})

	w.webview.Bind("installerCancel", func() {
		fmt.Printf("[WebView] Cancel button clicked\n")
		go func() {
			if err := w.controller.Cancel(); err != nil {
				fmt.Printf("[WebView] Cancel error: %v\n", err)
			}
			w.done <- struct{}{}
		}()
	})

	// State query function for WebView to know current state
	w.webview.Bind("getCurrentState", func() string {
		return string(w.currentState)
	})
}

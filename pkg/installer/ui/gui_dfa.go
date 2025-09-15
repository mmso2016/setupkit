//go:build !nogui
// +build !nogui

package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/mmso2016/setupkit/pkg/html"
	"github.com/mmso2016/setupkit/pkg/installer/controller"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// webViewUIDFA implements the InstallerView interface for browser-based interaction
type webViewUIDFA struct {
	context    *core.Context
	controller *controller.InstallerController
	server     *http.Server
	renderer   *html.SSRRenderer

	// Current state
	currentState wizard.State
	isVisible    bool
	port         int

	// Control channels
	done     chan struct{}
	finished chan struct{}

	// User input storage
	userInputs map[string]interface{}
}

// NewGUIDFA creates a new DFA-controlled GUI instance (public interface)
func NewGUIDFA() controller.InstallerView {
	return &webViewUIDFA{}
}

// createGUIDFA creates a new DFA-controlled GUI instance
func createGUIDFA() (core.UI, error) {
	return &webViewUIDFA{}, nil
}

// Initialize sets up the GUI with context and controller
func (w *webViewUIDFA) Initialize(ctx *core.Context) error {
	w.context = ctx
	w.renderer = html.NewSSRRenderer()
	w.port = 8080 // TODO: Find available port
	w.done = make(chan struct{})
	w.finished = make(chan struct{})
	w.userInputs = make(map[string]interface{})

	// Get installer from context
	installer, ok := ctx.Metadata["installer"].(*core.Installer)
	if !ok {
		return fmt.Errorf("no installer found in context")
	}

	// Create DFA controller
	w.controller = controller.NewInstallerController(ctx.Config, installer)
	w.controller.SetView(w)

	// Setup HTTP server
	w.setupHTTPServer()

	return nil
}

// Run starts the DFA-controlled installation flow
func (w *webViewUIDFA) Run() error {
	// Start HTTP server in goroutine
	go func() {
		if err := w.server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Open browser
	url := fmt.Sprintf("http://localhost:%d", w.port)
	fmt.Printf("Opening installer UI in browser: %s\n", url)

	if err := w.openBrowser(url); err != nil {
		fmt.Printf("Failed to open browser automatically. Please open: %s\n", url)
	}

	fmt.Println("Starting DFA-controlled GUI installation...")

	// Start DFA controller
	go func() {
		if err := w.controller.Start(); err != nil {
			fmt.Printf("DFA controller error: %v\n", err)
			w.done <- struct{}{}
		}
	}()

	// Keep server running - wait for installation to finish or be cancelled
	<-w.done

	return nil
}

func (w *webViewUIDFA) Shutdown() error {
	if w.server != nil {
		w.server.Close()
	}
	return nil
}

// ============================================================================
// InstallerView Interface Implementation
// ============================================================================

// ShowWelcome displays the welcome screen
func (w *webViewUIDFA) ShowWelcome() error {
	w.currentState = controller.StateWelcome
	fmt.Printf("[GUI] Showing welcome page\n")
	return nil // Page is rendered via HTTP handler
}

// ShowLicense displays license and returns acceptance
func (w *webViewUIDFA) ShowLicense(license string) (accepted bool, err error) {
	w.currentState = controller.StateLicense
	w.userInputs["license_text"] = license
	fmt.Printf("[GUI] Showing license page\n")

	// In GUI mode, this method should just set up the state
	// User interaction happens via HTTP API handlers
	// The method returns immediately, actual acceptance is handled in HTTP handlers
	return true, nil // Default acceptance for GUI flow
}

// ShowComponents displays component selection
func (w *webViewUIDFA) ShowComponents(components []core.Component) (selected []core.Component, err error) {
	w.currentState = controller.StateComponents
	w.userInputs["available_components"] = components
	fmt.Printf("[GUI] Showing components page\n")

	// For now, return default selection - should be handled via HTTP
	var result []core.Component
	for _, comp := range components {
		if comp.Selected || comp.Required {
			result = append(result, comp)
		}
	}
	
	return result, nil
}

// ShowInstallPath allows user to select installation path
func (w *webViewUIDFA) ShowInstallPath(defaultPath string) (path string, err error) {
	w.currentState = controller.StateInstallPath
	w.userInputs["default_path"] = defaultPath
	fmt.Printf("[GUI] Showing install path page\n")

	// Return default path for now - should be handled via HTTP
	return defaultPath, nil
}

// ShowSummary displays installation summary and gets confirmation
func (w *webViewUIDFA) ShowSummary(config *core.Config, selectedComponents []core.Component, installPath string) (proceed bool, err error) {
	w.currentState = controller.StateSummary
	w.userInputs["summary_config"] = config
	w.userInputs["summary_components"] = selectedComponents
	w.userInputs["summary_path"] = installPath
	fmt.Printf("[GUI] Showing summary page\n")

	// Auto-proceed for now - should wait for user confirmation via HTTP
	return true, nil
}

// ShowProgress displays installation progress
func (w *webViewUIDFA) ShowProgress(progress *core.Progress) error {
	w.currentState = controller.StateProgress
	w.userInputs["progress"] = progress
	fmt.Printf("[GUI] Progress: %.0f%% - %s\n", progress.OverallProgress*100, progress.ComponentName)
	return nil
}

// ShowComplete displays installation completion
func (w *webViewUIDFA) ShowComplete(summary *core.InstallSummary) error {
	w.currentState = controller.StateComplete
	fmt.Printf("[GUI] Installation completed successfully!\n")
	
	// Signal completion
	go func() {
		w.done <- struct{}{}
	}()
	
	return nil
}

// ShowErrorMessage displays an error message (InstallerView interface)
func (w *webViewUIDFA) ShowErrorMessage(err error) error {
	fmt.Printf("[GUI] Error: %v\n", err)
	return nil
}

// OnStateChanged handles state change notifications
func (w *webViewUIDFA) OnStateChanged(oldState, newState wizard.State) error {
	fmt.Printf("[GUI] State transition: %s → %s\n", oldState, newState)
	w.currentState = newState
	return nil
}

// ============================================================================
// core.UI Interface Compatibility Methods
// ============================================================================

// SelectComponents - adapts InstallerView.ShowComponents to core.UI interface
func (w *webViewUIDFA) SelectComponents(components []core.Component) ([]core.Component, error) {
	return w.ShowComponents(components)
}

// SelectInstallPath - adapts InstallerView.ShowInstallPath to core.UI interface
func (w *webViewUIDFA) SelectInstallPath(defaultPath string) (string, error) {
	return w.ShowInstallPath(defaultPath)
}

// ShowSuccess - adapts InstallerView.ShowComplete to core.UI interface
func (w *webViewUIDFA) ShowSuccess(summary *core.InstallSummary) error {
	return w.ShowComplete(summary)
}

// ShowError - core.UI interface method with retry logic
func (w *webViewUIDFA) ShowError(err error, canRetry bool) (retry bool, errOut error) {
	w.ShowErrorMessage(err)
	// GUI doesn't support interactive retry - return false
	return false, nil
}

// RequestElevation requests elevated privileges
func (w *webViewUIDFA) RequestElevation(reason string) (bool, error) {
	fmt.Printf("[GUI] Administrative privileges required: %s\n", reason)
	// Auto-grant for now - should be handled via GUI dialog
	return true, nil
}

// ============================================================================
// ExtendedInstallerView Interface Implementation
// ============================================================================

// ShowCustomState handles custom states in GUI mode
func (w *webViewUIDFA) ShowCustomState(stateID wizard.State, data controller.CustomStateData) (controller.CustomStateData, error) {
	fmt.Printf("[GUI] Showing custom state: %s\n", stateID)
	w.currentState = stateID
	w.userInputs["custom_state_data"] = data

	switch stateID {
	case controller.StateDBConfig:
		// Database configuration - use defaults or pre-configured values
		if config, exists := data["config"]; exists {
			fmt.Printf("[GUI] Using provided database configuration\n")
			return controller.CustomStateData{"config": config}, nil
		}

		// Use default database configuration
		defaultDB := controller.DefaultDatabaseConfig()
		fmt.Printf("[GUI] Using default database configuration: %s\n", defaultDB.String())
		return controller.CustomStateData{"config": defaultDB}, nil

	default:
		fmt.Printf("[GUI] Unknown custom state: %s\n", stateID)
		// For GUI mode, return defaults for unknown states
		return data, nil
	}
}

// ============================================================================
// HTTP Server Setup and Handlers
// ============================================================================

// setupHTTPServer sets up the HTTP server for browser-based UI
func (w *webViewUIDFA) setupHTTPServer() {
	mux := http.NewServeMux()

	// Main page handler
	mux.HandleFunc("/", w.handleMainPage)

	// API handlers for installer interaction
	mux.HandleFunc("/api/next", w.handleNext)
	mux.HandleFunc("/api/prev", w.handlePrev)
	mux.HandleFunc("/api/cancel", w.handleCancel)
	mux.HandleFunc("/api/finish", w.handleFinish)
	mux.HandleFunc("/api/components", w.handleComponents)
	mux.HandleFunc("/api/license", w.handleLicense)
	mux.HandleFunc("/api/path", w.handlePath)

	w.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", w.port),
		Handler: mux,
	}
}

// HTTP handler functions

func (w *webViewUIDFA) handleMainPage(wr http.ResponseWriter, req *http.Request) {
	// Generate appropriate page based on current state
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

	wr.Header().Set("Content-Type", "text/html")
	wr.Write([]byte(doc.Render()))
}

func (w *webViewUIDFA) handleNext(wr http.ResponseWriter, req *http.Request) {
	fmt.Printf("[GUI] Next button clicked from state: %s\n", w.currentState)
	
	// Forward to DFA controller
	go func() {
		if err := w.controller.Next(); err != nil {
			fmt.Printf("[GUI] Next transition error: %v\n", err)
		}
	}()
	
	fmt.Fprintf(wr, "{\"status\": \"ok\", \"action\": \"next\"}")
}

func (w *webViewUIDFA) handlePrev(wr http.ResponseWriter, req *http.Request) {
	fmt.Printf("[GUI] Back button clicked from state: %s\n", w.currentState)
	
	// Forward to DFA controller
	go func() {
		if err := w.controller.Back(); err != nil {
			fmt.Printf("[GUI] Back transition error: %v\n", err)
		}
	}()
	
	fmt.Fprintf(wr, "{\"status\": \"ok\", \"action\": \"prev\"}")
}

func (w *webViewUIDFA) handleCancel(wr http.ResponseWriter, req *http.Request) {
	fmt.Printf("[GUI] Cancel button clicked\n")
	
	// Forward to DFA controller
	go func() {
		if err := w.controller.Cancel(); err != nil {
			fmt.Printf("[GUI] Cancel error: %v\n", err)
		}
		w.done <- struct{}{}
	}()
	
	fmt.Fprintf(wr, "{\"status\": \"cancelled\"}")
}

func (w *webViewUIDFA) handleFinish(wr http.ResponseWriter, req *http.Request) {
	fmt.Printf("[GUI] Finish button clicked\n")
	
	fmt.Fprintf(wr, "{\"status\": \"finished\"}")
	
	// Signal completion
	go func() {
		w.done <- struct{}{}
	}()
}

func (w *webViewUIDFA) handleComponents(wr http.ResponseWriter, req *http.Request) {
	// Return current component state as JSON
	if components, ok := w.userInputs["available_components"].([]core.Component); ok {
		wr.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(components)
		wr.Write(data)
	} else {
		wr.Write([]byte("[]"))
	}
}

func (w *webViewUIDFA) handleLicense(wr http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		// Handle license acceptance
		accepted := req.FormValue("accepted") == "true"
		w.userInputs["license_accepted"] = accepted
		
		fmt.Printf("[GUI] License accepted: %v\n", accepted)
		
		if accepted {
			go func() {
				w.controller.Next()
			}()
		}
		
		fmt.Fprintf(wr, "{\"status\": \"ok\", \"accepted\": %s}", strconv.FormatBool(accepted))
	} else {
		// Return license text
		if license, ok := w.userInputs["license_text"].(string); ok {
			wr.Header().Set("Content-Type", "application/json")
			data, _ := json.Marshal(map[string]string{"license": license})
			wr.Write(data)
		} else {
			wr.Write([]byte("{\"license\": \"\"}"))
		}
	}
}

func (w *webViewUIDFA) handlePath(wr http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		// Handle path selection
		path := req.FormValue("path")
		if path == "" {
			path = w.userInputs["default_path"].(string)
		}
		w.userInputs["selected_path"] = path
		
		fmt.Printf("[GUI] Install path selected: %s\n", path)
		
		go func() {
			w.controller.Next()
		}()
		
		fmt.Fprintf(wr, "{\"status\": \"ok\", \"path\": \"%s\"}", path)
	} else {
		// Return default path
		if defaultPath, ok := w.userInputs["default_path"].(string); ok {
			wr.Header().Set("Content-Type", "application/json")
			data, _ := json.Marshal(map[string]string{"path": defaultPath})
			wr.Write(data)
		} else {
			wr.Write([]byte("{\"path\": \"\"}"))
		}
	}
}

func (w *webViewUIDFA) openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}
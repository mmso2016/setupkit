//go:build !nogui
// +build !nogui

package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/mmso2016/setupkit/pkg/html"
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// createGUI creates a DFA-controlled webview-based GUI UI
func createGUI() (core.UI, error) {
	return createGUIDFA()
}

// webViewUI implements a browser-based UI using HTML Builder System
type webViewUI struct {
	context   *core.Context
	installer *core.Installer
	server    *http.Server
	renderer  *html.SSRRenderer
	
	// Current state
	currentPage string
	isVisible   bool
	port        int
	
	// Control channels
	done       chan struct{}
	finished   chan struct{}
}

// Implement core.UI interface methods for webViewUI

// Initialize initializes the browser-based UI
func (w *webViewUI) Initialize(ctx *core.Context) error {
	w.context = ctx
	w.renderer = html.NewSSRRenderer()
	w.port = 8080 // TODO: Find available port
	w.done = make(chan struct{})
	w.finished = make(chan struct{})
	
	// Store installer reference for later use
	if installer, ok := ctx.Metadata["installer"].(*core.Installer); ok {
		w.installer = installer
	}
	
	// Setup HTTP server
	w.setupHTTPServer()
	
	return nil
}

// Run starts the browser-based UI
func (w *webViewUI) Run() error {
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
	
	// Show welcome page first
	if err := w.ShowWelcome(); err != nil {
		return err
	}
	
	w.isVisible = true
	
	// Keep server running - wait for installation to finish or be cancelled
	<-w.done
	
	return nil
}

// Shutdown closes the browser-based UI
func (w *webViewUI) Shutdown() error {
	if w.server != nil {
		w.server.Close()
	}
	return nil
}

// ShowWelcome displays the welcome screen
func (w *webViewUI) ShowWelcome() error {
	w.currentPage = "welcome"
	
	// For browser-based UI, the content is served via HTTP
	// The actual rendering happens in the HTTP handler
	
	return nil
}

// ShowLicense displays license agreement and returns acceptance
func (w *webViewUI) ShowLicense(license string) (accepted bool, err error) {
	return false, fmt.Errorf("webview license screen not implemented")
}

// SelectComponents allows user to select installation components
func (w *webViewUI) SelectComponents(components []core.Component) ([]core.Component, error) {
	w.currentPage = "components"
	
	// For browser-based UI, return default selection for now
	var selected []core.Component
	for _, comp := range components {
		if comp.Selected || comp.Required {
			selected = append(selected, comp)
		}
	}
	
	return selected, nil
}

// SelectInstallPath allows user to select installation path
func (w *webViewUI) SelectInstallPath(defaultPath string) (string, error) {
	return "", fmt.Errorf("webview path selection not implemented")
}

// ShowProgress displays installation progress
func (w *webViewUI) ShowProgress(progress *core.Progress) error {
	w.currentPage = "progress"
	
	// For browser-based UI, the progress would be updated via WebSocket or AJAX
	fmt.Printf("Progress: %.0f%% - %s\n", progress.OverallProgress*100, progress.Message)
	
	return nil
}

// ShowError displays an error message
func (w *webViewUI) ShowError(err error, canRetry bool) (retry bool, errOut error) {
	return false, fmt.Errorf("webview error display not implemented: %w", err)
}

// ShowSuccess displays installation success
func (w *webViewUI) ShowSuccess(summary *core.InstallSummary) error {
	w.currentPage = "complete"
	
	// For browser-based UI, show completion message
	fmt.Println("✅ Installation completed successfully!")
	
	return nil
}

// RequestElevation requests elevated privileges
func (w *webViewUI) RequestElevation(reason string) (bool, error) {
	return false, fmt.Errorf("webview elevation request not implemented")
}

// setupHTTPServer sets up the HTTP server for browser-based UI
func (w *webViewUI) setupHTTPServer() {
	mux := http.NewServeMux()
	
	// Main page handler
	mux.HandleFunc("/", w.handleMainPage)
	
	// API handlers for installer interaction
	mux.HandleFunc("/api/next", w.handleNext)
	mux.HandleFunc("/api/prev", w.handlePrev)
	mux.HandleFunc("/api/cancel", w.handleCancel)
	mux.HandleFunc("/api/finish", w.handleFinish)
	mux.HandleFunc("/api/components", w.handleComponents)
	
	w.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", w.port),
		Handler: mux,
	}
}

// HTTP handler functions

func (w *webViewUI) handleMainPage(wr http.ResponseWriter, req *http.Request) {
	// Generate appropriate page based on current state
	var doc *html.Document
	
	switch w.currentPage {
	case "license":
		// TODO: Implement RenderLicensePage
		doc = w.renderer.RenderWelcomePage(w.context.Config) // Temporary placeholder
	case "components":
		doc = w.renderer.RenderComponentsPage(w.context.Config)
	case "install-path":
		// TODO: Implement RenderInstallPathPage
		doc = w.renderer.RenderWelcomePage(w.context.Config) // Temporary placeholder
	case "summary":
		// TODO: Implement RenderSummaryPage
		doc = w.renderer.RenderWelcomePage(w.context.Config) // Temporary placeholder
	case "progress":
		doc = w.renderer.RenderProgressPage(w.context.Config, 50, "Installing...")
	case "complete":
		doc = w.renderer.RenderCompletionPage(w.context.Config, true)
	default:
		doc = w.renderer.RenderWelcomePage(w.context.Config)
	}
	
	wr.Header().Set("Content-Type", "text/html")
	wr.Write([]byte(doc.Render()))
}

func (w *webViewUI) handleNext(wr http.ResponseWriter, req *http.Request) {
	// Handle next page navigation - follow same flow as CLI
	switch w.currentPage {
	case "welcome", "":
		if w.context.Config.License != "" {
			w.currentPage = "license"
		} else {
			w.currentPage = "components"
		}
	case "license":
		w.currentPage = "components"
	case "components":
		w.currentPage = "install-path"
	case "install-path":
		w.currentPage = "summary"
	case "summary":
		w.currentPage = "progress"
		// Start installation process
		go w.performInstallation()
	case "progress":
		w.currentPage = "complete"
	}
	
	fmt.Fprintf(wr, "{\"status\": \"ok\", \"action\": \"next\"}")
}

func (w *webViewUI) handlePrev(wr http.ResponseWriter, req *http.Request) {
	// Handle previous page navigation - reverse of next flow
	switch w.currentPage {
	case "license":
		w.currentPage = "welcome"
	case "components":
		if w.context.Config.License != "" {
			w.currentPage = "license"
		} else {
			w.currentPage = "welcome"
		}
	case "install-path":
		w.currentPage = "components"
	case "summary":
		w.currentPage = "install-path"
	case "progress":
		w.currentPage = "summary"
	}
	
	fmt.Fprintf(wr, "{\"status\": \"ok\", \"action\": \"prev\"}")
}

func (w *webViewUI) handleCancel(wr http.ResponseWriter, req *http.Request) {
	// Cancel installation
	fmt.Fprintf(wr, "{\"status\": \"cancelled\"}")
	
	// Signal that we're done
	go func() {
		w.done <- struct{}{}
	}()
}

func (w *webViewUI) handleFinish(wr http.ResponseWriter, req *http.Request) {
	// Finish installation
	fmt.Fprintf(wr, "{\"status\": \"finished\"}")
	
	// Signal that we're done
	go func() {
		w.done <- struct{}{}
	}()
}

func (w *webViewUI) handleComponents(wr http.ResponseWriter, req *http.Request) {
	// Return current component state as JSON
	if w.context != nil && w.context.Config.Components != nil {
		wr.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(w.context.Config.Components)
		wr.Write(data)
	} else {
		wr.Write([]byte("[]"))
	}
}

func (w *webViewUI) performInstallation() {
	// Execute the installation using the installer
	if w.installer != nil {
		w.installer.SetUI(w)
		err := w.installer.ExecuteInstallation()
		
		// After installation, move to completion page
		if err != nil {
			fmt.Printf("Installation failed: %v\n", err)
		} else {
			w.currentPage = "complete"
		}
	}
}

func (w *webViewUI) openBrowser(url string) error {
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

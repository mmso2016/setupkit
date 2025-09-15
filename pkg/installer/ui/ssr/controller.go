// Package ssr implements Server-Side Rendering UI controller
// This provides a unified interface for both WebView and CLI using the same templates and logic
package ssr

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/installer/ui/views"
)

// SSRController handles Server-Side Rendering for both HTML and CLI
type SSRController struct {
	renderer *views.ViewRenderer
	viewType views.ViewType
	context  *core.Context
	reader   *bufio.Reader
	
	// State
	currentView      string
	viewData         *views.ViewData
	licenseAccepted  bool
	selectedComponents []core.Component
	installPath      string
}

// NewSSRController creates a new SSR controller
func NewSSRController(viewType views.ViewType) *SSRController {
	return &SSRController{
		renderer: views.NewViewRenderer(),
		viewType: viewType,
		reader:   bufio.NewReader(os.Stdin),
	}
}

// Initialize initializes the controller with context
func (c *SSRController) Initialize(ctx *core.Context) error {
	c.context = ctx
	c.installPath = ctx.Config.InstallDir
	
	// Initialize selected components
	for _, comp := range ctx.Config.Components {
		if comp.Selected || comp.Required {
			c.selectedComponents = append(c.selectedComponents, comp)
		}
	}
	
	return nil
}

// Run starts the SSR UI flow
func (c *SSRController) Run() error {
	// Navigation flow through installer steps
	steps := []string{"welcome", "license", "components", "location", "summary", "progress", "complete"}
	
	for _, step := range steps {
		if err := c.showView(step); err != nil {
			return err
		}
		
		// Handle user interaction for each step
		if err := c.handleInteraction(step); err != nil {
			return err
		}
	}
	
	return nil
}

// showView renders and displays a view
func (c *SSRController) showView(viewName string) error {
	c.currentView = viewName
	
	// Prepare view data
	c.updateViewData(viewName)
	
	// Render view
	output, err := c.renderer.RenderView(viewName, c.viewType, c.viewData)
	if err != nil {
		return fmt.Errorf("failed to render view %s: %w", viewName, err)
	}
	
	// Display output
	if c.viewType == views.ViewCLI {
		fmt.Print(output)
	} else if c.viewType == views.ViewHTML {
		// For HTML, we'd send this to WebView
		// For now, just print (would be sent to WebView in real implementation)
		fmt.Println("HTML Output for WebView:")
		fmt.Println(output)
	}
	
	return nil
}

// updateViewData prepares data for the current view
func (c *SSRController) updateViewData(viewName string) {
	// Create base view data
	c.viewData = views.ConfigToViewData(c.context.Config, getPageTitle(viewName))
	
	// Update with current state
	c.viewData.InstallPath = c.installPath
	c.viewData.SelectedComponents = make([]views.ComponentViewModel, len(c.selectedComponents))
	
	var totalSize int64
	for i, comp := range c.selectedComponents {
		c.viewData.SelectedComponents[i] = views.ComponentToViewModel(comp, i+1)
		totalSize += comp.Size
	}
	c.viewData.TotalSize = views.FormatSize(totalSize)
	
	// Update components with current selection state
	for i, comp := range c.viewData.Components {
		c.viewData.Components[i].Selected = c.isComponentSelected(comp.ID)
	}
	
	// View-specific updates
	switch viewName {
	case "welcome":
		c.viewData.PageDescription = fmt.Sprintf("Welcome to %s installation wizard", c.context.Config.AppName)
	case "license":
		c.viewData.PageDescription = "Please review and accept the license agreement"
		c.viewData.CanGoNext = c.licenseAccepted
	case "components":
		c.viewData.PageDescription = "Select components to install"
	case "location":
		c.viewData.PageDescription = "Choose installation directory"
	case "summary":
		c.viewData.PageDescription = "Review installation settings"
	case "progress":
		c.viewData.PageDescription = "Installation in progress"
		c.viewData.CanGoBack = false
		c.viewData.CanCancel = false
	case "complete":
		c.viewData.PageDescription = "Installation completed successfully"
		c.viewData.IsComplete = true
		c.viewData.CanGoBack = false
		c.viewData.CanCancel = false
	}
}

// handleInteraction processes user input for each view
func (c *SSRController) handleInteraction(viewName string) error {
	if c.viewType == views.ViewHTML {
		// For HTML views, interaction would be handled via JavaScript callbacks
		// For demo purposes, we'll simulate the interaction
		return c.simulateHTMLInteraction(viewName)
	}
	
	// Handle CLI interactions
	switch viewName {
	case "welcome":
		return c.handleWelcomeInteraction()
	case "license":
		return c.handleLicenseInteraction()
	case "components":
		return c.handleComponentsInteraction()
	case "location":
		return c.handleLocationInteraction()
	case "summary":
		return c.handleSummaryInteraction()
	case "progress":
		return c.handleProgressInteraction()
	case "complete":
		return c.handleCompleteInteraction()
	}
	
	return nil
}

func (c *SSRController) handleWelcomeInteraction() error {
	// Welcome screen - just continue
	return nil
}

func (c *SSRController) handleLicenseInteraction() error {
	input, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}
	
	response := strings.ToLower(strings.TrimSpace(input))
	c.licenseAccepted = (response == "y" || response == "yes")
	
	if !c.licenseAccepted {
		return fmt.Errorf("license not accepted")
	}
	
	return nil
}

func (c *SSRController) handleComponentsInteraction() error {
	input, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}
	
	input = strings.TrimSpace(input)
	if input == "" {
		return nil // Continue with current selection
	}
	
	// Parse component selection
	selections := strings.Split(input, ",")
	for _, s := range selections {
		s = strings.TrimSpace(s)
		if idx, err := strconv.Atoi(s); err == nil {
			if idx >= 1 && idx <= len(c.context.Config.Components) {
				comp := &c.context.Config.Components[idx-1]
				if !comp.Required {
					comp.Selected = !comp.Selected
				}
			}
		}
	}
	
	// Update selected components
	c.selectedComponents = []core.Component{}
	for _, comp := range c.context.Config.Components {
		if comp.Selected || comp.Required {
			c.selectedComponents = append(c.selectedComponents, comp)
		}
	}
	
	return nil
}

func (c *SSRController) handleLocationInteraction() error {
	input, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}
	
	input = strings.TrimSpace(input)
	if input != "" {
		c.installPath = input
	}
	
	return nil
}

func (c *SSRController) handleSummaryInteraction() error {
	input, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}
	
	response := strings.ToLower(strings.TrimSpace(input))
	if response != "y" && response != "yes" {
		return fmt.Errorf("installation cancelled")
	}
	
	return nil
}

func (c *SSRController) handleProgressInteraction() error {
	// Simulate installation progress
	for i := 0; i <= 100; i += 10 {
		c.viewData.Progress = i
		c.viewData.ProgressText = fmt.Sprintf("Installing components... %d%%", i)
		
		// Re-render progress view
		output, _ := c.renderer.RenderView("progress", c.viewType, c.viewData)
		if c.viewType == views.ViewCLI {
			fmt.Printf("\r%s", strings.ReplaceAll(output, "\n", ""))
		}
		
		// Simulate work
		// time.Sleep(200 * time.Millisecond)
	}
	
	if c.viewType == views.ViewCLI {
		fmt.Println() // New line after progress
	}
	
	return nil
}

func (c *SSRController) handleCompleteInteraction() error {
	if c.viewType == views.ViewCLI {
		c.reader.ReadString('\n') // Wait for Enter
	}
	return nil
}

// simulateHTMLInteraction simulates HTML/WebView interactions for demo
func (c *SSRController) simulateHTMLInteraction(viewName string) error {
	fmt.Printf("\n[WebView would handle %s interaction via JavaScript]\n", viewName)
	
	switch viewName {
	case "license":
		c.licenseAccepted = true
	case "components":
		// Keep default selection
	case "summary":
		// Proceed with installation
	}
	
	return nil
}

// Helper functions
func (c *SSRController) isComponentSelected(componentID string) bool {
	for _, comp := range c.selectedComponents {
		if comp.ID == componentID {
			return true
		}
	}
	return false
}

func getPageTitle(viewName string) string {
	titles := map[string]string{
		"welcome":    "Welcome",
		"license":    "License Agreement",
		"components": "Component Selection",
		"location":   "Installation Location",
		"summary":    "Installation Summary",
		"progress":   "Installing",
		"complete":   "Installation Complete",
	}
	
	if title, exists := titles[viewName]; exists {
		return title
	}
	return strings.Title(viewName)
}

// Shutdown cleanup
func (c *SSRController) Shutdown() error {
	return nil
}

// Interface implementations for core.UI compatibility
func (c *SSRController) ShowWelcome() error {
	return c.showView("welcome")
}

func (c *SSRController) ShowLicense(license string) (bool, error) {
	if err := c.showView("license"); err != nil {
		return false, err
	}
	if err := c.handleLicenseInteraction(); err != nil {
		return false, err
	}
	return c.licenseAccepted, nil
}

func (c *SSRController) SelectComponents(components []core.Component) ([]core.Component, error) {
	if err := c.showView("components"); err != nil {
		return nil, err
	}
	if err := c.handleComponentsInteraction(); err != nil {
		return nil, err
	}
	return c.selectedComponents, nil
}

func (c *SSRController) SelectInstallPath(defaultPath string) (string, error) {
	c.installPath = defaultPath
	if err := c.showView("location"); err != nil {
		return "", err
	}
	if err := c.handleLocationInteraction(); err != nil {
		return "", err
	}
	return c.installPath, nil
}

func (c *SSRController) ShowProgress(progress *core.Progress) error {
	c.viewData.Progress = int(progress.OverallProgress * 100)
	c.viewData.ProgressText = progress.Message
	return c.showView("progress")
}

func (c *SSRController) ShowError(err error, canRetry bool) (bool, error) {
	fmt.Printf("Error: %v\n", err)
	if canRetry {
		fmt.Print("Retry? (y/n): ")
		input, _ := c.reader.ReadString('\n')
		return strings.ToLower(strings.TrimSpace(input)) == "y", nil
	}
	return false, nil
}

func (c *SSRController) ShowSuccess(summary *core.InstallSummary) error {
	return c.showView("complete")
}

func (c *SSRController) RequestElevation(reason string) (bool, error) {
	fmt.Printf("Administrative privileges required: %s\n", reason)
	fmt.Print("Grant privileges? (y/n): ")
	input, _ := c.reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(input)) == "y", nil
}
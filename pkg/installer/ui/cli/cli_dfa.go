// Package cli provides a DFA-controlled command-line interface for the installer
package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mmso2016/setupkit/pkg/html"
	"github.com/mmso2016/setupkit/pkg/installer/controller"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// CLIDFA implements the InstallerView interface for command-line interaction
type CLIDFA struct {
	context    *core.Context
	controller *controller.InstallerController
	reader     *bufio.Reader
	renderer   *html.SSRRenderer  // For HTML export capability
}

// NewDFA creates a new DFA-controlled CLI instance
func NewDFA() *CLIDFA {
	return &CLIDFA{
		reader:   bufio.NewReader(os.Stdin),
		renderer: html.NewSSRRenderer(),
	}
}

// NewDFAWithReader creates a new DFA-controlled CLI instance with custom reader (for testing)
func NewDFAWithReader(reader *bufio.Reader) *CLIDFA {
	return &CLIDFA{
		reader:   reader,
		renderer: html.NewSSRRenderer(),
	}
}

// Initialize sets up the CLI with context
func (c *CLIDFA) Initialize(ctx *core.Context) error {
	c.context = ctx
	return nil
}

// SetController sets the DFA controller reference (called by main controller)
func (c *CLIDFA) SetController(ctrl *controller.InstallerController) {
	c.controller = ctrl
}

// Run starts the DFA-controlled installation flow
func (c *CLIDFA) Run() error {
	fmt.Println("Starting DFA-controlled CLI installation...")
	return c.controller.Start()
}

func (c *CLIDFA) Shutdown() error {
	return nil
}

// ============================================================================
// InstallerView Interface Implementation
// ============================================================================

// ShowWelcome displays the welcome screen
func (c *CLIDFA) ShowWelcome() error {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("  Welcome to %s Setup\n", c.context.Config.AppName)
	fmt.Printf("  Version: %s\n", c.context.Config.Version)
	fmt.Printf("  Publisher: %s\n", c.context.Config.Publisher)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	
	// Wait for user to proceed
	return c.waitForNext("Press Enter to continue or 'q' to quit...")
}

// ShowLicense displays license and returns acceptance
func (c *CLIDFA) ShowLicense(license string) (accepted bool, err error) {
	fmt.Println("License Agreement:")
	fmt.Println(strings.Repeat("-", 50))
	
	// Show license text (truncated for demo)
	lines := strings.Split(license, "\n")
	for i, line := range lines {
		fmt.Println(line)
		if i > 0 && (i+1)%20 == 0 {
			fmt.Println("\n... [" + strconv.Itoa(len(lines)-i-1) + " more lines] ...")
			break
		}
	}
	
	fmt.Println(strings.Repeat("-", 50))
	
	for {
		fmt.Print("Do you accept the license agreement? (y/n): ")
		input, err := c.reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		
		input = strings.TrimSpace(strings.ToLower(input))
		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no":
			fmt.Println("License not accepted. Installation cancelled.")
			return false, nil
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no.")
		}
	}
}

// ShowComponents displays component selection
func (c *CLIDFA) ShowComponents(components []core.Component) (selected []core.Component, err error) {
	fmt.Println("Select components to install:")
	fmt.Println()
	
	// Clone components to avoid modifying originals
	workingComponents := make([]core.Component, len(components))
	copy(workingComponents, components)
	
	for {
		// Display current state
		for i, comp := range workingComponents {
			status := " "
			if comp.Required {
				status = "R"
			} else if comp.Selected {
				status = "X"
			}
			
			fmt.Printf("  [%s] %d. %s (%.1f KB)\n", status, i+1, comp.Name, float64(comp.Size)/1024)
			if comp.Description != "" {
				fmt.Printf("      %s\n", comp.Description)
			}
		}
		
		fmt.Println()
		fmt.Println("  R = Required, X = Selected")
		fmt.Println()
		fmt.Print("Enter component numbers to toggle (comma-separated), or press Enter to continue: ")
		
		input, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		
		input = strings.TrimSpace(input)
		if input == "" {
			break // User pressed Enter, continue with current selection
		}
		
		// Parse component numbers
		numbers := strings.Split(input, ",")
		for _, numStr := range numbers {
			numStr = strings.TrimSpace(numStr)
			num, err := strconv.Atoi(numStr)
			if err != nil || num < 1 || num > len(workingComponents) {
				fmt.Printf("Invalid component number: %s\n", numStr)
				continue
			}
			
			idx := num - 1
			if !workingComponents[idx].Required {
				workingComponents[idx].Selected = !workingComponents[idx].Selected
			} else {
				fmt.Printf("Component %d (%s) is required and cannot be deselected.\n", num, workingComponents[idx].Name)
			}
		}
		fmt.Println()
	}
	
	// Return selected components
	var result []core.Component
	for _, comp := range workingComponents {
		if comp.Selected || comp.Required {
			result = append(result, comp)
		}
	}
	
	return result, nil
}

// ShowInstallPath allows user to select installation path
func (c *CLIDFA) ShowInstallPath(defaultPath string) (path string, err error) {
	fmt.Printf("Install location [%s]: ", defaultPath)
	
	input, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultPath, nil
	}
	
	// Expand path if needed
	if strings.HasPrefix(input, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			input = filepath.Join(homeDir, input[2:])
		}
	}
	
	return input, nil
}

// ShowSummary displays installation summary and gets confirmation
func (c *CLIDFA) ShowSummary(config *core.Config, selectedComponents []core.Component, installPath string) (proceed bool, err error) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Ready to install:")
	fmt.Printf("  Application: %s v%s\n", config.AppName, config.Version)
	fmt.Printf("  Install to: %s\n", installPath)
	fmt.Printf("  Components: %d selected\n", len(selectedComponents))
	
	// Show selected components
	fmt.Println("\nSelected components:")
	for _, comp := range selectedComponents {
		marker := ""
		if comp.Required {
			marker = " (required)"
		}
		fmt.Printf("  • %s%s\n", comp.Name, marker)
	}
	
	fmt.Println(strings.Repeat("=", 50))
	
	return c.confirm("Proceed with installation?"), nil
}

// ShowProgress displays installation progress
func (c *CLIDFA) ShowProgress(progress *core.Progress) error {
	// Simple progress display
	percentage := int(progress.OverallProgress * 100)
	fmt.Printf("\rInstalling %s... %d%% complete", progress.ComponentName, percentage)
	
	if progress.OverallProgress >= 1.0 {
		fmt.Println() // New line when complete
	}
	
	return nil
}

// ShowComplete displays installation completion
func (c *CLIDFA) ShowComplete(summary *core.InstallSummary) error {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ Installation completed successfully!")
	fmt.Printf("  Installed to: %s\n", summary.InstallPath)
	fmt.Printf("  Duration: %v\n", summary.Duration)
	fmt.Printf("  Components installed: %d\n", len(summary.ComponentsInstalled))
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	
	return nil
}

// ShowErrorMessage displays an error message (InstallerView interface)
func (c *CLIDFA) ShowErrorMessage(err error) error {
	fmt.Printf("\n❌ Error: %v\n\n", err)
	return nil
}

// OnStateChanged handles state change notifications
func (c *CLIDFA) OnStateChanged(oldState, newState wizard.State) error {
	fmt.Printf("[DEBUG] State transition: %s → %s\n", oldState, newState)
	return nil
}

// RequestElevation requests elevated privileges (core.UI interface requirement)
func (c *CLIDFA) RequestElevation(reason string) (bool, error) {
	fmt.Printf("⚠️  Administrative privileges required: %s\n", reason)
	return c.confirm("Continue with elevation request?"), nil
}

// ============================================================================
// core.UI Interface Compatibility Methods
// ============================================================================

// SelectComponents - adapts InstallerView.ShowComponents to core.UI interface
func (c *CLIDFA) SelectComponents(components []core.Component) ([]core.Component, error) {
	return c.ShowComponents(components)
}

// SelectInstallPath - adapts InstallerView.ShowInstallPath to core.UI interface  
func (c *CLIDFA) SelectInstallPath(defaultPath string) (string, error) {
	return c.ShowInstallPath(defaultPath)
}

// ShowSuccess - adapts InstallerView.ShowComplete to core.UI interface
func (c *CLIDFA) ShowSuccess(summary *core.InstallSummary) error {
	return c.ShowComplete(summary)
}

// ShowError - core.UI interface method with retry logic
func (c *CLIDFA) ShowError(err error, canRetry bool) (retry bool, errOut error) {
	fmt.Printf("\n❌ Error: %v\n\n", err)
	
	if canRetry {
		return c.confirm("Would you like to retry?"), nil
	}
	
	return false, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// waitForNext waits for user input to proceed
func (c *CLIDFA) waitForNext(message string) error {
	fmt.Print(message + " ")

	input, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "q" || input == "quit" {
		return fmt.Errorf("installation cancelled by user")
	}

	// After user confirms, advance to next state via DFA controller
	if c.controller != nil {
		go func() {
			if err := c.controller.Next(); err != nil {
				fmt.Printf("Error advancing to next state: %v\n", err)
			}
		}()
	}

	return nil
}

// confirm asks for yes/no confirmation
func (c *CLIDFA) confirm(message string) bool {
	for {
		fmt.Print(message + " (y/n): ")
		input, err := c.reader.ReadString('\n')
		if err != nil {
			return false
		}
		
		input = strings.TrimSpace(strings.ToLower(input))
		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no.")
		}
	}
}

// ExportHTMLPages exports all installer pages as HTML files (for development/testing)
func (c *CLIDFA) ExportHTMLPages(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	
	// Export welcome page
	welcomeDoc := c.renderer.RenderWelcomePage(c.context.Config)
	if err := c.writeHTMLFile(filepath.Join(outputDir, "welcome.html"), welcomeDoc.Render()); err != nil {
		return err
	}
	
	// Export components page
	componentsDoc := c.renderer.RenderComponentsPage(c.context.Config)
	if err := c.writeHTMLFile(filepath.Join(outputDir, "components.html"), componentsDoc.Render()); err != nil {
		return err
	}
	
	// Export progress page
	progressDoc := c.renderer.RenderProgressPage(c.context.Config, 50, "Installing...")
	if err := c.writeHTMLFile(filepath.Join(outputDir, "progress.html"), progressDoc.Render()); err != nil {
		return err
	}
	
	// Export completion page
	completeDoc := c.renderer.RenderCompletionPage(c.context.Config, true)
	if err := c.writeHTMLFile(filepath.Join(outputDir, "complete.html"), completeDoc.Render()); err != nil {
		return err
	}
	
	fmt.Printf("HTML pages exported to: %s\n", outputDir)
	return nil
}

// writeHTMLFile writes HTML content to a file
func (c *CLIDFA) writeHTMLFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// ============================================================================
// ExtendedInstallerView Interface Implementation
// ============================================================================

// ShowCustomState handles custom states in CLI mode
func (c *CLIDFA) ShowCustomState(stateID wizard.State, data controller.CustomStateData) (controller.CustomStateData, error) {
	fmt.Printf("\n=== Custom Configuration: %s ===\n", stateID)

	switch stateID {
	case controller.StateDBConfig:
		return c.handleDatabaseConfig(data)
	default:
		fmt.Printf("Unknown custom state: %s\n", stateID)
		return data, nil
	}
}

// handleDatabaseConfig handles database configuration in CLI mode
func (c *CLIDFA) handleDatabaseConfig(data controller.CustomStateData) (controller.CustomStateData, error) {
	fmt.Println("Database Configuration")
	fmt.Println(strings.Repeat("-", 50))

	// Get current config or use defaults
	var dbConfig *controller.DatabaseConfig
	if config, exists := data["config"]; exists {
		if existing, ok := config.(*controller.DatabaseConfig); ok {
			dbConfig = existing
		} else {
			dbConfig = controller.DefaultDatabaseConfig()
		}
	} else {
		dbConfig = controller.DefaultDatabaseConfig()
	}

	// Show current configuration
	fmt.Printf("Current configuration:\n")
	fmt.Printf("  Database Type: %s\n", dbConfig.Type)
	fmt.Printf("  Host: %s\n", dbConfig.Host)
	fmt.Printf("  Port: %d\n", dbConfig.Port)
	fmt.Printf("  Database: %s\n", dbConfig.Database)
	fmt.Printf("  Username: %s\n", dbConfig.Username)
	fmt.Printf("  SSL: %v\n", dbConfig.UseSSL)
	fmt.Println()

	// Interactive configuration
	newConfig := &controller.DatabaseConfig{
		Type:     dbConfig.Type,
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		Database: dbConfig.Database,
		Username: dbConfig.Username,
		Password: dbConfig.Password,
		UseSSL:   dbConfig.UseSSL,
	}

	// Database type selection
	supportedDBs := []string{"mysql", "postgresql", "sqlite", "sqlserver"}
	fmt.Printf("Supported database types: %s\n", strings.Join(supportedDBs, ", "))
	fmt.Printf("Database type [%s]: ", newConfig.Type)
	if input, err := c.readInput(); err == nil && strings.TrimSpace(input) != "" {
		newConfig.Type = strings.TrimSpace(input)
	}

	// Skip host/port for SQLite
	if newConfig.Type != "sqlite" {
		fmt.Printf("Host [%s]: ", newConfig.Host)
		if input, err := c.readInput(); err == nil && strings.TrimSpace(input) != "" {
			newConfig.Host = strings.TrimSpace(input)
		}

		fmt.Printf("Port [%d]: ", newConfig.Port)
		if input, err := c.readInput(); err == nil && strings.TrimSpace(input) != "" {
			if port, err := strconv.Atoi(strings.TrimSpace(input)); err == nil {
				newConfig.Port = port
			}
		}

		fmt.Printf("Username [%s]: ", newConfig.Username)
		if input, err := c.readInput(); err == nil && strings.TrimSpace(input) != "" {
			newConfig.Username = strings.TrimSpace(input)
		}

		fmt.Print("Password: ")
		if input, err := c.readInput(); err == nil {
			newConfig.Password = strings.TrimSpace(input)
		}

		fmt.Printf("Use SSL [%v]: ", newConfig.UseSSL)
		if input, err := c.readInput(); err == nil {
			input = strings.TrimSpace(strings.ToLower(input))
			if input == "y" || input == "yes" || input == "true" {
				newConfig.UseSSL = true
			} else if input == "n" || input == "no" || input == "false" {
				newConfig.UseSSL = false
			}
		}
	}

	fmt.Printf("Database name [%s]: ", newConfig.Database)
	if input, err := c.readInput(); err == nil && strings.TrimSpace(input) != "" {
		newConfig.Database = strings.TrimSpace(input)
	}

	fmt.Println()
	fmt.Printf("Final configuration: %s\n", newConfig.String())

	return controller.CustomStateData{"config": newConfig}, nil
}

// readInput reads a line of input from the reader
func (c *CLIDFA) readInput() (string, error) {
	return c.reader.ReadString('\n')
}
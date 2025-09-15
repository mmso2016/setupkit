// Package cli provides a command-line interface for the installer
package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mmso2016/setupkit/pkg/html"
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// CLI implements the command-line user interface
type CLI struct {
	context   *core.Context
	installer *core.Installer
	reader    *bufio.Reader
	renderer  *html.SSRRenderer  // For HTML export capability
}

// New creates a new CLI instance
func New() *CLI {
	return &CLI{
		reader:   bufio.NewReader(os.Stdin),
		renderer: html.NewSSRRenderer(),
	}
}

func (c *CLI) Initialize(ctx *core.Context) error {
	c.context = ctx
	// Store installer reference for later use
	if installer, ok := ctx.Metadata["installer"].(*core.Installer); ok {
		c.installer = installer
	}
	return nil
}

func (c *CLI) Run() error {
	// Welcome
	if err := c.ShowWelcome(); err != nil {
		return err
	}

	// License
	if c.context.Config.License != "" {
		accepted, err := c.ShowLicense(c.context.Config.License)
		if err != nil {
			return err
		}
		if !accepted {
			fmt.Println("\nLicense not accepted. Installation cancelled.")
			return fmt.Errorf("license not accepted")
		}
	}

	// Component selection
	components, err := c.SelectComponents(c.context.Config.Components)
	if err != nil {
		return err
	}
	c.installer.SetSelectedComponents(components)

	// Install path
	defaultPath := c.context.Config.InstallDir
	if defaultPath == "" {
		defaultPath = filepath.Join("/opt", c.context.Config.AppName)
	}

	path, err := c.SelectInstallPath(defaultPath)
	if err != nil {
		return err
	}
	c.installer.SetInstallPath(path)

	// Confirm installation
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Ready to install:")
	fmt.Printf("  Application: %s v%s\n", c.context.Config.AppName, c.context.Config.Version)
	fmt.Printf("  Install to: %s\n", path)
	fmt.Printf("  Components: %d selected\n", len(components))
	fmt.Println(strings.Repeat("=", 50))

	if !c.confirm("Proceed with installation?") {
		fmt.Println("Installation cancelled.")
		return fmt.Errorf("installation cancelled by user")
	}

	// Execute installation
	fmt.Println("\nStarting installation...")
	c.installer.SetUI(c) // Set the UI reference on the installer
	if err := c.installer.ExecuteInstallation(); err != nil {
		return err
	}

	// Show success
	summary := c.installer.CreateSummary()
	return c.ShowSuccess(summary)
}

func (c *CLI) Shutdown() error {
	return nil
}

func (c *CLI) ShowWelcome() error {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("  Welcome to %s Setup\n", c.context.Config.AppName)
	fmt.Printf("  Version: %s\n", c.context.Config.Version)
	if c.context.Config.Publisher != "" {
		fmt.Printf("  Publisher: %s\n", c.context.Config.Publisher)
	}
	fmt.Println(strings.Repeat("=", 50) + "\n")
	return nil
}

func (c *CLI) ShowLicense(license string) (bool, error) {
	fmt.Println("License Agreement:")
	fmt.Println(strings.Repeat("-", 50))

	// Show license text (truncated for CLI)
	lines := strings.Split(license, "\n")
	maxLines := 20
	if len(lines) > maxLines {
		for i := 0; i < maxLines; i++ {
			fmt.Println(lines[i])
		}
		fmt.Printf("\n... [%d more lines] ...\n", len(lines)-maxLines)
	} else {
		fmt.Println(license)
	}

	fmt.Println(strings.Repeat("-", 50))
	return c.confirm("Do you accept the license agreement?"), nil
}

func (c *CLI) SelectComponents(components []core.Component) ([]core.Component, error) {
	fmt.Println("Select components to install:")
	fmt.Println()

	// Show component list
	for i, comp := range components {
		status := " "
		if comp.Required {
			status = "R" // Required
		} else if comp.Selected {
			status = "X" // Selected
		}

		fmt.Printf("  [%s] %d. %s", status, i+1, comp.Name)
		if comp.Size > 0 {
			fmt.Printf(" (%s)", formatSize(comp.Size))
		}
		fmt.Println()

		if comp.Description != "" {
			fmt.Printf("      %s\n", comp.Description)
		}
	}

	fmt.Println("\n  R = Required, X = Selected")
	fmt.Println("\nEnter component numbers to toggle (comma-separated), or press Enter to continue:")

	input, _ := c.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Parse selections
	if input != "" {
		selections := strings.Split(input, ",")
		for _, s := range selections {
			s = strings.TrimSpace(s)
			var idx int
			if _, err := fmt.Sscanf(s, "%d", &idx); err == nil {
				idx-- // Convert to 0-based
				if idx >= 0 && idx < len(components) {
					if !components[idx].Required {
						components[idx].Selected = !components[idx].Selected
					}
				}
			}
		}
	}

	// Return selected components
	var selected []core.Component
	for _, comp := range components {
		if comp.Selected || comp.Required {
			selected = append(selected, comp)
		}
	}

	return selected, nil
}

func (c *CLI) SelectInstallPath(defaultPath string) (string, error) {
	fmt.Printf("\nInstall location [%s]: ", defaultPath)

	input, _ := c.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultPath, nil
	}

	// Expand home directory if needed
	if strings.HasPrefix(input, "~") {
		home, _ := os.UserHomeDir()
		input = filepath.Join(home, input[1:])
	}

	return filepath.Clean(input), nil
}

func (c *CLI) ShowProgress(progress *core.Progress) error {
	// Simple progress display
	barWidth := 40
	filled := int(progress.OverallProgress * float64(barWidth))

	fmt.Printf("\r[%s%s] %.0f%% - %s",
		strings.Repeat("=", filled),
		strings.Repeat(" ", barWidth-filled),
		progress.OverallProgress*100,
		progress.Message)

	if progress.OverallProgress >= 1.0 {
		fmt.Println() // New line when complete
	}

	return nil
}

func (c *CLI) ShowError(err error, canRetry bool) (bool, error) {
	fmt.Printf("\n\nError: %v\n", err)

	if canRetry {
		return c.confirm("Would you like to retry?"), nil
	}

	return false, nil
}

func (c *CLI) ShowSuccess(summary *core.InstallSummary) error {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("  Installation Completed Successfully!")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("\nInstalled to: %s\n", summary.InstallPath)
	fmt.Printf("Duration: %v\n", summary.Duration.Round(1))

	if len(summary.ComponentsInstalled) > 0 {
		fmt.Println("\nInstalled components:")
		for _, comp := range summary.ComponentsInstalled {
			fmt.Printf("  - %s\n", comp)
		}
	}

	if len(summary.NextSteps) > 0 {
		fmt.Println("\nNext steps:")
		for _, step := range summary.NextSteps {
			fmt.Printf("  • %s\n", step)
		}
	}

	fmt.Println()
	return nil
}

func (c *CLI) RequestElevation(reason string) (bool, error) {
	fmt.Printf("\nAdministrative privileges required: %s\n", reason)
	return c.confirm("Grant administrative privileges?"), nil
}

// Helper functions

func (c *CLI) confirm(prompt string) bool {
	fmt.Printf("%s (y/n): ", prompt)

	input, _ := c.reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	return input == "y" || input == "yes"
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ExportHTMLPages exports all installer pages as HTML files for debugging/preview
func (c *CLI) ExportHTMLPages(outputDir string) error {
	if c.context == nil {
		return fmt.Errorf("CLI not initialized")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("Exporting HTML pages to: %s\n", outputDir)

	// Export welcome page
	welcomeDoc := c.renderer.RenderWelcomePage(c.context.Config)
	welcomePath := filepath.Join(outputDir, "cli-welcome.html")
	if err := c.writeHTMLFile(welcomePath, welcomeDoc.Render()); err != nil {
		return fmt.Errorf("failed to export welcome page: %w", err)
	}
	fmt.Printf("  ✓ cli-welcome.html\n")

	// Export components page
	componentsDoc := c.renderer.RenderComponentsPage(c.context.Config)
	componentsPath := filepath.Join(outputDir, "cli-components.html")
	if err := c.writeHTMLFile(componentsPath, componentsDoc.Render()); err != nil {
		return fmt.Errorf("failed to export components page: %w", err)
	}
	fmt.Printf("  ✓ cli-components.html\n")

	// Export progress page (75% as example)
	progressDoc := c.renderer.RenderProgressPage(c.context.Config, 75, "Installing components...")
	progressPath := filepath.Join(outputDir, "cli-progress.html")
	if err := c.writeHTMLFile(progressPath, progressDoc.Render()); err != nil {
		return fmt.Errorf("failed to export progress page: %w", err)
	}
	fmt.Printf("  ✓ cli-progress.html\n")

	// Export complete page
	completeDoc := c.renderer.RenderCompletionPage(c.context.Config, true)
	completePath := filepath.Join(outputDir, "cli-complete.html")
	if err := c.writeHTMLFile(completePath, completeDoc.Render()); err != nil {
		return fmt.Errorf("failed to export complete page: %w", err)
	}
	fmt.Printf("  ✓ cli-complete.html\n")

	fmt.Printf("\nHTML export completed successfully!\n")
	return nil
}

// writeHTMLFile writes HTML content to a file
func (c *CLI) writeHTMLFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

package main

import (
	"log"

	"github.com/mmso2016/setupkit/installer"
	"github.com/mmso2016/setupkit/installer/core"
	// Import UI factory
	_ "github.com/mmso2016/setupkit/installer/ui"
)

func main() {
	// Simple installer example using the new DFA-based framework
	// This version runs in silent mode to show automated DFA flow
	app, err := installer.New(
		// Basic application information
		installer.WithAppName("My Application"),
		installer.WithVersion("1.0.0"),
		installer.WithPublisher("Example Corp"),
		installer.WithWebsite("https://example.com"),
		installer.WithLicense("MIT License - Auto-accepted for demo"),

		// Components to install
		installer.WithComponents(
			core.Component{
				ID:       "core",
				Name:     "Core Application",
				Size:     45 * 1024 * 1024, // 45 MB
				Required: true,
				Selected: true,
			},
			core.Component{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and API documentation",
				Size:        12 * 1024 * 1024, // 12 MB
				Selected:    true,
			},
		),

		// Enable silent mode with DFA wizard for automated demo
		installer.WithMode(installer.ModeSilent),
		installer.WithDFAWizard(),
		installer.WithInstallDir("C:\\temp\\MyApp"), // Test install directory
		installer.WithDryRun(true), // Dry run to avoid actual file operations
	)

	if err != nil {
		log.Fatal(err)
	}

	// Show DFA wizard status
	if app.IsUsingDFAWizard() {
		log.Println("✅ DFA Wizard enabled - No hardcoded JavaScript navigation!")
	}

	// Run the installer with DFA wizard
	log.Println("Running installer with DFA wizard system...")
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("✅ DFA-based installation completed successfully!")
}
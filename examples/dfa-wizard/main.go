// Package main demonstrates the DFA-based wizard system
package main

import (
	"embed"
	"log"

	"github.com/mmso2016/setupkit/installer"
	"github.com/mmso2016/setupkit/installer/core"
)

//go:embed assets/*
var assets embed.FS

func main() {
	// Example 1: Standard DFA Wizard (Express Mode)
	log.Println("Example 1: Standard Express Wizard")
	runStandardExpressWizard()
	
	// Example 2: Extended Wizard with Theme Selection
	log.Println("\nExample 2: Extended Wizard with Theme Selection")
	runExtendedWizardWithThemes()
	
	// Example 3: Custom DFA Wizard with Advanced Options
	log.Println("\nExample 3: Advanced Custom Wizard")
	runAdvancedCustomWizard()
}

// runStandardExpressWizard demonstrates the standard express wizard
func runStandardExpressWizard() {
	// Register built-in providers
	registerBuiltinProviders()
	
	components := []installer.Component{
		{
			ID:          "core",
			Name:        "Core Application",
			Description: "Essential application files",
			Required:    true,
			Selected:    true,
			Size:        50 * 1024 * 1024, // 50MB
		},
		{
			ID:          "docs",
			Name:        "Documentation",
			Description: "User manual and help files",
			Required:    false,
			Selected:    false,
			Size:        10 * 1024 * 1024, // 10MB
		},
	}
	
	app, err := installer.New(
		installer.WithAppName("DFA Wizard Demo"),
		installer.WithVersion("1.0.0"),
		installer.WithPublisher("SetupKit Demo"),
		installer.WithWebsite("https://github.com/setupkit/setupkit"),
		installer.WithComponents(components...),
		installer.WithInstallDir("C:\\Program Files\\DFA Wizard Demo"),
		installer.WithAssets(assets),
		installer.WithLicense("This is a demo license agreement."),
		
		// Enable DFA-based wizard (express mode)
		installer.WithDFAWizard(),
		
		installer.WithDryRun(true), // Demo mode
		installer.WithVerbose(true),
	)
	
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}
	
	log.Printf("Running express wizard (DFA-based)...")
	if err := app.Run(); err != nil {
		log.Printf("Installation failed: %v", err)
	} else {
		log.Printf("Installation completed successfully!")
	}
}

// runExtendedWizardWithThemes demonstrates extended wizard with theme selection
func runExtendedWizardWithThemes() {
	registerBuiltinProviders()
	
	components := []installer.Component{
		{
			ID:          "core",
			Name:        "Core Application",
			Description: "Essential application files",
			Required:    true,
			Selected:    true,
		},
		{
			ID:          "themes",
			Name:        "Additional Themes",
			Description: "Extra visual themes",
			Required:    false,
			Selected:    true,
		},
		{
			ID:          "examples",
			Name:        "Examples",
			Description: "Sample projects and tutorials",
			Required:    false,
			Selected:    false,
		},
	}
	
	// Available themes for selection
	availableThemes := []string{"default", "dark", "corporate", "modern"}
	
	app, err := installer.New(
		installer.WithAppName("ThemeWiz"),
		installer.WithVersion("2.0.0"),
		installer.WithPublisher("SetupKit Themes"),
		installer.WithComponents(components...),
		installer.WithInstallDir("C:\\Program Files\\ThemeWiz"),
		installer.WithAssets(assets),
		installer.WithLicense("Extended demo license with theme support."),
		
		// Enable extended wizard with theme selection
		installer.WithExtendedWizard(availableThemes, "default"),
		
		installer.WithDryRun(true),
		installer.WithVerbose(true),
	)
	
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}
	
	log.Printf("Running extended wizard with theme selection...")
	if err := app.Run(); err != nil {
		log.Printf("Installation failed: %v", err)
	} else {
		log.Printf("Installation completed successfully!")
		
		// Access wizard data
		if app.IsUsingDFAWizard() {
			adapter := app.GetWizardAdapter()
			if adapter != nil {
				wizardData := adapter.GetWizardData()
				if selectedTheme, ok := wizardData["selected_theme"]; ok {
					log.Printf("User selected theme: %v", selectedTheme)
				}
			}
		}
	}
}

// runAdvancedCustomWizard demonstrates advanced custom wizard
func runAdvancedCustomWizard() {
	registerBuiltinProviders()
	
	components := []installer.Component{
		{
			ID:          "core",
			Name:        "Core Application",
			Description: "Essential application files",
			Required:    true,
			Selected:    true,
		},
		{
			ID:          "server",
			Name:        "Server Components",
			Description: "Backend server files",
			Required:    false,
			Selected:    true,
		},
		{
			ID:          "client",
			Name:        "Client Tools",
			Description: "Command-line utilities",
			Required:    false,
			Selected:    true,
		},
		{
			ID:          "dev_tools",
			Name:        "Development Tools",
			Description: "SDK and development utilities",
			Required:    false,
			Selected:    false,
		},
	}
	
	app, err := installer.New(
		installer.WithAppName("Advanced Setup"),
		installer.WithVersion("3.0.0"),
		installer.WithPublisher("SetupKit Advanced"),
		installer.WithComponents(components...),
		installer.WithInstallDir("C:\\Program Files\\Advanced Setup"),
		installer.WithAssets(assets),
		installer.WithLicense("Advanced license agreement with full customization."),
		
		// Enable advanced wizard with mode selection
		installer.WithAdvancedDFAWizard(),
		
		installer.WithDryRun(true),
		installer.WithVerbose(true),
	)
	
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}
	
	log.Printf("Running advanced custom wizard...")
	if err := app.Run(); err != nil {
		log.Printf("Installation failed: %v", err)
	} else {
		log.Printf("Installation completed successfully!")
		
		// Show wizard history
		if app.IsUsingDFAWizard() {
			adapter := app.GetWizardAdapter()
			if adapter != nil {
				history := adapter.GetStateHistory()
				log.Printf("Wizard traversed states: %v", history)
				
				if adapter.GetDryRunLog() != nil {
					log.Printf("Dry-run log entries: %d", len(adapter.GetDryRunLog()))
				}
			}
		}
	}
}

// registerBuiltinProviders registers the built-in wizard providers
func registerBuiltinProviders() {
	// Register standard providers
	core.RegisterWizardProvider("standard-express", core.NewStandardWizardProvider(core.ModeExpress))
	core.RegisterWizardProvider("standard-custom", core.NewStandardWizardProvider(core.ModeCustom))
	core.RegisterWizardProvider("standard-advanced", core.NewStandardWizardProvider(core.ModeAdvanced))
	
	// Register extended provider
	extendedProvider := core.NewExtendedWizardProvider(core.ModeCustom)
	core.RegisterWizardProvider("extended", extendedProvider)
	
	// Set default
	core.SetDefaultWizardProvider("standard-custom")
	
	log.Printf("Registered DFA wizard providers: standard-express, standard-custom, standard-advanced, extended")
}
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mmso2016/setupkit/pkg/installer/controller"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/installer/ui"
	"github.com/mmso2016/setupkit/pkg/installer/ui/cli"
)

func main() {
	// Command line flags
	var (
		mode       = flag.String("mode", "auto", "UI mode: gui, browser, cli, silent, auto")
		installDir = flag.String("dir", "", "Installation directory (overrides default temp dir)")
		help       = flag.Bool("help", false, "Show this help message")
	)
	flag.Parse()

	// Show help and exit
	if *help {
		fmt.Println("=== Custom State Demo: Database Configuration ===")
		fmt.Println("This demo shows how to add a database configuration step")
		fmt.Println("to the installer flow using custom states.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Flow: Welcome → License → Components → Install Path → DB Config → Summary → Progress → Complete")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Printf("  %s -mode=gui          # Native WebView GUI mode\n", os.Args[0])
		fmt.Printf("  %s -mode=browser      # Browser-based interface\n", os.Args[0])
		fmt.Printf("  %s -mode=cli          # Interactive CLI mode\n", os.Args[0])
		fmt.Printf("  %s -mode=silent       # Automated silent mode\n", os.Args[0])
		fmt.Printf("  %s -dir=\"./install\"   # Custom install directory\n", os.Args[0])
		return
	}

	// Create a demo configuration for an app that needs database configuration
	tempDir, err := os.MkdirTemp("", "custom_state_demo")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Determine install directory
	installPath := filepath.Join(tempDir, "install")
	if *installDir != "" {
		installPath = *installDir
	}

	config := &core.Config{
		AppName:       "DatabaseApp",
		Version:       "1.0.0",
		Publisher:     "Custom State Demo",
		InstallDir:    installPath,
		AcceptLicense: true,
		License:       "Demo License Agreement - Database Application",
		Components: []core.Component{
			{ID: "app", Name: "Application Core", Required: true, Selected: true, Size: 1024000},
			{ID: "db-schema", Name: "Database Schema", Required: false, Selected: true, Size: 512000},
			{ID: "sample-data", Name: "Sample Data", Required: false, Selected: false, Size: 2048000},
		},
	}

	// Setup logger
	logger := core.NewLogger("info", "")

	// Create context
	context := &core.Context{
		Config:   config,
		Logger:   logger,
		Metadata: make(map[string]interface{}),
	}

	// Create installer
	installer := core.New(config)
	installer.SetContext(context)  // Set the context on the installer

	// Set install handler for demo (creates dummy files)
	installer.SetInstallHandler(func(installPath string, components []core.Component) error {
		for _, comp := range components {
			// Create dummy file for each component
			compFile := filepath.Join(installPath, comp.ID+".txt")
			content := fmt.Sprintf("Component: %s\nName: %s\nInstalled at: %s\n",
				comp.ID, comp.Name, time.Now().Format(time.RFC3339))
			if err := os.WriteFile(compFile, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to install component %s: %w", comp.ID, err)
			}
		}
		return nil
	})

	context.Metadata["installer"] = installer

	// Create DFA controller
	dfaController := controller.NewInstallerController(config, installer)

	// Register the database configuration custom state
	dbConfigHandler := controller.NewDatabaseConfigHandler()
	if err := dfaController.RegisterCustomState(dbConfigHandler); err != nil {
		log.Fatalf("Failed to register database config state: %v", err)
	}

	fmt.Println("=== Custom State Demo: Database Configuration ===")
	fmt.Println("This demo shows how to add a database configuration step")
	fmt.Println("to the installer flow using custom states.")
	fmt.Println()
	fmt.Printf("Mode: %s | Install Dir: %s\n", *mode, installPath)
	fmt.Println("Flow: Welcome → License → Components → Install Path → DB Config → Summary → Progress → Complete")
	fmt.Println()

	// Create UI based on mode parameter
	var installerView controller.InstallerView
	switch strings.ToLower(*mode) {
	case "cli":
		fmt.Println("Starting interactive CLI mode...")
		cliUI := cli.NewDFA()
		if err := cliUI.Initialize(context); err != nil {
			log.Fatalf("Failed to initialize CLI UI: %v", err)
		}
		cliUI.SetController(dfaController)
		installerView = cliUI

	case "gui":
		fmt.Println("Starting native WebView GUI mode...")
		guiUI := ui.NewWebViewGUI()
		if guiInitializer, ok := guiUI.(interface{ Initialize(*core.Context) error }); ok {
			if err := guiInitializer.Initialize(context); err != nil {
				log.Fatalf("Failed to initialize GUI UI: %v", err)
			}
		}
		if setController, ok := guiUI.(interface{ SetController(*controller.InstallerController) }); ok {
			setController.SetController(dfaController)
		}
		installerView = guiUI

	case "browser":
		fmt.Println("Starting browser mode...")
		browserUI := ui.NewGUIDFA()
		if browserInitializer, ok := browserUI.(interface{ Initialize(*core.Context) error }); ok {
			if err := browserInitializer.Initialize(context); err != nil {
				log.Fatalf("Failed to initialize browser UI: %v", err)
			}
		}
		if setController, ok := browserUI.(interface{ SetController(*controller.InstallerController) }); ok {
			setController.SetController(dfaController)
		}
		installerView = browserUI

	case "silent":
		fmt.Println("Starting silent mode...")
		fmt.Println("This will automatically navigate through all states.")
		silentUI := ui.NewSilentUIDFA()
		if err := silentUI.Initialize(context); err != nil {
			log.Fatalf("Failed to initialize Silent UI: %v", err)
		}
		installerView = silentUI

	case "auto":
		// Auto-detect best mode (default to CLI for demo)
		fmt.Println("Auto mode: selecting CLI...")
		cliUI := cli.NewDFA()
		if err := cliUI.Initialize(context); err != nil {
			log.Fatalf("Failed to initialize CLI UI: %v", err)
		}
		cliUI.SetController(dfaController)
		installerView = cliUI

	default:
		log.Fatalf("Unknown mode: %s. Use gui, browser, cli, silent, or auto", *mode)
	}

	// Set the UI view on the DFA controller
	dfaController.SetView(installerView)

	// Handle different UI modes
	if strings.ToLower(*mode) == "gui" || strings.ToLower(*mode) == "browser" {
		// For GUI/browser modes, use the UI's Run() method which handles the interface and waits
		modeName := "GUI"
		if strings.ToLower(*mode) == "browser" {
			modeName = "browser"
		}
		fmt.Printf("Starting %s installation...\n", modeName)
		if err := installerView.(interface{ Run() error }).Run(); err != nil {
			log.Fatalf("%s installation failed: %v", modeName, err)
		}
	} else {
		// Start the DFA controller for CLI/Silent modes
		if err := dfaController.Start(); err != nil {
			log.Fatalf("Installation failed: %v", err)
		}

		// For silent mode ONLY, automatically navigate through states for demo purposes
		if strings.ToLower(*mode) == "silent" {
			fmt.Println("\n--- Automatically navigating through installer states ---")
			states := []string{"License", "Components", "Install Path", "Database Configuration", "Summary", "Progress", "Complete"}
			for i, stateName := range states {
				fmt.Printf("\n[%d/%d] Proceeding to: %s\n", i+1, len(states), stateName)

				if err := dfaController.Next(); err != nil {
					log.Printf("Navigation to %s failed: %v", stateName, err)
					break
				}

				// Longer delay for Progress state to let installation complete
				if stateName == "Progress" {
					fmt.Println("Installation is running...")
					time.Sleep(2 * time.Second) // Give time for actual installation
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}
			fmt.Println("\n--- Installation flow demonstration completed ---")
		} else {
			// For CLI modes, the UI will handle user interaction
			// The DFA controller Start() method will manage the flow
			fmt.Println("Installation started. Use the UI to navigate through the states.")
		}
	}

	// Demo finished
	fmt.Println("\n=== Custom State Demo Completed ===")

	// Show the final state data
	stateData := dfaController.GetStateData()
	if dbConfig, exists := stateData["database_config"]; exists {
		if config, ok := dbConfig.(*controller.DatabaseConfig); ok {
			fmt.Printf("Final database configuration: %s\n", config.String())
			fmt.Printf("Connection string: %s\n", config.GetConnectionString())
		}
	}
}
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mmso2016/setupkit/installer"
)

func main() {
	// Command line flags
	var (
		mode         = flag.String("mode", "auto", "UI mode: auto, cli, gui, silent")
		title        = flag.String("title", "My Application", "Application title")
		responseFile = flag.String("response", "", "Response file for silent installation")
		logFile      = flag.String("log", "", "Log file for installation")
		verbose      = flag.Bool("v", false, "Verbose output")
		// Theme and UI configuration options
		configFile = flag.String("config", "", "UI configuration file (YAML)")
		themeName  = flag.String("theme", "default", "Theme name (default, corporate-blue, medical-green, tech-dark, minimal-light)")
		genConfig  = flag.String("generate-config", "", "Generate default config file and exit")
		listThemes = flag.Bool("list-themes", false, "List available themes and exit")
		_          = flag.Bool("demo", false, "Run in demo mode")
	)

	flag.Parse()

	// Handle utility commands
	if *listThemes {
		listAvailableThemes()
		return
	}

	if *genConfig != "" {
		generateConfig(*genConfig, *title)
		return
	}

	// Parse UI mode
	var uiMode installer.Mode
	switch *mode {
	case "cli":
		uiMode = installer.ModeCLI
	case "gui":
		uiMode = installer.ModeGUI
	case "silent":
		uiMode = installer.ModeSilent
	default:
		uiMode = installer.ModeAuto
	}

	// Get default installation directory
	installDir := getDefaultInstallDir(*title)

	// Build installer options
	opts := []installer.Option{
		installer.WithAppName(*title),
		installer.WithVersion("1.0.0"),
		installer.WithPublisher("Example Publisher"),
		installer.WithMode(uiMode),
		installer.WithInstallDir(installDir),
		installer.WithVerbose(*verbose),
		installer.WithLicense(getLicenseText()),
		installer.WithComponents(
			installer.Component{
				ID:          "core",
				Name:        "Core Files",
				Description: "Required application files",
				Required:    true,
				Selected:    true,
				Size:        10485760, // 10 MB
				Installer:   installCore,
			},
			installer.Component{
				ID:          "docs",
				Name:        "Documentation",
				Description: "User manual and help files",
				Required:    false,
				Selected:    true,
				Size:        5242880, // 5 MB
				Installer:   installDocs,
			},
			installer.Component{
				ID:          "examples",
				Name:        "Examples",
				Description: "Sample projects and templates",
				Required:    false,
				Selected:    false,
				Size:        2097152, // 2 MB
				Installer:   installExamples,
			},
		),
	}

	// Add response file support
	if *responseFile != "" {
		opts = append(opts, installer.WithResponseFile(*responseFile))
	}

	// Add log file support
	if *logFile != "" {
		// TODO: Add log file option to installer package
		log.Printf("Log file: %s", *logFile)
	}

	// Apply UI configuration if specified
	if *configFile != "" {
		opts = append(opts, installer.WithUIConfig(*configFile))
		log.Printf("Using UI configuration: %s", *configFile)
	} else if *themeName != "" && *themeName != "default" {
		// Apply theme directly if no config file and not default
		opts = append(opts, installer.WithTheme(*themeName))
		log.Printf("Using theme: %s", *themeName)
	} else {
		// Use appropriate default theme based on mode
		defaultTheme := getDefaultThemeForMode(uiMode)
		opts = append(opts, installer.WithTheme(defaultTheme))
		log.Printf("Using default theme: %s", defaultTheme)
	}

	// Create installer
	inst, err := installer.New(opts...)
	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}

	// Run installer
	ctx := context.Background()
	if err := inst.RunWithContext(ctx); err != nil {
		log.Fatalf("Installation failed: %v", err)
	}

	fmt.Println("Installation completed successfully!")
}

func getDefaultInstallDir(appName string) string {
	// Remove spaces and special characters for directory name
	dirName := appName

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("PROGRAMFILES"), dirName)
	case "darwin":
		return filepath.Join("/Applications", dirName+".app")
	default:
		return filepath.Join("/opt", dirName)
	}
}

// Component installation functions
func installCore(ctx context.Context) error {
	log.Println("Installing core files...")
	// TODO: Implement actual core installation
	return nil
}

func installDocs(ctx context.Context) error {
	log.Println("Installing documentation...")
	// TODO: Implement actual docs installation
	return nil
}

func installExamples(ctx context.Context) error {
	log.Println("Installing examples...")
	// TODO: Implement actual examples installation
	return nil
}

// Theme and configuration utilities
func listAvailableThemes() {
	log.Println("Available themes:")
	themes := installer.ListThemes()
	for _, theme := range themes {
		log.Printf("  %-20s - %s", theme.Name, theme.Description)
	}
}

func generateConfig(filename, appName string) {
	log.Printf("Generating default configuration: %s", filename)
	if err := installer.GenerateConfigFile(filename, appName); err != nil {
		log.Fatal("Failed to generate config:", err)
	}
	log.Println("Configuration file generated successfully!")
	log.Println("You can now customize it and use with: -config", filename)
}

func getDefaultThemeForMode(mode installer.Mode) string {
	switch mode {
	case installer.ModeGUI:
		return "default" // Modern GUI theme
	case installer.ModeCLI:
		return "minimal-light" // Simple for CLI
	case installer.ModeSilent:
		return "minimal-light" // Minimal for silent
	default:
		return "default" // Auto mode default
	}
}

func getLicenseText() string {
	return `MIT License

Copyright (c) 2024 Example Publisher

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`
}

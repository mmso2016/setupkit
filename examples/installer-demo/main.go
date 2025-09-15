// DemoApp Installer - Complete installer with embedded YAML configuration and assets
package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mmso2016/setupkit/pkg/installer/controller"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/installer/ui"
	"github.com/mmso2016/setupkit/pkg/installer/ui/cli"
	"gopkg.in/yaml.v3"
)

//go:embed installer.yml
var embeddedConfig []byte

//go:embed assets/*
var embeddedAssets embed.FS

// ConsoleLogger is a simple console logger implementation
type ConsoleLogger struct {
	verbose bool
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{verbose: false}
}

func (l *ConsoleLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.verbose {
		fmt.Printf("[DEBUG] %s", msg)
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
			}
		}
		fmt.Println()
	}
}

func (l *ConsoleLogger) Info(msg string, keysAndValues ...interface{}) {
	fmt.Printf("[INFO] %s", msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
		}
	}
	fmt.Println()
}

func (l *ConsoleLogger) Warn(msg string, keysAndValues ...interface{}) {
	fmt.Printf("[WARN] %s", msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
		}
	}
	fmt.Println()
}

func (l *ConsoleLogger) Error(msg string, keysAndValues ...interface{}) {
	fmt.Printf("[ERROR] %s", msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
		}
	}
	fmt.Println()
}

func (l *ConsoleLogger) Verbose(msg string, keysAndValues ...interface{}) {
	if l.verbose {
		fmt.Printf("[VERBOSE] %s", msg)
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
			}
		}
		fmt.Println()
	}
}

func (l *ConsoleLogger) VerboseSection(section string) {
	if l.verbose {
		fmt.Printf("[VERBOSE] === %s ===\n", section)
	}
}

func (l *ConsoleLogger) SetVerbose(verbose bool) {
	l.verbose = verbose
}

func (l *ConsoleLogger) Close() error {
	return nil
}

// YAML configuration structures
type InstallerConfig struct {
	AppName       string           `yaml:"app_name"`
	Version       string           `yaml:"version"`
	Publisher     string           `yaml:"publisher"`
	Website       string           `yaml:"website"`
	Mode          string           `yaml:"mode"`
	Unattended    bool             `yaml:"unattended"`
	AcceptLicense bool             `yaml:"accept_license"`
	InstallDir    string           `yaml:"install_dir"`
	License       string           `yaml:"license"`
	Components    []ComponentYAML  `yaml:"components"`
	Settings      SettingsYAML     `yaml:"settings"`
	Profiles      map[string]ProfileYAML `yaml:"profiles"`
}

type ComponentYAML struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Required    bool     `yaml:"required"`
	Selected    bool     `yaml:"selected"`
	Files       []string `yaml:"files"`
}

type SettingsYAML struct {
	CreateShortcuts     bool `yaml:"create_shortcuts"`
	AddToPath          bool `yaml:"add_to_path"`
	BackupExisting     bool `yaml:"backup_existing"`
	VerifyInstallation bool `yaml:"verify_installation"`
}

type ProfileYAML struct {
	Description     string   `yaml:"description"`
	Components      []string `yaml:"components"`
	CreateShortcuts bool     `yaml:"create_shortcuts"`
	AddToPath       bool     `yaml:"add_to_path"`
}

func main() {
	// Command line flags
	var (
		mode         = flag.String("mode", "", "UI mode: gui, browser, cli, auto (overrides config)")
		installDir   = flag.String("dir", "", "Installation directory (overrides config)")
		silent       = flag.Bool("silent", false, "Silent installation")
		configFile   = flag.String("config", "", "YAML configuration file (if not specified, uses embedded config)")
		profile      = flag.String("profile", "", "Installation profile: minimal, full, developer")
		unattended   = flag.Bool("unattended", false, "Unattended installation (auto-accept license)")
		listProfiles = flag.Bool("list-profiles", false, "List available installation profiles")
	)
	flag.Parse()

	fmt.Printf("DemoApp Installer\n")
	fmt.Printf("Built with SetupKit Framework\n\n")

	// Load YAML configuration
	yamlConfig, err := loadYAMLConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Handle profile listing
	if *listProfiles {
		listInstallationProfiles(yamlConfig)
		return
	}

	// Apply profile if specified
	if *profile != "" {
		if err := applyProfile(yamlConfig, *profile); err != nil {
			log.Fatalf("Failed to apply profile '%s': %v", *profile, err)
		}
		fmt.Printf("Applied installation profile: %s\n", *profile)
	}

	// Override YAML config with command-line flags
	if *mode != "" {
		yamlConfig.Mode = *mode
	}
	if *installDir != "" {
		yamlConfig.InstallDir = *installDir
	}
	if *unattended {
		yamlConfig.Unattended = true
		yamlConfig.AcceptLicense = true
	}
	if *silent {
		yamlConfig.Mode = "silent"
		yamlConfig.Unattended = true
		yamlConfig.AcceptLicense = true
	}

	// Create installer configuration from YAML
	config := createConfigFromYAML(yamlConfig)
	
	fmt.Printf("Installing: %s v%s\n", config.AppName, config.Version)
	fmt.Printf("Publisher: %s\n", config.Publisher)
	if yamlConfig.Unattended {
		fmt.Printf("Mode: Unattended installation\n")
	} else {
		fmt.Printf("Mode: %s\n", yamlConfig.Mode)
	}
	fmt.Printf("Target: %s\n\n", config.InstallDir)

	// Create installer context
	ctx := &core.Context{
		Config:   config,
		Logger:   NewConsoleLogger(),
		Metadata: make(map[string]interface{}),
	}

	// Create and configure installer
	installer := core.New(config)
	installer.SetContext(ctx)
	installer.SetInstallHandler(func(installPath string, components []core.Component) error {
		return copyInstallationFiles(installPath, components)
	})
	ctx.Metadata["installer"] = installer

	// Create DFA controller - ALL UI modes use the same DFA approach
	dfaController := controller.NewInstallerController(config, installer)

	// Determine UI mode
	uiMode := determineUIMode(yamlConfig.Mode, yamlConfig.Unattended)

	// Create DFA-controlled UI based on mode
	fmt.Printf("Starting installation with %s interface...\n", getModeName(uiMode))
	var installerView controller.InstallerView
	switch uiMode {
	case core.ModeCLI:
		cliUI := cli.NewDFA()
		if err := cliUI.Initialize(ctx); err != nil {
			log.Fatalf("Failed to initialize CLI UI: %v", err)
		}
		cliUI.SetController(dfaController)
		installerView = cliUI

	case core.ModeGUI:
		guiUI := ui.NewWebViewGUI()
		if guiInitializer, ok := guiUI.(interface{ Initialize(*core.Context) error }); ok {
			if err := guiInitializer.Initialize(ctx); err != nil {
				log.Fatalf("Failed to initialize GUI UI: %v", err)
			}
		}
		if setController, ok := guiUI.(interface{ SetController(*controller.InstallerController) }); ok {
			setController.SetController(dfaController)
		}
		installerView = guiUI

	case core.ModeBrowser:
		browserUI := ui.NewGUIDFA()
		if browserInitializer, ok := browserUI.(interface{ Initialize(*core.Context) error }); ok {
			if err := browserInitializer.Initialize(ctx); err != nil {
				log.Fatalf("Failed to initialize browser UI: %v", err)
			}
		}
		if setController, ok := browserUI.(interface{ SetController(*controller.InstallerController) }); ok {
			setController.SetController(dfaController)
		}
		installerView = browserUI

	case core.ModeSilent:
		silentUI := ui.NewSilentUIDFA()
		if err := silentUI.Initialize(ctx); err != nil {
			log.Fatalf("Failed to initialize Silent UI: %v", err)
		}
		silentUI.SetController(dfaController)
		installerView = silentUI

	case core.ModeAuto:
		// Auto-detect best mode (default to CLI)
		cliUI := cli.NewDFA()
		if err := cliUI.Initialize(ctx); err != nil {
			log.Fatalf("Failed to initialize CLI UI: %v", err)
		}
		cliUI.SetController(dfaController)
		installerView = cliUI

	default:
		log.Fatalf("Unknown UI mode: %v", uiMode)
	}

	// Set the UI view on the DFA controller
	dfaController.SetView(installerView)

	// Start the DFA controller - ALL modes use the same DFA approach
	if err := dfaController.Start(); err != nil {
		log.Fatalf("Installation failed: %v", err)
	}

	// Cleanup
	if shutdownable, ok := installerView.(interface{ Shutdown() error }); ok {
		if err := shutdownable.Shutdown(); err != nil {
			log.Printf("Shutdown warning: %v", err)
		}
	}

	fmt.Printf("\n%s installation completed successfully! 🎉\n", config.AppName)
}

// loadYAMLConfig loads configuration from external file or embedded config
func loadYAMLConfig(filename string) (*InstallerConfig, error) {
	var data []byte
	var err error

	if filename != "" {
		// Load from external file if specified
		data, err = os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file '%s': %w", filename, err)
		}
		fmt.Printf("Using external config file: %s\n", filename)
	} else {
		// Use embedded configuration
		data = embeddedConfig
		fmt.Printf("Using embedded configuration\n")
	}

	var config InstallerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// listInstallationProfiles displays available installation profiles
func listInstallationProfiles(config *InstallerConfig) {
	fmt.Println("Available installation profiles:")
	fmt.Println()
	
	for name, profile := range config.Profiles {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    Description: %s\n", profile.Description)
		fmt.Printf("    Components: %v\n", profile.Components)
		fmt.Printf("    Shortcuts: %v\n", profile.CreateShortcuts)
		if profile.AddToPath {
			fmt.Printf("    Add to PATH: %v\n", profile.AddToPath)
		}
		fmt.Println()
	}
	
	fmt.Println("Usage: installer -profile=<name>")
}

// applyProfile applies an installation profile to the configuration
func applyProfile(config *InstallerConfig, profileName string) error {
	profile, exists := config.Profiles[profileName]
	if !exists {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	// Reset all component selections
	for i := range config.Components {
		config.Components[i].Selected = config.Components[i].Required
	}

	// Select components specified in profile
	for _, componentID := range profile.Components {
		found := false
		for i, comp := range config.Components {
			if comp.ID == componentID {
				config.Components[i].Selected = true
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("component '%s' specified in profile not found", componentID)
		}
	}

	// Apply profile settings
	config.Settings.CreateShortcuts = profile.CreateShortcuts
	if profile.AddToPath {
		config.Settings.AddToPath = profile.AddToPath
	}

	return nil
}

// createConfigFromYAML converts YAML config to core.Config
func createConfigFromYAML(yamlConfig *InstallerConfig) *core.Config {
	// Determine installation directory
	installDir := yamlConfig.InstallDir
	if installDir == "" {
		installDir = getDefaultInstallDir(yamlConfig.AppName)
	}

	// Convert components
	var components []core.Component
	for _, comp := range yamlConfig.Components {
		components = append(components, core.Component{
			ID:          comp.ID,
			Name:        comp.Name,
			Description: comp.Description,
			Size:        calculateComponentSize(comp.Files),
			Required:    comp.Required,
			Selected:    comp.Selected,
			Files:       comp.Files,
		})
	}

	config := &core.Config{
		AppName:       yamlConfig.AppName,
		Version:       yamlConfig.Version,
		Publisher:     yamlConfig.Publisher,
		Website:       yamlConfig.Website,
		InstallDir:    installDir,
		License:       yamlConfig.License,
		Components:    components,
		Unattended:    yamlConfig.Unattended,
		AcceptLicense: yamlConfig.AcceptLicense,
		
		// Installation callbacks
		BeforeInstall: func() error {
			fmt.Println("Preparing installation...")
			return nil
		},
		
		OnProgress: func(progress float64, message string) {
			// Progress updates handled by UI
		},
		
		AfterInstall: func() error {
			fmt.Println("Finalizing installation...")
			if yamlConfig.Settings.CreateShortcuts {
				return createShortcuts(installDir)
			}
			return nil
		},
	}

	return config
}

// getDefaultInstallDir returns the default installation directory for the current platform
func getDefaultInstallDir(appName string) string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("C:\\Program Files\\%s", appName)
	case "darwin":
		return fmt.Sprintf("/Applications/%s", appName)
	default:
		return fmt.Sprintf("/opt/%s", strings.ToLower(appName))
	}
}

// calculateComponentSize calculates the size of a component based on its files
func calculateComponentSize(files []string) int64 {
	var totalSize int64

	for _, filename := range files {
		// Try to get size from embedded assets first
		assetPath := "assets/" + filename
		if info, err := fs.Stat(embeddedAssets, assetPath); err == nil {
			totalSize += info.Size()
		} else {
			// Estimate size if file doesn't exist
			totalSize += 1024 // 1KB default
		}
	}

	return totalSize
}

// determineUIMode determines the best UI mode based on parameters
func determineUIMode(mode string, unattended bool) core.Mode {
	if unattended {
		return core.ModeSilent
	}

	switch mode {
	case "gui":
		return core.ModeGUI  // Native WebView GUI
	case "browser":
		return core.ModeBrowser  // Browser-based UI (need to add this enum)
	case "cli":
		return core.ModeCLI
	case "silent":
		return core.ModeSilent
	case "auto":
		return core.ModeAuto
	default:
		return core.ModeAuto
	}
}

// getModeName returns a human-readable mode name
func getModeName(mode core.Mode) string {
	switch mode {
	case core.ModeGUI:
		return "GUI"
	case core.ModeBrowser:
		return "Browser"
	case core.ModeCLI:
		return "CLI"
	case core.ModeSilent:
		return "Silent"
	case core.ModeAuto:
		return "Auto"
	default:
		return "Unknown"
	}
}

// copyInstallationFiles copies the selected component files to the installation directory
func copyInstallationFiles(installPath string, components []core.Component) error {
	fmt.Printf("Installing files to: %s\n", installPath)

	// Create installation directory
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	totalFiles := 0
	copiedFiles := 0

	// Count total files to copy
	for _, comp := range components {
		if comp.Selected || comp.Required {
			totalFiles += len(comp.Files)
		}
	}

	// Copy files for each selected component
	for _, comp := range components {
		if !comp.Selected && !comp.Required {
			continue
		}

		fmt.Printf("Installing component: %s\n", comp.Name)

		for _, filename := range comp.Files {
			assetPath := "assets/" + filename
			dstPath := filepath.Join(installPath, filename)

			if err := copyEmbeddedFile(assetPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy %s: %w", filename, err)
			}

			copiedFiles++
			progress := float64(copiedFiles) / float64(totalFiles)
			fmt.Printf("  Copied %s (%.0f%%)\n", filename, progress*100)
		}
	}

	fmt.Printf("Successfully installed %d files from embedded assets\n", copiedFiles)
	return nil
}

// copyEmbeddedFile copies a file from embedded assets to destination
func copyEmbeddedFile(embeddedPath, dstPath string) error {
	// Open embedded file
	srcFile, err := embeddedAssets.Open(embeddedPath)
	if err != nil {
		return fmt.Errorf("failed to open embedded file %s: %w", embeddedPath, err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	// Copy data
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

// copyFile copies a single file from src to dst (kept for backward compatibility)
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// createShortcuts creates desktop and start menu shortcuts (platform-specific)
func createShortcuts(installPath string) error {
	fmt.Println("Creating shortcuts...")
	
	// This is a simplified implementation
	// In a real installer, you would use platform-specific APIs
	switch runtime.GOOS {
	case "windows":
		fmt.Println("  Windows shortcuts created")
	case "darwin":
		fmt.Println("  macOS shortcuts created")
	default:
		fmt.Println("  Linux shortcuts created")
	}
	
	return nil
}
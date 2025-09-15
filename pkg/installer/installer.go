// Package installer provides a framework for building native installers in pure Go
// This file serves as the main entry point and maintains backward compatibility
package installer

import (
	"context"
	"embed"
	"fmt"

	"github.com/mmso2016/setupkit/pkg/installer/config"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/installer/themes"
)

// Re-export core types for backward compatibility
type (
	Mode              = core.Mode
	RollbackStrategy  = core.RollbackStrategy
	Component         = core.Component
	Config            = core.Config
	PathConfiguration = core.PathConfiguration
	Context           = core.Context
	Checkpoint        = core.Checkpoint
	Logger            = core.Logger
	ProgressReporter  = core.ProgressReporter
	PlatformInstaller = core.PlatformInstaller
)

// Re-export constants
const (
	ModeAuto   = core.ModeAuto
	ModeGUI    = core.ModeGUI
	ModeCLI    = core.ModeCLI
	ModeSilent = core.ModeSilent

	RollbackNone    = core.RollbackNone
	RollbackPartial = core.RollbackPartial
	RollbackFull    = core.RollbackFull
)

// Installer wraps the core installer for backward compatibility
type Installer struct {
	core *core.Installer
}

// Option is a functional option for configuring the installer
type Option func(*Config) error

// WithPublisher sets the publisher
func WithPublisher(publisher string) Option {
	return func(c *Config) error {
		c.Publisher = publisher
		return nil
	}
}

// WithWebsite sets the website
func WithWebsite(website string) Option {
	return func(c *Config) error {
		c.Website = website
		return nil
	}
}

// WithPathConfig sets the PATH configuration
func WithPathConfig(pathConfig *PathConfiguration) Option {
	return func(c *Config) error {
		c.PathConfig = pathConfig
		return nil
	}
}

// WithAssets sets the embedded assets
func WithAssets(assets embed.FS) Option {
	return func(c *Config) error {
		c.Assets = assets
		return nil
	}
}

// WithAppName sets the application name
func WithAppName(name string) Option {
	return func(c *Config) error {
		if name == "" {
			return fmt.Errorf("app name cannot be empty")
		}
		c.AppName = name
		return nil
	}
}

// WithVersion sets the version
func WithVersion(version string) Option {
	return func(c *Config) error {
		c.Version = version
		return nil
	}
}

// WithMode sets the installation mode
func WithMode(mode Mode) Option {
	return func(c *Config) error {
		c.Mode = mode
		return nil
	}
}

// WithComponents sets the installable components
func WithComponents(components ...Component) Option {
	return func(c *Config) error {
		c.Components = components
		return nil
	}
}

// WithRollback sets the rollback strategy
func WithRollback(strategy RollbackStrategy) Option {
	return func(c *Config) error {
		c.Rollback = strategy
		return nil
	}
}

// WithInstallDir sets the installation directory
func WithInstallDir(dir string) Option {
	return func(c *Config) error {
		c.InstallDir = dir
		return nil
	}
}

// WithResponseFile sets the response file for unattended installation
func WithResponseFile(file string) Option {
	return func(c *Config) error {
		c.ResponseFile = file
		c.Unattended = true
		return nil
	}
}

// WithVerbose enables or disables verbose logging
func WithVerbose(verbose bool) Option {
	return func(c *Config) error {
		c.Verbose = verbose
		return nil
	}
}

// WithLicense sets the license text
func WithLicense(license string) Option {
	return func(c *Config) error {
		c.License = license
		return nil
	}
}

// WithPathConfiguration enables PATH management with specified scope
func WithPathConfiguration(enabled bool, system bool) Option {
	return func(c *Config) error {
		c.PathConfig = &PathConfiguration{
			Enabled: enabled,
			System:  system,
		}
		return nil
	}
}

// WithElevationStrategy sets the elevation strategy
func WithElevationStrategy(strategy core.ElevationStrategy) Option {
	return func(c *Config) error {
		c.ElevationStrategy = strategy
		return nil
	}
}

// WithWizardProvider enables DFA-based wizard with the specified provider
func WithWizardProvider(providerName string) Option {
	return func(c *Config) error {
		c.WizardProvider = providerName
		return nil
	}
}

// WithDFAWizard enables the standard DFA wizard (express mode)
func WithDFAWizard() Option {
	return WithWizardProvider("standard-express")
}

// WithCustomDFAWizard enables the standard DFA wizard (custom mode)
func WithCustomDFAWizard() Option {
	return WithWizardProvider("standard-custom")
}

// WithAdvancedDFAWizard enables the standard DFA wizard (advanced mode) 
func WithAdvancedDFAWizard() Option {
	return WithWizardProvider("standard-advanced")
}

// WithThemeSelection enables theme selection in the wizard
func WithThemeSelection(enabled bool) Option {
	return func(c *Config) error {
		c.EnableThemeSelection = enabled
		return nil
	}
}

// WithExtendedWizard enables an extended wizard with theme selection
func WithExtendedWizard(themes []string, defaultTheme string) Option {
	return func(c *Config) error {
		c.WizardProvider = "extended"
		c.EnableThemeSelection = true
		if c.WizardOptions == nil {
			c.WizardOptions = make(map[string]interface{})
		}
		c.WizardOptions["themes"] = themes
		c.WizardOptions["default_theme"] = defaultTheme
		return nil
	}
}

// WithDryRun enables or disables dry run mode
func WithDryRun(dryRun bool) Option {
	return func(c *Config) error {
		c.DryRun = dryRun
		return nil
	}
}

// WithForce enables or disables force installation
func WithForce(force bool) Option {
	return func(c *Config) error {
		c.Force = force
		return nil
	}
}

// UI Configuration Options

// WithUIConfig loads UI configuration from a YAML file
func WithUIConfig(filename string) Option {
	return func(c *Config) error {
		return c.LoadUIConfig(filename)
	}
}

// WithTheme sets the theme by name
func WithTheme(themeName string) Option {
	return func(c *Config) error {
		theme, err := themes.GetTheme(themeName)
		if err != nil {
			return fmt.Errorf("failed to load theme '%s': %w", themeName, err)
		}
		c.Theme = theme
		return nil
	}
}

// WithThemeColors applies custom branding colors to the current theme
func WithThemeColors(primaryColor, secondaryColor string) Option {
	return func(c *Config) error {
		// Ensure we have a theme to modify
		if c.Theme.Name == "" {
			if err := WithTheme("default")(c); err != nil {
				return err
			}
		}

		// Create custom branding
		if c.UIConfig == nil {
			c.UIConfig = config.GetDefaultConfig()
		}
		c.UIConfig.Branding.PrimaryColor = primaryColor
		c.UIConfig.Branding.SecondaryColor = secondaryColor

		return c.ApplyBranding()
	}
}

// WithBranding applies complete branding configuration
func WithBranding(primaryColor, secondaryColor, logo, fontFamily string) Option {
	return func(c *Config) error {
		// Ensure we have a UI config
		if c.UIConfig == nil {
			c.UIConfig = config.GetDefaultConfig()
		}

		branding := &c.UIConfig.Branding
		if primaryColor != "" {
			branding.PrimaryColor = primaryColor
		}
		if secondaryColor != "" {
			branding.SecondaryColor = secondaryColor
		}
		if logo != "" {
			branding.Logo = logo
		}
		if fontFamily != "" {
			branding.FontFamily = fontFamily
		}

		// Apply the theme if not already set
		if c.Theme.Name == "" {
			if err := c.ApplyTheme(); err != nil {
				return err
			}
		}

		return c.ApplyBranding()
	}
}

// WithScreenConfig configures which screens are enabled
func WithScreenConfig(screenConfigs map[string]bool) Option {
	return func(c *Config) error {
		// Ensure we have a UI config
		if c.UIConfig == nil {
			c.UIConfig = config.GetDefaultConfig()
		}

		// Apply screen configuration
		for screenID, enabled := range screenConfigs {
			if screen, exists := c.UIConfig.Screens[screenID]; exists {
				screen.Enabled = enabled
				c.UIConfig.Screens[screenID] = screen
			}
		}

		return config.ValidateConfig(c.UIConfig)
	}
}

// WithScreenTitle sets custom title for a screen
func WithScreenTitle(screenID, title string) Option {
	return func(c *Config) error {
		// Ensure we have a UI config
		if c.UIConfig == nil {
			c.UIConfig = config.GetDefaultConfig()
		}

		// Update screen title
		if screen, exists := c.UIConfig.Screens[screenID]; exists {
			screen.Title = title
			c.UIConfig.Screens[screenID] = screen
		} else {
			return fmt.Errorf("screen '%s' not found", screenID)
		}

		return nil
	}
}

// WithWelcomeMessage sets a custom welcome message
func WithWelcomeMessage(title, message string) Option {
	return func(c *Config) error {
		// Ensure we have a UI config
		if c.UIConfig == nil {
			c.UIConfig = config.GetDefaultConfig()
		}

		// Update welcome screen
		if screen, exists := c.UIConfig.Screens[config.ScreenWelcome]; exists {
			if title != "" {
				screen.Title = title
			}
			if message != "" {
				screen.Message = message
			}
			c.UIConfig.Screens[config.ScreenWelcome] = screen
		}

		return nil
	}
}

// New creates a new installer with the given options
func New(opts ...Option) (*Installer, error) {
	config := &Config{
		Mode:     ModeAuto,
		Rollback: RollbackPartial,
		LogLevel: "info",
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create core installer
	coreInstaller := core.New(config)

	return &Installer{
		core: coreInstaller,
	}, nil
}

// Run executes the installer
func (i *Installer) Run() error {
	return i.core.Run(context.Background())
}

// RunWithContext executes the installer with a custom context
func (i *Installer) RunWithContext(ctx context.Context) error {
	return i.core.Run(ctx)
}

// GetConfig returns the installer configuration
func (i *Installer) GetConfig() *Config {
	return i.core.GetConfig()
}

// GetComponents returns the available components
func (i *Installer) GetComponents() []Component {
	return i.core.GetComponents()
}

// SetSelectedComponents updates the selected state of components
func (i *Installer) SetSelectedComponents(components []Component) {
	i.core.SetSelectedComponents(components)
}

// SetInstallPath sets the installation directory
func (i *Installer) SetInstallPath(path string) {
	i.core.SetInstallPath(path)
}

// IsUsingDFAWizard returns true if the installer is using the DFA-based wizard
func (i *Installer) IsUsingDFAWizard() bool {
	return i.core.IsUsingDFAWizard()
}

// GetWizardAdapter returns the wizard UI adapter (if DFA wizard is enabled)
func (i *Installer) GetWizardAdapter() *core.WizardUIAdapter {
	return i.core.GetWizardAdapter()
}

// EnableDFAWizard enables the DFA-based wizard system
func (i *Installer) EnableDFAWizard(providerName string) error {
	return i.core.EnableDFAWizard(providerName)
}

// EnableExtendedWizardWithThemes enables an extended wizard with theme selection
func (i *Installer) EnableExtendedWizardWithThemes(themes []string, defaultTheme string) error {
	return i.core.EnableExtendedWizardWithThemes(themes, defaultTheme)
}

// Utility functions exported for backward compatibility

// CheckDiskSpace verifies sufficient disk space is available
func CheckDiskSpace(path string, required int64) error {
	return core.CheckDiskSpace(path, required)
}

// ExtractAssets extracts embedded assets to the target directory
func ExtractAssets(assets embed.FS, targetDir string) error {
	return core.ExtractAssets(assets, targetDir)
}

// NewLogger creates a new logger instance
func NewLogger(level, logFile string) Logger {
	return core.NewLogger(level, logFile)
}

// UI Configuration Utilities

// GenerateConfigFile creates a default UI configuration file
func GenerateConfigFile(filename, appName string) error {
	return core.GenerateDefaultConfig(filename, appName)
}

// ListThemes returns information about all available themes
func ListThemes() []themes.ThemeInfo {
	return core.ListAvailableThemes()
}

// PreviewTheme returns CSS for a theme without applying it
func PreviewTheme(themeName string) (string, error) {
	return core.PreviewTheme(themeName)
}

// LoadUIConfig loads a UI configuration from file
func LoadUIConfig(filename string) (*config.UIConfig, error) {
	return config.LoadConfig(filename)
}

// ValidateUIConfig validates a UI configuration
func ValidateUIConfig(uiConfig *config.UIConfig) error {
	return config.ValidateConfig(uiConfig)
}

// GetBuiltinThemes returns all built-in themes
func GetBuiltinThemes() map[string]themes.Theme {
	return themes.GetBuiltinThemes()
}

// Screen configuration constants (re-exported for convenience)
const (
	ScreenWelcome      = config.ScreenWelcome
	ScreenLicense      = config.ScreenLicense
	ScreenComponents   = config.ScreenComponents
	ScreenDirectory    = config.ScreenDirectory
	ScreenInstallation = config.ScreenInstallation
	ScreenFinish       = config.ScreenFinish
)

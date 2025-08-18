// Package core provides configuration utilities for YAML and theme integration
package core

import (
	"fmt"
	"strings"

	"github.com/mmso2016/setupkit/installer/config"
	"github.com/mmso2016/setupkit/installer/themes"
)

// LoadUIConfig loads and applies UI configuration from YAML file
func (c *Config) LoadUIConfig(filename string) error {
	uiConfig, err := config.LoadConfig(filename)
	if err != nil {
		return fmt.Errorf("failed to load UI config: %w", err)
	}

	if err := config.ValidateConfig(uiConfig); err != nil {
		return fmt.Errorf("invalid UI config: %w", err)
	}

	c.UIConfig = uiConfig
	c.ConfigFile = filename

	// Load theme
	if err := c.ApplyTheme(); err != nil {
		return fmt.Errorf("failed to apply theme: %w", err)
	}

	// Apply branding overrides if specified
	if err := c.ApplyBranding(); err != nil {
		return fmt.Errorf("failed to apply branding: %w", err)
	}

	// Apply screen configuration
	if err := c.ApplyScreenConfig(); err != nil {
		return fmt.Errorf("failed to apply screen config: %w", err)
	}

	return nil
}

// ApplyTheme loads and applies the specified theme
func (c *Config) ApplyTheme() error {
	if c.UIConfig == nil {
		return fmt.Errorf("UI config not loaded")
	}

	themeName := c.UIConfig.UI.Theme
	if themeName == "" {
		themeName = "default"
	}

	theme, err := themes.GetTheme(themeName)
	if err != nil {
		return fmt.Errorf("failed to get theme '%s': %w", themeName, err)
	}

	c.Theme = theme
	return nil
}

// ApplyBranding applies branding overrides to the theme
func (c *Config) ApplyBranding() error {
	if c.UIConfig == nil || c.UIConfig.Branding.PrimaryColor == "" {
		return nil // No branding to apply
	}

	branding := c.UIConfig.Branding

	// Create a copy of the theme to modify
	customTheme := c.Theme
	customTheme.Name = "custom-" + c.Theme.Name
	customTheme.Description = "Custom branded theme"

	// Apply branding colors
	if branding.PrimaryColor != "" {
		customTheme.Colors.Primary = branding.PrimaryColor
	}
	if branding.SecondaryColor != "" {
		customTheme.Colors.Secondary = branding.SecondaryColor
	}
	if branding.FontFamily != "" {
		customTheme.Typography.FontFamily = branding.FontFamily
		customTheme.Typography.HeadingFamily = branding.FontFamily
	}

	// Regenerate CSS with new branding
	customTheme.CSS = themes.GenerateCSS(customTheme)
	c.Theme = customTheme

	// Apply branding to general config
	if branding.CompanyName != "" && c.Publisher == "" {
		c.Publisher = branding.CompanyName
	}

	return nil
}

// ApplyScreenConfig applies screen configuration from UI config
func (c *Config) ApplyScreenConfig() error {
	if c.UIConfig == nil {
		return nil
	}

	// Apply general UI settings
	if c.UIConfig.UI.Title != "" && c.AppName == "" {
		c.AppName = c.UIConfig.UI.Title
	}

	// Screen configuration will be handled by the UI implementation
	// This is just validation and preparation
	return nil
}

// GetEnabledScreens returns a list of enabled screens in order
func (c *Config) GetEnabledScreens() []config.ScreenConfig {
	if c.UIConfig == nil {
		return nil
	}

	var screens []config.ScreenConfig
	
	// Collect enabled screens
	for screenID, screen := range c.UIConfig.Screens {
		if screen.Enabled {
			// Set ID for reference
			screenCopy := screen
			screenCopy.Type = screenID
			screens = append(screens, screenCopy)
		}
	}

	// Sort by position
	for i := 0; i < len(screens)-1; i++ {
		for j := 0; j < len(screens)-i-1; j++ {
			if screens[j].Position > screens[j+1].Position {
				screens[j], screens[j+1] = screens[j+1], screens[j]
			}
		}
	}

	return screens
}

// IsScreenEnabled checks if a specific screen is enabled
func (c *Config) IsScreenEnabled(screenID string) bool {
	if c.UIConfig == nil {
		return true // Default to enabled if no config
	}

	screen, exists := c.UIConfig.Screens[screenID]
	if !exists {
		return true // Default to enabled if not configured
	}

	return screen.Enabled
}

// GetScreenConfig returns configuration for a specific screen
func (c *Config) GetScreenConfig(screenID string) (config.ScreenConfig, bool) {
	if c.UIConfig == nil {
		return config.ScreenConfig{}, false
	}

	screen, exists := c.UIConfig.Screens[screenID]
	return screen, exists
}

// GetThemeCSS returns the CSS for the current theme
func (c *Config) GetThemeCSS() string {
	return c.Theme.CSS
}

// GetThemeInfo returns basic information about the current theme
func (c *Config) GetThemeInfo() themes.ThemeInfo {
	return themes.ThemeInfo{
		Name:        c.Theme.Name,
		Description: c.Theme.Description,
	}
}

// ValidateUIConfig validates the current UI configuration
func (c *Config) ValidateUIConfig() error {
	if c.UIConfig == nil {
		return nil // No UI config to validate
	}

	return config.ValidateConfig(c.UIConfig)
}

// GenerateDefaultConfig creates a default UI configuration file
func GenerateDefaultConfig(filename string, appName string) error {
	uiConfig := config.GetDefaultConfig()
	
	// Customize for the application
	if appName != "" {
		uiConfig.UI.Title = appName + " Installer"
		
		// Update screen titles
		if welcome, exists := uiConfig.Screens[config.ScreenWelcome]; exists {
			welcome.Title = "Welcome to " + appName
			welcome.Message = fmt.Sprintf("This will install %s on your computer.", appName)
			uiConfig.Screens[config.ScreenWelcome] = welcome
		}
	}

	return config.SaveConfig(uiConfig, filename)
}

// ListAvailableThemes returns information about all available themes
func ListAvailableThemes() []themes.ThemeInfo {
	return themes.ListThemes()
}

// PreviewTheme returns CSS for a theme without applying it
func PreviewTheme(themeName string) (string, error) {
	theme, err := themes.GetTheme(themeName)
	if err != nil {
		return "", err
	}
	return theme.CSS, nil
}

// MergeConfigs merges a UI config with an existing installer config
func MergeConfigs(installerConfig *Config, uiConfigFile string) error {
	if installerConfig == nil {
		return fmt.Errorf("installer config is nil")
	}

	return installerConfig.LoadUIConfig(uiConfigFile)
}

// GetConfigSummary returns a summary of the current configuration
func (c *Config) GetConfigSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"app_name":    c.AppName,
		"version":     c.Version,
		"publisher":   c.Publisher,
		"install_dir": c.InstallDir,
		"mode":        c.Mode,
	}

	if c.UIConfig != nil {
		summary["ui_theme"] = c.UIConfig.UI.Theme
		summary["enabled_screens"] = len(c.GetEnabledScreens())
		
		var screenNames []string
		for _, screen := range c.GetEnabledScreens() {
			screenNames = append(screenNames, screen.Type)
		}
		summary["screen_order"] = strings.Join(screenNames, " → ")
	}

	if c.Theme.Name != "" {
		summary["theme_name"] = c.Theme.Name
		summary["theme_description"] = c.Theme.Description
	}

	return summary
}

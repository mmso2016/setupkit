// Package config provides YAML-based configuration for the installer UI
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// UIConfig represents the complete installer configuration
type UIConfig struct {
	UI       UI                     `yaml:"ui"`
	Branding Branding               `yaml:"branding,omitempty"`
	Screens  map[string]ScreenConfig `yaml:"screens,omitempty"`
}

// UI contains general UI settings
type UI struct {
	Theme      string `yaml:"theme"`
	Logo       string `yaml:"logo,omitempty"`
	Title      string `yaml:"title,omitempty"`
	AutoTheme  bool   `yaml:"auto_theme,omitempty"`
}

// Branding contains brand-specific settings
type Branding struct {
	PrimaryColor   string `yaml:"primary_color,omitempty"`
	SecondaryColor string `yaml:"secondary_color,omitempty"`
	Logo           string `yaml:"logo,omitempty"`
	FontFamily     string `yaml:"font_family,omitempty"`
	CompanyName    string `yaml:"company_name,omitempty"`
}

// ScreenConfig represents configuration for a single screen
type ScreenConfig struct {
	Enabled      bool              `yaml:"enabled"`
	Title        string            `yaml:"title,omitempty"`
	Message      string            `yaml:"message,omitempty"`
	CustomText   string            `yaml:"custom_text,omitempty"`
	Template     string            `yaml:"template,omitempty"`
	Type         string            `yaml:"type,omitempty"` // "standard" or "custom"
	Position     int               `yaml:"position,omitempty"`
	Properties   map[string]any    `yaml:"properties,omitempty"`
}

// Default screens that can be configured
const (
	ScreenWelcome      = "welcome"
	ScreenLicense      = "license"
	ScreenComponents   = "components"
	ScreenDirectory    = "directory"
	ScreenInstallation = "installation"
	ScreenFinish       = "finish"
)

// LoadConfig loads configuration from YAML file
func LoadConfig(filename string) (*UIConfig, error) {
	if filename == "" {
		return GetDefaultConfig(), nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return GetDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config UIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Apply defaults for missing screens
	if config.Screens == nil {
		config.Screens = make(map[string]ScreenConfig)
	}

	// Ensure all default screens exist
	defaultScreens := GetDefaultScreens()
	for screenID, defaultScreen := range defaultScreens {
		if _, exists := config.Screens[screenID]; !exists {
			config.Screens[screenID] = defaultScreen
		}
	}

	return &config, nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *UIConfig {
	return &UIConfig{
		UI: UI{
			Theme: "default",
			Title: "Application Installer",
		},
		Screens: GetDefaultScreens(),
	}
}

// GetDefaultScreens returns default screen configurations
func GetDefaultScreens() map[string]ScreenConfig {
	return map[string]ScreenConfig{
		ScreenWelcome: {
			Enabled:  true,
			Title:    "Welcome",
			Message:  "Welcome to the installation wizard",
			Position: 1,
		},
		ScreenLicense: {
			Enabled:  true,
			Title:    "License Agreement",
			Message:  "Please read and accept the license agreement",
			Position: 2,
		},
		ScreenComponents: {
			Enabled:  true,
			Title:    "Select Components",
			Message:  "Choose which components to install",
			Position: 3,
		},
		ScreenDirectory: {
			Enabled:  true,
			Title:    "Installation Directory",
			Message:  "Choose where to install the application",
			Position: 4,
		},
		ScreenInstallation: {
			Enabled:  true,
			Title:    "Installing",
			Message:  "Please wait while the application is being installed",
			Position: 5,
		},
		ScreenFinish: {
			Enabled:  true,
			Title:    "Installation Complete",
			Message:  "The application has been successfully installed",
			Position: 6,
		},
	}
}

// SaveConfig saves configuration to YAML file
func SaveConfig(config *UIConfig, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig validates the configuration for common issues
func ValidateConfig(config *UIConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if config.UI.Theme == "" {
		return fmt.Errorf("theme is required")
	}

	// Validate screens have valid positions
	positions := make(map[int]string)
	for screenID, screen := range config.Screens {
		if screen.Enabled && screen.Position > 0 {
			if existing, exists := positions[screen.Position]; exists {
				return fmt.Errorf("duplicate position %d for screens %s and %s", 
					screen.Position, existing, screenID)
			}
			positions[screen.Position] = screenID
		}
	}

	return nil
}

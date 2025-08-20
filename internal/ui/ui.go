package ui

import (
	_ "embed"
	"html/template"
	"strings"
)

// Embedded UI assets
var (
	//go:embed templates/installer.html
	defaultHTML string

	//go:embed templates/installer.css
	defaultCSS string

	//go:embed templates/installer.js
	defaultJS string
)

// Config represents the installer UI configuration
type Config struct {
	AppName         string
	Version         string
	Publisher       string
	Website         string
	License         string
	InstallPath     string
	Components      []Component
	Theme           ThemeConfig
	ShowThemeSelect bool
}

// Component represents an installable component
type Component struct {
	ID          string
	Name        string
	Description string
	Size        int64
	Required    bool
	Selected    bool
}

// ThemeConfig allows customization of the UI appearance
type ThemeConfig struct {
	PrimaryColor   string
	SecondaryColor string
	LogoURL        string
	CustomCSS      string
}

// GenerateHTML creates the final installer HTML with embedded assets
func GenerateHTML(config *Config) (string, error) {
	// Set defaults
	if config.AppName == "" {
		config.AppName = "Application"
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	if config.Publisher == "" {
		config.Publisher = "Unknown Publisher"
	}
	if config.InstallPath == "" {
		config.InstallPath = "C:\\Program Files\\" + config.AppName
	}

	// Build the HTML with embedded CSS and JS
	html := strings.ReplaceAll(defaultHTML, `<link rel="stylesheet" href="style.css">`, 
		`<style>`+defaultCSS+`</style>`)
	
	// Embed JS directly
	html = strings.ReplaceAll(html, `<script src="./wailsjs/runtime/runtime.js"></script>`, "")
	html = strings.ReplaceAll(html, `<script src="app.js"></script>`, 
		`<script>`+defaultJS+`</script>`)

	// Apply theme if provided
	if config.Theme.CustomCSS != "" {
		html = strings.ReplaceAll(html, `</style>`, 
			config.Theme.CustomCSS+`</style>`)
	}

	// Process template with config data
	tmpl, err := template.New("installer").Parse(html)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, config); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetDefaultHTML returns the raw HTML template
func GetDefaultHTML() string {
	return defaultHTML
}

// GetDefaultCSS returns the default stylesheet
func GetDefaultCSS() string {
	return defaultCSS
}

// GetDefaultJS returns the default JavaScript
func GetDefaultJS() string {
	return defaultJS
}

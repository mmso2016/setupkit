// Package themes provides theme management for the installer UI
package themes

import (
	"fmt"
	"strings"
)

// Theme represents a complete theme configuration
type Theme struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Colors      Colors            `json:"colors"`
	Typography  Typography        `json:"typography"`
	Layout      Layout            `json:"layout"`
	Animations  Animations        `json:"animations"`
	CSS         string            `json:"-"` // Generated CSS
}

// Colors defines the color palette for a theme
type Colors struct {
	Primary     string `json:"primary"`
	Secondary   string `json:"secondary"`
	Background  string `json:"background"`
	Surface     string `json:"surface"`
	Text        string `json:"text"`
	TextMuted   string `json:"text_muted"`
	Success     string `json:"success"`
	Warning     string `json:"warning"`
	Error       string `json:"error"`
	Border      string `json:"border"`
	Hover       string `json:"hover"`
	Active      string `json:"active"`
}

// Typography defines font settings
type Typography struct {
	FontFamily      string `json:"font_family"`
	FontSize        string `json:"font_size"`
	FontWeight      string `json:"font_weight"`
	LineHeight      string `json:"line_height"`
	HeadingFamily   string `json:"heading_family"`
	HeadingWeight   string `json:"heading_weight"`
}

// Layout defines spacing and sizing
type Layout struct {
	BorderRadius   string `json:"border_radius"`
	Spacing        string `json:"spacing"`
	SpacingLarge   string `json:"spacing_large"`
	SpacingSmall   string `json:"spacing_small"`
	ButtonHeight   string `json:"button_height"`
	InputHeight    string `json:"input_height"`
	SidebarWidth   string `json:"sidebar_width"`
}

// Animations defines animation settings
type Animations struct {
	Transition     string `json:"transition"`
	Duration       string `json:"duration"`
	Easing         string `json:"easing"`
	HoverTransform string `json:"hover_transform"`
}

// Built-in themes
var (
	DefaultTheme = Theme{
		Name:        "default",
		Description: "Clean and modern default theme",
		Colors: Colors{
			Primary:     "#2563eb",
			Secondary:   "#64748b",
			Background:  "#ffffff",
			Surface:     "#f8fafc",
			Text:        "#1e293b",
			TextMuted:   "#64748b",
			Success:     "#10b981",
			Warning:     "#f59e0b",
			Error:       "#ef4444",
			Border:      "#e2e8f0",
			Hover:       "#f1f5f9",
			Active:      "#e2e8f0",
		},
		Typography: Typography{
			FontFamily:    "'Inter', -apple-system, BlinkMacSystemFont, sans-serif",
			FontSize:      "14px",
			FontWeight:    "400",
			LineHeight:    "1.5",
			HeadingFamily: "'Inter', -apple-system, BlinkMacSystemFont, sans-serif",
			HeadingWeight: "600",
		},
		Layout: Layout{
			BorderRadius:  "8px",
			Spacing:       "16px",
			SpacingLarge:  "24px",
			SpacingSmall:  "8px",
			ButtonHeight:  "40px",
			InputHeight:   "40px",
			SidebarWidth:  "280px",
		},
		Animations: Animations{
			Transition:     "all 0.2s ease",
			Duration:       "0.2s",
			Easing:         "ease",
			HoverTransform: "translateY(-1px)",
		},
	}

	CorporateBlueTheme = Theme{
		Name:        "corporate-blue",
		Description: "Professional corporate theme in blue",
		Colors: Colors{
			Primary:     "#1e3a8a",
			Secondary:   "#3b82f6",
			Background:  "#f8fafc",
			Surface:     "#ffffff",
			Text:        "#1e293b",
			TextMuted:   "#475569",
			Success:     "#059669",
			Warning:     "#d97706",
			Error:       "#dc2626",
			Border:      "#cbd5e1",
			Hover:       "#f1f5f9",
			Active:      "#e2e8f0",
		},
		Typography: Typography{
			FontFamily:    "'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif",
			FontSize:      "14px",
			FontWeight:    "400",
			LineHeight:    "1.6",
			HeadingFamily: "'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif",
			HeadingWeight: "600",
		},
		Layout: Layout{
			BorderRadius:  "6px",
			Spacing:       "20px",
			SpacingLarge:  "32px",
			SpacingSmall:  "12px",
			ButtonHeight:  "44px",
			InputHeight:   "44px",
			SidebarWidth:  "300px",
		},
		Animations: Animations{
			Transition:     "all 0.3s ease",
			Duration:       "0.3s",
			Easing:         "ease-out",
			HoverTransform: "translateY(-2px)",
		},
	}

	MedicalGreenTheme = Theme{
		Name:        "medical-green",
		Description: "Clean medical theme in green",
		Colors: Colors{
			Primary:     "#059669",
			Secondary:   "#10b981",
			Background:  "#f0fdf4",
			Surface:     "#ffffff",
			Text:        "#064e3b",
			TextMuted:   "#047857",
			Success:     "#10b981",
			Warning:     "#f59e0b",
			Error:       "#ef4444",
			Border:      "#bbf7d0",
			Hover:       "#ecfdf5",
			Active:      "#d1fae5",
		},
		Typography: Typography{
			FontFamily:    "'Inter', -apple-system, BlinkMacSystemFont, sans-serif",
			FontSize:      "15px",
			FontWeight:    "400",
			LineHeight:    "1.6",
			HeadingFamily: "'Inter', -apple-system, BlinkMacSystemFont, sans-serif",
			HeadingWeight: "500",
		},
		Layout: Layout{
			BorderRadius:  "12px",
			Spacing:       "18px",
			SpacingLarge:  "28px",
			SpacingSmall:  "10px",
			ButtonHeight:  "42px",
			InputHeight:   "42px",
			SidebarWidth:  "320px",
		},
		Animations: Animations{
			Transition:     "all 0.25s ease",
			Duration:       "0.25s",
			Easing:         "ease-in-out",
			HoverTransform: "scale(1.02)",
		},
	}

	TechDarkTheme = Theme{
		Name:        "tech-dark",
		Description: "Modern dark theme for tech applications",
		Colors: Colors{
			Primary:     "#6366f1",
			Secondary:   "#8b5cf6",
			Background:  "#0f172a",
			Surface:     "#1e293b",
			Text:        "#f1f5f9",
			TextMuted:   "#94a3b8",
			Success:     "#22c55e",
			Warning:     "#eab308",
			Error:       "#ef4444",
			Border:      "#334155",
			Hover:       "#475569",
			Active:      "#64748b",
		},
		Typography: Typography{
			FontFamily:    "'JetBrains Mono', 'Menlo', 'Monaco', monospace",
			FontSize:      "14px",
			FontWeight:    "400",
			LineHeight:    "1.5",
			HeadingFamily: "'Inter', -apple-system, BlinkMacSystemFont, sans-serif",
			HeadingWeight: "600",
		},
		Layout: Layout{
			BorderRadius:  "4px",
			Spacing:       "16px",
			SpacingLarge:  "24px",
			SpacingSmall:  "8px",
			ButtonHeight:  "38px",
			InputHeight:   "38px",
			SidebarWidth:  "280px",
		},
		Animations: Animations{
			Transition:     "all 0.15s cubic-bezier(0.4, 0, 0.2, 1)",
			Duration:       "0.15s",
			Easing:         "cubic-bezier(0.4, 0, 0.2, 1)",
			HoverTransform: "scale(1.05)",
		},
	}

	MinimalLightTheme = Theme{
		Name:        "minimal-light",
		Description: "Clean minimal theme with light colors",
		Colors: Colors{
			Primary:     "#000000",
			Secondary:   "#6b7280",
			Background:  "#ffffff",
			Surface:     "#fafafa",
			Text:        "#111827",
			TextMuted:   "#6b7280",
			Success:     "#059669",
			Warning:     "#d97706",
			Error:       "#dc2626",
			Border:      "#e5e7eb",
			Hover:       "#f9fafb",
			Active:      "#f3f4f6",
		},
		Typography: Typography{
			FontFamily:    "'system-ui', -apple-system, sans-serif",
			FontSize:      "14px",
			FontWeight:    "400",
			LineHeight:    "1.5",
			HeadingFamily: "'system-ui', -apple-system, sans-serif",
			HeadingWeight: "500",
		},
		Layout: Layout{
			BorderRadius:  "0px",
			Spacing:       "16px",
			SpacingLarge:  "24px",
			SpacingSmall:  "8px",
			ButtonHeight:  "36px",
			InputHeight:   "36px",
			SidebarWidth:  "260px",
		},
		Animations: Animations{
			Transition:     "none",
			Duration:       "0s",
			Easing:         "linear",
			HoverTransform: "none",
		},
	}
)

// GetBuiltinThemes returns all built-in themes
func GetBuiltinThemes() map[string]Theme {
	themes := map[string]Theme{
		DefaultTheme.Name:        DefaultTheme,
		CorporateBlueTheme.Name:  CorporateBlueTheme,
		MedicalGreenTheme.Name:   MedicalGreenTheme,
		TechDarkTheme.Name:       TechDarkTheme,
		MinimalLightTheme.Name:   MinimalLightTheme,
	}

	// Generate CSS for each theme
	for name, theme := range themes {
		theme.CSS = GenerateCSS(theme)
		themes[name] = theme
	}

	return themes
}

// GetTheme returns a theme by name
func GetTheme(name string) (Theme, error) {
	themes := GetBuiltinThemes()
	theme, exists := themes[name]
	if !exists {
		return Theme{}, fmt.Errorf("theme '%s' not found", name)
	}
	return theme, nil
}

// ListThemes returns a list of available theme names and descriptions
func ListThemes() []ThemeInfo {
	themes := GetBuiltinThemes()
	var result []ThemeInfo
	
	for _, theme := range themes {
		result = append(result, ThemeInfo{
			Name:        theme.Name,
			Description: theme.Description,
		})
	}
	
	return result
}

// ThemeInfo represents basic theme information
type ThemeInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GenerateCSS generates CSS from theme configuration
func GenerateCSS(theme Theme) string {
	var css strings.Builder
	
	css.WriteString("/* Generated theme CSS */\n")
	css.WriteString(":root {\n")
	
	// Colors
	css.WriteString(fmt.Sprintf("  --color-primary: %s;\n", theme.Colors.Primary))
	css.WriteString(fmt.Sprintf("  --color-secondary: %s;\n", theme.Colors.Secondary))
	css.WriteString(fmt.Sprintf("  --color-background: %s;\n", theme.Colors.Background))
	css.WriteString(fmt.Sprintf("  --color-surface: %s;\n", theme.Colors.Surface))
	css.WriteString(fmt.Sprintf("  --color-text: %s;\n", theme.Colors.Text))
	css.WriteString(fmt.Sprintf("  --color-text-muted: %s;\n", theme.Colors.TextMuted))
	css.WriteString(fmt.Sprintf("  --color-success: %s;\n", theme.Colors.Success))
	css.WriteString(fmt.Sprintf("  --color-warning: %s;\n", theme.Colors.Warning))
	css.WriteString(fmt.Sprintf("  --color-error: %s;\n", theme.Colors.Error))
	css.WriteString(fmt.Sprintf("  --color-border: %s;\n", theme.Colors.Border))
	css.WriteString(fmt.Sprintf("  --color-hover: %s;\n", theme.Colors.Hover))
	css.WriteString(fmt.Sprintf("  --color-active: %s;\n", theme.Colors.Active))
	
	// Typography
	css.WriteString(fmt.Sprintf("  --font-family: %s;\n", theme.Typography.FontFamily))
	css.WriteString(fmt.Sprintf("  --font-size: %s;\n", theme.Typography.FontSize))
	css.WriteString(fmt.Sprintf("  --font-weight: %s;\n", theme.Typography.FontWeight))
	css.WriteString(fmt.Sprintf("  --line-height: %s;\n", theme.Typography.LineHeight))
	css.WriteString(fmt.Sprintf("  --heading-family: %s;\n", theme.Typography.HeadingFamily))
	css.WriteString(fmt.Sprintf("  --heading-weight: %s;\n", theme.Typography.HeadingWeight))
	
	// Layout
	css.WriteString(fmt.Sprintf("  --border-radius: %s;\n", theme.Layout.BorderRadius))
	css.WriteString(fmt.Sprintf("  --spacing: %s;\n", theme.Layout.Spacing))
	css.WriteString(fmt.Sprintf("  --spacing-large: %s;\n", theme.Layout.SpacingLarge))
	css.WriteString(fmt.Sprintf("  --spacing-small: %s;\n", theme.Layout.SpacingSmall))
	css.WriteString(fmt.Sprintf("  --button-height: %s;\n", theme.Layout.ButtonHeight))
	css.WriteString(fmt.Sprintf("  --input-height: %s;\n", theme.Layout.InputHeight))
	css.WriteString(fmt.Sprintf("  --sidebar-width: %s;\n", theme.Layout.SidebarWidth))
	
	// Animations
	css.WriteString(fmt.Sprintf("  --transition: %s;\n", theme.Animations.Transition))
	css.WriteString(fmt.Sprintf("  --duration: %s;\n", theme.Animations.Duration))
	css.WriteString(fmt.Sprintf("  --easing: %s;\n", theme.Animations.Easing))
	css.WriteString(fmt.Sprintf("  --hover-transform: %s;\n", theme.Animations.HoverTransform))
	
	css.WriteString("}\n\n")
	
	// Base styles
	css.WriteString(generateBaseStyles())
	
	return css.String()
}

// generateBaseStyles returns common CSS styles that use the theme variables
func generateBaseStyles() string {
	return `/* Base theme styles */
body {
  font-family: var(--font-family);
  font-size: var(--font-size);
  font-weight: var(--font-weight);
  line-height: var(--line-height);
  color: var(--color-text);
  background-color: var(--color-background);
  margin: 0;
  padding: 0;
}

.installer-container {
  background-color: var(--color-background);
  color: var(--color-text);
  min-height: 100vh;
}

.installer-surface {
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius);
  padding: var(--spacing);
}

.installer-button {
  background-color: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--border-radius);
  height: var(--button-height);
  padding: 0 var(--spacing);
  font-family: var(--font-family);
  font-size: var(--font-size);
  font-weight: var(--font-weight);
  cursor: pointer;
  transition: var(--transition);
}

.installer-button:hover {
  transform: var(--hover-transform);
  opacity: 0.9;
}

.installer-button:active {
  transform: scale(0.98);
}

.installer-button.secondary {
  background-color: var(--color-secondary);
}

.installer-input {
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius);
  height: var(--input-height);
  padding: 0 var(--spacing-small);
  font-family: var(--font-family);
  font-size: var(--font-size);
  background-color: var(--color-surface);
  color: var(--color-text);
  transition: var(--transition);
}

.installer-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px rgba(var(--color-primary), 0.1);
}

.installer-heading {
  font-family: var(--heading-family);
  font-weight: var(--heading-weight);
  color: var(--color-text);
  margin: 0 0 var(--spacing) 0;
}

.installer-text-muted {
  color: var(--color-text-muted);
}

.installer-success {
  color: var(--color-success);
}

.installer-warning {
  color: var(--color-warning);
}

.installer-error {
  color: var(--color-error);
}

.installer-sidebar {
  width: var(--sidebar-width);
  background-color: var(--color-surface);
  border-right: 1px solid var(--color-border);
}

.installer-content {
  flex: 1;
  padding: var(--spacing-large);
}

.installer-progress-bar {
  background-color: var(--color-border);
  border-radius: var(--border-radius);
  overflow: hidden;
  height: 8px;
}

.installer-progress-fill {
  background-color: var(--color-primary);
  height: 100%;
  transition: width var(--duration) var(--easing);
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .installer-sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--color-border);
  }
  
  .installer-content {
    padding: var(--spacing);
  }
}
`
}

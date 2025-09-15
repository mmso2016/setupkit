// Package providers implements different wizard flow providers for SetupKit
// This package provides factory functions and utilities for creating wizard providers
// that integrate with the SetupKit installer framework.
package providers

import (
	"fmt"
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// InstallMode defines the installation mode
type InstallMode = core.InstallMode

// Re-export InstallMode constants from core package
const (
	ModeExpress     = core.ModeExpress
	ModeCustom      = core.ModeCustom
	ModeAdvanced    = core.ModeAdvanced
	ModeRepair      = core.ModeRepair
	ModeUninstall   = core.ModeUninstall
	ModeUserDefined = core.ModeUserDefined
)

// ProviderType defines the type of wizard provider
type ProviderType string

const (
	ProviderTypeStandard ProviderType = "standard"
	ProviderTypeExtended ProviderType = "extended"
	ProviderTypeCustom   ProviderType = "custom"
)

// ProviderConfig defines configuration for provider creation
type ProviderConfig struct {
	Type            ProviderType
	Mode            InstallMode
	EnableThemes    bool
	AvailableThemes []string
	DefaultTheme    string
	CustomStates    []core.StateInsertion
	Options         map[string]interface{}
}

// ProviderFactory creates wizard providers based on configuration
type ProviderFactory struct {
	registry map[ProviderType]ProviderCreator
}

// ProviderCreator is a function that creates a wizard provider
type ProviderCreator func(config ProviderConfig) core.WizardProvider

// DefaultFactory is the default provider factory
var DefaultFactory = NewProviderFactory()

// NewProviderFactory creates a new provider factory with built-in providers
func NewProviderFactory() *ProviderFactory {
	factory := &ProviderFactory{
		registry: make(map[ProviderType]ProviderCreator),
	}
	
	// Register built-in providers
	factory.RegisterProvider(ProviderTypeStandard, createStandardProvider)
	factory.RegisterProvider(ProviderTypeExtended, createExtendedProvider)
	
	return factory
}

// RegisterProvider registers a provider creator function
func (f *ProviderFactory) RegisterProvider(providerType ProviderType, creator ProviderCreator) {
	f.registry[providerType] = creator
}

// CreateProvider creates a wizard provider based on the configuration
func (f *ProviderFactory) CreateProvider(config ProviderConfig) (core.WizardProvider, error) {
	creator, exists := f.registry[config.Type]
	if !exists {
		return nil, fmt.Errorf("unknown provider type: %s", config.Type)
	}
	
	provider := creator(config)
	if provider == nil {
		return nil, fmt.Errorf("failed to create provider of type: %s", config.Type)
	}
	
	return provider, nil
}

// GetAvailableProviders returns the list of registered provider types
func (f *ProviderFactory) GetAvailableProviders() []ProviderType {
	types := make([]ProviderType, 0, len(f.registry))
	for providerType := range f.registry {
		types = append(types, providerType)
	}
	return types
}

// createStandardProvider creates a standard wizard provider
func createStandardProvider(config ProviderConfig) core.WizardProvider {
	return core.NewStandardWizardProvider(config.Mode)
}

// createExtendedProvider creates an extended wizard provider with theme support
func createExtendedProvider(config ProviderConfig) core.WizardProvider {
	provider := core.NewExtendedWizardProvider(config.Mode)
	
	// Enable theme selection if configured
	if config.EnableThemes {
		themeConfig := &core.ThemeSelectionConfig{
			Enabled:         true,
			DefaultTheme:    config.DefaultTheme,
			AvailableThemes: config.AvailableThemes,
			ShowPreview:     true,
			AllowCustom:     false,
		}
		if themeConfig.DefaultTheme == "" {
			themeConfig.DefaultTheme = "default"
		}
		provider.EnableThemeSelection(themeConfig)
	}
	
	// Insert custom states if configured
	for _, insertion := range config.CustomStates {
		provider.InsertCustomState(insertion)
	}
	
	return provider
}

// Convenience functions for creating providers

// CreateStandardProvider creates a standard wizard provider
func CreateStandardProvider(mode InstallMode) core.WizardProvider {
	provider, _ := DefaultFactory.CreateProvider(ProviderConfig{
		Type: ProviderTypeStandard,
		Mode: mode,
	})
	return provider
}

// CreateExtendedProvider creates an extended wizard provider
func CreateExtendedProvider(mode InstallMode) core.WizardProvider {
	provider, _ := DefaultFactory.CreateProvider(ProviderConfig{
		Type: ProviderTypeExtended,
		Mode: mode,
	})
	return provider
}

// CreateProviderWithThemes creates an extended provider with theme selection
func CreateProviderWithThemes(mode InstallMode, themes []string, defaultTheme string) core.WizardProvider {
	provider, _ := DefaultFactory.CreateProvider(ProviderConfig{
		Type:            ProviderTypeExtended,
		Mode:            mode,
		EnableThemes:    true,
		AvailableThemes: themes,
		DefaultTheme:    defaultTheme,
	})
	return provider
}

// CreateCustomProvider creates a provider with custom configuration
func CreateCustomProvider(config ProviderConfig) (core.WizardProvider, error) {
	return DefaultFactory.CreateProvider(config)
}

// Builder provides a fluent interface for creating wizard providers
type Builder struct {
	config ProviderConfig
	err    error
}

// NewBuilder creates a new provider builder
func NewBuilder() *Builder {
	return &Builder{
		config: ProviderConfig{
			Type: ProviderTypeStandard,
			Mode: ModeExpress,
			Options: make(map[string]interface{}),
		},
	}
}

// WithType sets the provider type
func (b *Builder) WithType(providerType ProviderType) *Builder {
	if b.err != nil {
		return b
	}
	b.config.Type = providerType
	return b
}

// WithMode sets the installation mode
func (b *Builder) WithMode(mode InstallMode) *Builder {
	if b.err != nil {
		return b
	}
	b.config.Mode = mode
	return b
}

// WithThemes enables theme selection with the specified themes
func (b *Builder) WithThemes(themes []string, defaultTheme string) *Builder {
	if b.err != nil {
		return b
	}
	b.config.EnableThemes = true
	b.config.AvailableThemes = themes
	b.config.DefaultTheme = defaultTheme
	return b
}

// WithCustomState adds a custom state insertion
func (b *Builder) WithCustomState(insertion core.StateInsertion) *Builder {
	if b.err != nil {
		return b
	}
	b.config.CustomStates = append(b.config.CustomStates, insertion)
	return b
}

// WithOption sets a custom option
func (b *Builder) WithOption(key string, value interface{}) *Builder {
	if b.err != nil {
		return b
	}
	b.config.Options[key] = value
	return b
}

// Build creates the configured wizard provider
func (b *Builder) Build() (core.WizardProvider, error) {
	if b.err != nil {
		return nil, b.err
	}
	return DefaultFactory.CreateProvider(b.config)
}

// Utility functions

// GetProviderForMode returns a suitable provider for the given mode
func GetProviderForMode(mode InstallMode) core.WizardProvider {
	switch mode {
	case ModeExpress:
		return CreateStandardProvider(mode)
	case ModeCustom, ModeAdvanced:
		return CreateExtendedProvider(mode)
	case ModeRepair, ModeUninstall:
		// These modes might need specialized providers in the future
		return CreateStandardProvider(mode)
	default:
		return CreateStandardProvider(ModeExpress)
	}
}

// ValidateProviderConfig validates a provider configuration
func ValidateProviderConfig(config ProviderConfig) error {
	if config.Type == "" {
		return fmt.Errorf("provider type is required")
	}
	
	if config.Mode == "" {
		return fmt.Errorf("installation mode is required")
	}
	
	if config.EnableThemes && config.DefaultTheme == "" {
		return fmt.Errorf("default theme is required when themes are enabled")
	}
	
	return nil
}

// GetProviderInfo returns information about a provider type
func GetProviderInfo(providerType ProviderType) map[string]interface{} {
	switch providerType {
	case ProviderTypeStandard:
		return map[string]interface{}{
			"name":        "Standard Provider",
			"description": "Standard wizard flow with basic states",
			"features":    []string{"welcome", "license", "components", "location", "installation", "completion"},
			"modes":       []InstallMode{ModeExpress, ModeCustom, ModeAdvanced},
		}
	case ProviderTypeExtended:
		return map[string]interface{}{
			"name":        "Extended Provider",
			"description": "Extended wizard flow with theme selection and custom states",
			"features":    []string{"theme_selection", "custom_states", "state_insertion", "all_standard_features"},
			"modes":       []InstallMode{ModeExpress, ModeCustom, ModeAdvanced},
		}
	default:
		return map[string]interface{}{
			"name":        "Unknown Provider",
			"description": "Unknown provider type",
			"features":    []string{},
			"modes":       []InstallMode{},
		}
	}
}
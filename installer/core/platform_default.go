package core

// DefaultPlatformInstaller provides a no-op implementation for testing and unsupported platforms
type DefaultPlatformInstaller struct {
	config *Config
}

// NewDefaultPlatformInstaller creates a default platform installer
// This is always available regardless of build tags
func NewDefaultPlatformInstaller(config *Config) PlatformInstaller {
	return &DefaultPlatformInstaller{config: config}
}

// Implement PlatformInstaller interface with no-op methods
func (d *DefaultPlatformInstaller) Initialize() error                          { return nil }
func (d *DefaultPlatformInstaller) CheckRequirements() error                   { return nil }
func (d *DefaultPlatformInstaller) IsElevated() bool                          { return false }
func (d *DefaultPlatformInstaller) RequiresElevation() bool                   { return false }
func (d *DefaultPlatformInstaller) RequestElevation() error                   { return nil }
func (d *DefaultPlatformInstaller) RegisterWithOS() error                     { return nil }
func (d *DefaultPlatformInstaller) CreateShortcuts() error                    { return nil }
func (d *DefaultPlatformInstaller) RegisterUninstaller() error                { return nil }
func (d *DefaultPlatformInstaller) UpdatePath(dirs []string, system bool) error { return nil }
func (d *DefaultPlatformInstaller) AddToPath(dir string, system bool) error    { return nil }
func (d *DefaultPlatformInstaller) RemoveFromPath(dir string, system bool) error { return nil }
func (d *DefaultPlatformInstaller) IsInPath(dir string, system bool) bool     { return false }

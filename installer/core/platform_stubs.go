//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package core

// Platform stub implementations for unsupported platforms
// These functions provide factory methods for platforms that don't have
// specific implementations

// NewWindowsPlatformInstaller creates a stub Windows platform installer
func NewWindowsPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

// NewLinuxPlatformInstaller creates a stub Linux platform installer  
func NewLinuxPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

// NewDarwinPlatformInstaller creates a stub macOS platform installer
func NewDarwinPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

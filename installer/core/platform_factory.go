package core

import "runtime"

// CreatePlatformInstaller creates the appropriate platform installer
// This is separated from individual platform files to avoid build issues
func CreatePlatformInstaller(config *Config) PlatformInstaller {
	switch runtime.GOOS {
	case "windows":
		return createWindowsPlatformInstaller(config)
	case "linux":
		return createLinuxPlatformInstaller(config)
	case "darwin":
		return createDarwinPlatformInstaller(config)
	default:
		return NewDefaultPlatformInstaller(config)
	}
}

// Platform-specific factory functions are implemented in their respective files
// with build tags. On unsupported platforms, these fall back to default.

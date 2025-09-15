//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package core

// On unsupported platforms, all factories return default
func createWindowsPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

func createLinuxPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

func createDarwinPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

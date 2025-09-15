//go:build windows
// +build windows

package core

// On Windows, provide stubs for other platform factories
func createLinuxPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

func createDarwinPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

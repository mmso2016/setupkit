//go:build darwin
// +build darwin

package core

// On Darwin, provide stubs for other platform factories
func createWindowsPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

func createLinuxPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

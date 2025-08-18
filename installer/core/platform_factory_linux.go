//go:build linux
// +build linux

package core

// On Linux, provide stubs for other platform factories
func createWindowsPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

func createDarwinPlatformInstaller(config *Config) PlatformInstaller {
	return NewDefaultPlatformInstaller(config)
}

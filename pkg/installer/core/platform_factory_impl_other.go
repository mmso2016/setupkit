//go:build !windows
// +build !windows

package core

// createPlatformInstaller creates a platform installer for non-Windows systems
func createPlatformInstaller() (PlatformInstaller, error) {
	config := &Config{}
	return CreatePlatformInstaller(config), nil
}

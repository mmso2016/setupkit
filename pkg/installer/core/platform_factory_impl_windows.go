//go:build windows
// +build windows

package core

// createPlatformInstaller creates a platform installer for Windows
func createPlatformInstaller() (PlatformInstaller, error) {
	config := &Config{}
	// Try to create extended installer first
	if extended := NewExtendedPlatformInstaller(config); extended != nil {
		return extended, nil
	}
	// Fallback to basic Windows installer
	return &WindowsPlatformInstaller{config: config}, nil
}

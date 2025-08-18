//go:build !windows
// +build !windows

package core

import "fmt"

// UnixExtendedInstaller is a stub for non-Windows platforms
type UnixExtendedInstaller struct {
	PlatformInstaller
}

// CreateExtendedPlatformInstaller creates a basic platform installer for non-Windows
func CreateExtendedPlatformInstaller() (PlatformInstaller, error) {
	config := &Config{}
	return CreatePlatformInstaller(config), nil
}

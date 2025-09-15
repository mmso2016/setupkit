//go:build windows
// +build windows

package core

import (
	"fmt"
	"os"
	
	"golang.org/x/sys/windows/registry"
)

// WindowsExtendedInstaller extends WindowsPlatformInstaller with additional features
type WindowsExtendedInstaller struct {
	*WindowsPlatformInstaller
}

// Ensure WindowsPlatformInstaller implements ExtendedPlatformInstaller
var _ ExtendedPlatformInstaller = (*WindowsExtendedInstaller)(nil)

// NewExtendedPlatformInstaller creates an extended platform installer for Windows
func NewExtendedPlatformInstaller(config *Config) ExtendedPlatformInstaller {
	base := &WindowsPlatformInstaller{config: config}
	return &WindowsExtendedInstaller{
		WindowsPlatformInstaller: base,
	}
}

// CanElevate checks if elevation is possible on Windows
func (w *WindowsExtendedInstaller) CanElevate() bool {
	// On Windows, we can always attempt to elevate via UAC
	return true
}

// WriteRegistryString writes a string value to the Windows registry
func (w *WindowsExtendedInstaller) WriteRegistryString(keyPath, valueName, value string) error {
	// Default to HKEY_LOCAL_MACHINE
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.CREATE_SUB_KEY|registry.SET_VALUE)
	if err != nil {
		// Try HKEY_CURRENT_USER as fallback
		key, err = registry.OpenKey(registry.CURRENT_USER, keyPath, registry.CREATE_SUB_KEY|registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("failed to open registry key: %w", err)
		}
	}
	defer key.Close()
	
	return key.SetStringValue(valueName, value)
}

// DeleteRegistryValue deletes a value from the Windows registry
func (w *WindowsExtendedInstaller) DeleteRegistryValue(keyPath, valueName string) error {
	// Try HKEY_LOCAL_MACHINE first
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.SET_VALUE)
	if err != nil {
		// Try HKEY_CURRENT_USER as fallback
		key, err = registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("failed to open registry key: %w", err)
		}
	}
	defer key.Close()
	
	return key.DeleteValue(valueName)
}

// SetEnv sets an environment variable on Windows
func (w *WindowsExtendedInstaller) SetEnv(key, value string, system bool) error {
	if system {
		// System environment variable (requires admin)
		regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, 
			`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
			registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("failed to open system environment: %w", err)
		}
		defer regKey.Close()
		
		if err := regKey.SetStringValue(key, value); err != nil {
			return fmt.Errorf("failed to set system environment variable: %w", err)
		}
		
		// Notify the system about the change
		notifyEnvironmentChange()
	} else {
		// User environment variable
		regKey, err := registry.OpenKey(registry.CURRENT_USER,
			`Environment`,
			registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("failed to open user environment: %w", err)
		}
		defer regKey.Close()
		
		if err := regKey.SetStringValue(key, value); err != nil {
			return fmt.Errorf("failed to set user environment variable: %w", err)
		}
		
		// Notify the system about the change
		notifyEnvironmentChange()
	}
	
	return nil
}

// UnsetEnv removes an environment variable on Windows
func (w *WindowsExtendedInstaller) UnsetEnv(key string, system bool) error {
	if system {
		// System environment variable (requires admin)
		regKey, err := registry.OpenKey(registry.LOCAL_MACHINE,
			`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
			registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("failed to open system environment: %w", err)
		}
		defer regKey.Close()
		
		if err := regKey.DeleteValue(key); err != nil {
			return fmt.Errorf("failed to delete system environment variable: %w", err)
		}
		
		// Notify the system about the change
		notifyEnvironmentChange()
	} else {
		// User environment variable
		regKey, err := registry.OpenKey(registry.CURRENT_USER,
			`Environment`,
			registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("failed to open user environment: %w", err)
		}
		defer regKey.Close()
		
		if err := regKey.DeleteValue(key); err != nil {
			return fmt.Errorf("failed to delete user environment variable: %w", err)
		}
		
		// Notify the system about the change
		notifyEnvironmentChange()
	}
	
	return nil
}

// notifyEnvironmentChange broadcasts WM_SETTINGCHANGE to notify about environment changes
func notifyEnvironmentChange() {
	// This is a simplified version - full implementation would use SendMessageTimeout
	// For now, we'll rely on the user logging out/in or restarting
	
	// Try to set in current process at least
	os.Setenv("PATH", os.Getenv("PATH"))
}

// CreateExtendedPlatformInstaller creates an extended platform installer for Windows
func CreateExtendedPlatformInstaller() (PlatformInstaller, error) {
	config := &Config{}
	return NewExtendedPlatformInstaller(config), nil
}

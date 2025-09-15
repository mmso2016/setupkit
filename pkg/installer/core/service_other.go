//go:build !windows
// +build !windows

package core

import "fmt"

// UnixServiceManager is a stub implementation for non-Windows platforms
type UnixServiceManager struct{}

// newServiceManager creates a stub service manager for non-Windows platforms
func newServiceManager() (ServiceManager, error) {
	return &UnixServiceManager{}, nil
}

// Install is not implemented on this platform
func (m *UnixServiceManager) Install(config *ServiceConfig) error {
	return fmt.Errorf("service installation not implemented on this platform")
}

// Uninstall is not implemented on this platform
func (m *UnixServiceManager) Uninstall(name string) error {
	return fmt.Errorf("service uninstallation not implemented on this platform")
}

// Start is not implemented on this platform
func (m *UnixServiceManager) Start(name string) error {
	return fmt.Errorf("service start not implemented on this platform")
}

// Stop is not implemented on this platform
func (m *UnixServiceManager) Stop(name string) error {
	return fmt.Errorf("service stop not implemented on this platform")
}

// Status is not implemented on this platform
func (m *UnixServiceManager) Status(name string) (ServiceStatus, error) {
	return ServiceUnknown, fmt.Errorf("service status not implemented on this platform")
}

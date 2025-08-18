//go:build windows
// +build windows

package core

import (
	"fmt"
	
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// WindowsServiceManager implements ServiceManager for Windows
type WindowsServiceManager struct{}

// newServiceManager creates a Windows service manager
func newServiceManager() (ServiceManager, error) {
	return &WindowsServiceManager{}, nil
}

// Install installs a Windows service
func (m *WindowsServiceManager) Install(config *ServiceConfig) error {
	mgrHandle, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer mgrHandle.Disconnect()
	
	// Service start type constants
	const (
		StartManual    = 0x00000003
		StartAutomatic = 0x00000002
		StartDisabled  = 0x00000004
	)
	
	var startType uint32
	switch config.StartType {
	case ServiceStartAutomatic:
		startType = StartAutomatic
	case ServiceStartDisabled:
		startType = StartDisabled
	default:
		startType = StartManual
	}
	
	// Create service with basic configuration
	s, err := mgrHandle.CreateService(config.Name, config.Executable,
		mgr.Config{
			DisplayName: config.DisplayName,
			Description: config.Description,
			StartType:   startType,
		})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()
	
	return nil
}

// Uninstall removes a Windows service
func (m *WindowsServiceManager) Uninstall(name string) error {
	mgrHandle, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer mgrHandle.Disconnect()
	
	s, err := mgrHandle.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	
	return s.Delete()
}

// Start starts a Windows service
func (m *WindowsServiceManager) Start(name string) error {
	mgrHandle, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer mgrHandle.Disconnect()
	
	s, err := mgrHandle.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	
	return s.Start()
}

// Stop stops a Windows service
func (m *WindowsServiceManager) Stop(name string) error {
	mgrHandle, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer mgrHandle.Disconnect()
	
	s, err := mgrHandle.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	
	_, err = s.Control(svc.Stop)
	return err
}

// Status gets the status of a Windows service
func (m *WindowsServiceManager) Status(name string) (ServiceStatus, error) {
	mgrHandle, err := mgr.Connect()
	if err != nil {
		return ServiceUnknown, err
	}
	defer mgrHandle.Disconnect()
	
	s, err := mgrHandle.OpenService(name)
	if err != nil {
		return ServiceUnknown, err
	}
	defer s.Close()
	
	status, err := s.Query()
	if err != nil {
		return ServiceUnknown, err
	}
	
	switch status.State {
	case svc.Stopped:
		return ServiceStopped, nil
	case svc.StartPending:
		return ServiceStartPending, nil
	case svc.StopPending:
		return ServiceStopPending, nil
	case svc.Running:
		return ServiceRunning, nil
	case svc.ContinuePending:
		return ServiceContinuePending, nil
	case svc.PausePending:
		return ServicePausePending, nil
	case svc.Paused:
		return ServicePaused, nil
	default:
		return ServiceUnknown, nil
	}
}

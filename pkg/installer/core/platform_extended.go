package core

// ExtendedPlatformInstaller provides additional platform-specific operations
type ExtendedPlatformInstaller interface {
	PlatformInstaller
	
	// Advanced elevation
	CanElevate() bool
	
	// Registry operations (Windows)
	WriteRegistryString(key, valueName, value string) error
	DeleteRegistryValue(key, valueName string) error
	
	// Environment variables
	SetEnv(key, value string, system bool) error
	UnsetEnv(key string, system bool) error
}

// ServiceManager handles service installation and management
type ServiceManager interface {
	Install(config *ServiceConfig) error
	Uninstall(name string) error
	Start(name string) error
	Stop(name string) error
	Status(name string) (ServiceStatus, error)
}

// ServiceConfig defines service configuration
type ServiceConfig struct {
	Name        string
	DisplayName string
	Description string
	Executable  string
	Arguments   []string
	StartType   ServiceStartType
	RunAs       string
	RestartPolicy RestartPolicy
}

// ServiceStartType defines when the service should start
type ServiceStartType int

const (
	ServiceStartManual ServiceStartType = iota
	ServiceStartAutomatic
	ServiceStartDisabled
)

// RestartPolicy defines what happens when service fails
type RestartPolicy int

const (
	RestartNever RestartPolicy = iota
	RestartOnFailure
	RestartAlways
)

// ServiceStatus represents the current status of a service
type ServiceStatus int

const (
	ServiceStopped ServiceStatus = iota
	ServiceStartPending
	ServiceStopPending
	ServiceRunning
	ServiceContinuePending
	ServicePausePending
	ServicePaused
	ServiceUnknown
)

// GetPlatformInstaller returns a platform-specific installer
func GetPlatformInstaller() (PlatformInstaller, error) {
	return createPlatformInstaller()
}

// GetServiceManager returns a platform-specific service manager
func GetServiceManager() (ServiceManager, error) {
	return newServiceManager()
}

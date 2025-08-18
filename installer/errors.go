package installer

import "errors"

// Common errors
var (
	// ErrNotSupported indicates that an operation is not supported on the current platform
	ErrNotSupported = errors.New("operation not supported on this platform")

	// ErrElevationRequired indicates that administrator/root privileges are required
	ErrElevationRequired = errors.New("elevation required")

	// ErrAlreadyElevated indicates that the process is already running with elevated privileges
	ErrAlreadyElevated = errors.New("already elevated")

	// ErrPathNotFound indicates that a path entry was not found
	ErrPathNotFound = errors.New("path entry not found")

	// ErrPathAlreadyExists indicates that a path entry already exists
	ErrPathAlreadyExists = errors.New("path entry already exists")

	// ErrRegistryKeyNotFound indicates that a registry key was not found (Windows)
	ErrRegistryKeyNotFound = errors.New("registry key not found")

	// ErrRegistryAccessDenied indicates that registry access was denied (Windows)
	ErrRegistryAccessDenied = errors.New("registry access denied")
)

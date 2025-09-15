package installer

import "fmt"

// Exit codes for installer operations
const (
	// Success
	ExitSuccess = 0
	
	// General errors (1-19)
	ExitGeneralError    = 1
	ExitInvalidArgs     = 2
	ExitConfigError     = 3
	ExitPermissionError = 4
	ExitFileError       = 5
	ExitNetworkError    = 6
	ExitTimeout         = 7
	
	// Pre-installation checks (20-39)
	ExitPrereqFailed       = 20
	ExitInsufficientDisk   = 21
	ExitInsufficientRAM    = 22
	ExitOSUnsupported      = 23
	ExitPortInUse          = 24
	ExitDependencyMissing  = 25
	ExitIncompatibleVersion = 26
	ExitAlreadyInstalled   = 27
	ExitConflictingApp     = 28
	
	// Installation failures (40-59)
	ExitExtractFailed        = 40
	ExitCopyFailed           = 41
	ExitDBInitFailed         = 42
	ExitServiceInstallFailed = 43
	ExitConfigWriteFailed    = 44
	ExitPermissionSetFailed  = 45
	ExitSymlinkFailed        = 46
	ExitRegistryFailed       = 47
	
	// Post-installation failures (60-79)
	ExitServiceStartFailed = 60
	ExitHealthCheckFailed  = 61
	ExitBackupFailed       = 62
	ExitMigrationFailed    = 63
	ExitActivationFailed   = 64
	
	// Rollback status (80-89)
	ExitRollbackSuccess = 80  // Error occurred but successfully rolled back
	ExitRollbackFailed  = 81  // Critical - system in inconsistent state
	ExitRollbackPartial = 82  // Partial rollback completed
	
	// User actions (90-99)
	ExitUserCancelled   = 90
	ExitLicenseDeclined = 91
	ExitTimeoutUser     = 92
	ExitUserAbort       = 93
)

// Error types for categorizing failures
type InstallError struct {
	Code    int
	Message string
	Cause   error
	Details map[string]interface{}
}

func (e *InstallError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *InstallError) ExitCode() int {
	return e.Code
}

// NewError creates a new installation error
func NewError(code int, message string, cause error) *InstallError {
	return &InstallError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

// GetExitCodeForError extracts exit code from error
func GetExitCodeForError(err error) int {
	if err == nil {
		return ExitSuccess
	}
	
	if installErr, ok := err.(*InstallError); ok {
		return installErr.ExitCode()
	}
	
	// Default to general error
	return ExitGeneralError
}

// ExitCodeDescription returns a human-readable description of an exit code
func ExitCodeDescription(code int) string {
	descriptions := map[int]string{
		ExitSuccess:              "Installation completed successfully",
		ExitGeneralError:         "General error occurred",
		ExitInvalidArgs:          "Invalid arguments provided",
		ExitConfigError:          "Configuration error",
		ExitPermissionError:      "Permission denied",
		ExitFileError:            "File operation failed",
		ExitNetworkError:         "Network error occurred",
		ExitTimeout:              "Operation timed out",
		
		ExitPrereqFailed:         "Prerequisites check failed",
		ExitInsufficientDisk:     "Insufficient disk space",
		ExitInsufficientRAM:      "Insufficient memory",
		ExitOSUnsupported:        "Operating system not supported",
		ExitPortInUse:            "Required port is already in use",
		ExitDependencyMissing:    "Required dependency is missing",
		ExitIncompatibleVersion:  "Incompatible version detected",
		ExitAlreadyInstalled:     "Application is already installed",
		ExitConflictingApp:       "Conflicting application detected",
		
		ExitExtractFailed:        "Failed to extract files",
		ExitCopyFailed:           "Failed to copy files",
		ExitDBInitFailed:         "Database initialization failed",
		ExitServiceInstallFailed: "Service installation failed",
		ExitConfigWriteFailed:    "Failed to write configuration",
		ExitPermissionSetFailed:  "Failed to set permissions",
		ExitSymlinkFailed:        "Failed to create symbolic links",
		ExitRegistryFailed:       "Registry operation failed",
		
		ExitServiceStartFailed:   "Failed to start service",
		ExitHealthCheckFailed:    "Health check failed",
		ExitBackupFailed:         "Backup operation failed",
		ExitMigrationFailed:      "Migration failed",
		ExitActivationFailed:     "Activation failed",
		
		ExitRollbackSuccess:      "Error occurred but rollback succeeded",
		ExitRollbackFailed:       "Rollback failed - system may be inconsistent",
		ExitRollbackPartial:      "Partial rollback completed",
		
		ExitUserCancelled:        "Installation cancelled by user",
		ExitLicenseDeclined:      "License agreement declined",
		ExitTimeoutUser:          "User response timeout",
		ExitUserAbort:            "Installation aborted by user",
	}
	
	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return fmt.Sprintf("Unknown exit code: %d", code)
}

// Package components provides reusable installer components
package components

import (
	"context"
	"fmt"
	"runtime"
	
	"github.com/mmso2016/setupkit/pkg/installer"
)

// PathScope defines the scope for PATH modifications
type PathScope int

const (
	// PathScopeAuto automatically selects based on privileges
	PathScopeAuto PathScope = iota
	// PathScopeUser modifies user PATH
	PathScopeUser
	// PathScopeSystem modifies system PATH (requires elevation)
	PathScopeSystem
)

// PathComponent handles PATH environment variable modifications
type PathComponent struct {
	installer.Component
	Directory     string
	Scope         PathScope
	SkipDuplicate bool
}

// NewPathComponent creates a new PATH management component
func NewPathComponent(directory string, scope PathScope) *PathComponent {
	pc := &PathComponent{
		Directory:     directory,
		Scope:         scope,
		SkipDuplicate: true,
	}
	
	pc.Component = installer.Component{
		ID:          "path-configuration",
		Name:        "PATH Environment Variable",
		Description: fmt.Sprintf("Add %s to PATH", directory),
		Required:    false,
		Selected:    true,
		Validator:   pc.validate,
		Installer:   pc.install,
		Uninstaller: pc.uninstall,
	}
	
	return pc
}

// validate checks if PATH modification is possible
func (pc *PathComponent) validate() error {
	// Check if directory exists or will be created
	// This is just validation, actual directory might be created by another component
	
	if pc.Directory == "" {
		return fmt.Errorf("directory cannot be empty")
	}
	
	// Platform-specific validation
	switch runtime.GOOS {
	case "windows":
		// Check for invalid characters in Windows paths
		// Check path length limits
	case "linux", "darwin":
		// Check for proper Unix path format
	}
	
	return nil
}

// install adds the directory to PATH
func (pc *PathComponent) install(ctx context.Context) error {
	// Get platform installer from context
	platform, ok := ctx.Value("platform").(installer.PlatformInstaller)
	if !ok {
		return fmt.Errorf("platform installer not available in context")
	}
	
	// Get logger from context
	logger, ok := ctx.Value("logger").(installer.Logger)
	if !ok {
		// Use a no-op logger if not available
		logger = installer.NewLogger("error", "")
	}
	
	// Determine actual scope based on configuration and privileges
	useSystemPath := pc.shouldUseSystemPath(platform)
	
	logger.Verbose("Installing PATH component", 
		"directory", pc.Directory,
		"scope", pc.getScopeString(),
		"system", useSystemPath)
	
	// Check if already in PATH
	if pc.SkipDuplicate {
		if platform.IsInPath(pc.Directory, useSystemPath) {
			logger.Info("Directory already in PATH, skipping", "directory", pc.Directory)
			return nil
		}
	}
	
	// Add to PATH
	if err := platform.AddToPath(pc.Directory, useSystemPath); err != nil {
		// Check if it's an elevation error
		if err == installer.ErrElevationRequired {
			if pc.Scope == PathScopeSystem {
				// User explicitly requested system PATH but lacks privileges
				return fmt.Errorf("administrator privileges required to modify system PATH")
			}
			// Fall back to user PATH
			logger.Verbose("Falling back to user PATH due to insufficient privileges")
			if err := platform.AddToPath(pc.Directory, false); err != nil {
				return fmt.Errorf("failed to add to user PATH: %w", err)
			}
			logger.Info("Added to user PATH", "directory", pc.Directory)
			return nil
		}
		return fmt.Errorf("failed to add to PATH: %w", err)
	}
	
	if useSystemPath {
		logger.Info("Added to system PATH", "directory", pc.Directory)
	} else {
		logger.Info("Added to user PATH", "directory", pc.Directory)
	}
	
	return nil
}

// uninstall removes the directory from PATH
func (pc *PathComponent) uninstall(ctx context.Context) error {
	// Get platform installer from context
	platform, ok := ctx.Value("platform").(installer.PlatformInstaller)
	if !ok {
		return fmt.Errorf("platform installer not available in context")
	}
	
	// Get logger from context
	logger, ok := ctx.Value("logger").(installer.Logger)
	if !ok {
		logger = installer.NewLogger("error", "")
	}
	
	// Try to remove from both user and system PATH
	// This handles cases where the installation scope might have changed
	
	removedUser := false
	removedSystem := false
	
	// Try user PATH first (doesn't require elevation)
	if platform.IsInPath(pc.Directory, false) {
		if err := platform.RemoveFromPath(pc.Directory, false); err != nil {
			logger.Warn("Failed to remove from user PATH", "error", err)
		} else {
			removedUser = true
			logger.Info("Removed from user PATH", "directory", pc.Directory)
		}
	}
	
	// Try system PATH if we have privileges
	if platform.IsElevated() && platform.IsInPath(pc.Directory, true) {
		if err := platform.RemoveFromPath(pc.Directory, true); err != nil {
			logger.Warn("Failed to remove from system PATH", "error", err)
		} else {
			removedSystem = true
			logger.Info("Removed from system PATH", "directory", pc.Directory)
		}
	}
	
	if !removedUser && !removedSystem {
		logger.Verbose("Directory was not found in PATH", "directory", pc.Directory)
	}
	
	return nil
}

// shouldUseSystemPath determines whether to use system or user PATH
func (pc *PathComponent) shouldUseSystemPath(platform installer.PlatformInstaller) bool {
	switch pc.Scope {
	case PathScopeSystem:
		// Explicitly requested system PATH
		return platform.IsElevated()
	case PathScopeUser:
		// Explicitly requested user PATH
		return false
	case PathScopeAuto:
		// Auto-detect based on privileges and platform
		if runtime.GOOS == "windows" {
			// On Windows, use system PATH if we have admin privileges
			return platform.IsElevated()
		}
		// On Unix-like systems, generally prefer user PATH unless root
		return platform.IsElevated() && runtime.GOOS != "darwin"
	default:
		return false
	}
}

// getScopeString returns a string representation of the scope
func (pc *PathComponent) getScopeString() string {
	switch pc.Scope {
	case PathScopeAuto:
		return "auto"
	case PathScopeUser:
		return "user"
	case PathScopeSystem:
		return "system"
	default:
		return "unknown"
	}
}

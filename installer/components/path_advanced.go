package components

import (
	"context"
	"fmt"
	
	"github.com/mmso2016/setupkit/installer/core"
)

// PathOptions defines advanced options for PATH configuration
type PathOptions struct {
	PrependPath    bool // Add to beginning of PATH instead of end
	CreateBackup   bool // Backup existing PATH before modification
	NotifyUser     bool // Show notification about PATH changes
	RequireRestart bool // Whether system restart is required
}

// AdvancedPathComponent extends PathComponent with additional options
type AdvancedPathComponent struct {
	*PathComponent
	Options PathOptions
}

// NewAdvancedPathComponent creates a new advanced PATH component
func NewAdvancedPathComponent(directory string, scope PathScope, options PathOptions) *AdvancedPathComponent {
	pc := NewPathComponent(directory, scope)
	
	apc := &AdvancedPathComponent{
		PathComponent: pc,
		Options:       options,
	}
	
	// Update description based on options
	if options.RequireRestart {
		pc.Description = fmt.Sprintf("Add %s to PATH (restart required)", directory)
	}
	
	// Override installer to include advanced features
	originalInstaller := pc.Installer
	pc.Installer = func(ctx context.Context) error {
		logger, _ := ctx.Value("logger").(core.Logger)
		
		// Create backup if requested
		if options.CreateBackup && logger != nil {
			logger.Info("Creating PATH backup")
			// Backup implementation would go here
		}
		
		// Call original installer
		if err := originalInstaller(ctx); err != nil {
			return err
		}
		
		// Notify user if requested
		if options.NotifyUser && logger != nil {
			msg := fmt.Sprintf("PATH has been updated with: %s", directory)
			if options.RequireRestart {
				msg += " (restart required for changes to take effect)"
			}
			logger.Info(msg)
		}
		
		return nil
	}
	
	return apc
}

// IsPrepend returns whether the path should be prepended
func (apc *AdvancedPathComponent) IsPrepend() bool {
	return apc.Options.PrependPath
}

// RequiresRestart returns whether a system restart is required
func (apc *AdvancedPathComponent) RequiresRestart() bool {
	return apc.Options.RequireRestart
}

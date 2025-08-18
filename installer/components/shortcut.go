package components

import (
	"context"
	"fmt"
	"path/filepath"
	
	"github.com/mmso2016/setupkit/installer/core"
)

// ShortcutOptions defines options for creating shortcuts
type ShortcutOptions struct {
	CreateDesktop     bool
	CreateStartMenu   bool
	CreateQuickLaunch bool
	IconPath          string
	Arguments         string
	Description       string
	WorkingDir        string
}

// ShortcutComponent represents a shortcut component
type ShortcutComponent struct {
	core.Component
	ShortcutName string // The name of the shortcut
	TargetPath   string
	Options      ShortcutOptions
}

// NewShortcutComponent creates a new shortcut component
func NewShortcutComponent(name string, targetPath string, options ShortcutOptions) *ShortcutComponent {
	sc := &ShortcutComponent{
		ShortcutName: name,
		TargetPath:   targetPath,
		Options:      options,
	}
	
	sc.Component = core.Component{
		ID:          fmt.Sprintf("shortcut-%s", name),
		Name:        fmt.Sprintf("Shortcuts: %s", name),
		Description: fmt.Sprintf("Create shortcuts for %s", name),
		Required:    false,
		Selected:    true,
		Validator:   sc.validate,
		Installer:   sc.install,
		Uninstaller: sc.uninstall,
	}
	
	return sc
}

func (sc *ShortcutComponent) validate() error {
	if sc.ShortcutName == "" {
		return fmt.Errorf("shortcut name cannot be empty")
	}
	
	if sc.TargetPath == "" {
		return fmt.Errorf("target path cannot be empty")
	}
	
	return nil
}

func (sc *ShortcutComponent) install(ctx context.Context) error {
	platform, ok := ctx.Value("platform").(core.PlatformInstaller)
	if !ok {
		return fmt.Errorf("platform installer not found in context")
	}
	
	logger, _ := ctx.Value("logger").(core.Logger)
	
	// Log installation
	if logger != nil {
		logger.Info("Creating shortcuts", 
			"name", sc.ShortcutName,
			"target", sc.TargetPath)
	}
	
	// Platform-specific shortcut creation would go here
	// For now, we just call the platform's CreateShortcuts method
	if err := platform.CreateShortcuts(); err != nil {
		return fmt.Errorf("failed to create shortcuts: %w", err)
	}
	
	return nil
}

func (sc *ShortcutComponent) uninstall(ctx context.Context) error {
	logger, _ := ctx.Value("logger").(core.Logger)
	
	// Log removal
	if logger != nil {
		logger.Info("Removing shortcuts", "name", sc.ShortcutName)
	}
	
	// Platform-specific shortcut removal would go here
	// This would typically remove desktop, start menu, and quick launch shortcuts
	
	return nil
}

// GetShortcutPath returns the path where the shortcut will be created
func (sc *ShortcutComponent) GetShortcutPath(location string) string {
	switch location {
	case "desktop":
		// Would return desktop path
		return filepath.Join("Desktop", sc.ShortcutName+".lnk")
	case "startmenu":
		// Would return start menu path
		return filepath.Join("Start Menu", "Programs", sc.ShortcutName+".lnk")
	case "quicklaunch":
		// Would return quick launch path
		return filepath.Join("Quick Launch", sc.ShortcutName+".lnk")
	default:
		return ""
	}
}

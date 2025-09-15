package components

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// ConfigComponent represents a configuration file component
type ConfigComponent struct {
	core.Component
	FileName    string
	DestDir     string
	Content     string
	Overwrite   bool
	Permissions os.FileMode
}

// NewConfigComponent creates a new configuration component
func NewConfigComponent(fileName, destDir, content string) *ConfigComponent {
	cc := &ConfigComponent{
		FileName:    fileName,
		DestDir:     destDir,
		Content:     content,
		Overwrite:   false,
		Permissions: 0644,
	}
	
	cc.Component = core.Component{
		ID:          fmt.Sprintf("config-%s", fileName),
		Name:        fmt.Sprintf("Configuration: %s", fileName),
		Description: fmt.Sprintf("Install configuration file %s", fileName),
		Required:    false,
		Selected:    true,
		Validator:   cc.validate,
		Installer:   cc.install,
		Uninstaller: cc.uninstall,
	}
	
	return cc
}

func (cc *ConfigComponent) validate() error {
	if cc.FileName == "" {
		return fmt.Errorf("configuration file name cannot be empty")
	}
	
	if cc.DestDir == "" {
		return fmt.Errorf("destination directory cannot be empty")
	}
	
	return nil
}

func (cc *ConfigComponent) install(ctx context.Context) error {
	logger, _ := ctx.Value("logger").(core.Logger)
	
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(cc.DestDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	destPath := filepath.Join(cc.DestDir, cc.FileName)
	
	// Check if file exists and whether to overwrite
	if _, err := os.Stat(destPath); err == nil && !cc.Overwrite {
		// File exists and we shouldn't overwrite
		if logger != nil {
			logger.Info("Config file already exists, skipping", 
				"file", destPath)
		}
		return nil
	}
	
	// Write configuration file
	if err := os.WriteFile(destPath, []byte(cc.Content), cc.Permissions); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	if logger != nil {
		logger.Info("Configuration file installed", 
			"file", destPath,
			"permissions", cc.Permissions)
	}
	
	return nil
}

func (cc *ConfigComponent) uninstall(ctx context.Context) error {
	logger, _ := ctx.Value("logger").(core.Logger)
	
	destPath := filepath.Join(cc.DestDir, cc.FileName)
	
	// Remove configuration file
	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}
	
	if logger != nil {
		logger.Info("Configuration file removed", "file", destPath)
	}
	
	// Try to remove directory if empty
	_ = os.Remove(cc.DestDir)
	
	return nil
}

// SetPermissions sets custom permissions for the config file
func (cc *ConfigComponent) SetPermissions(perms os.FileMode) {
	cc.Permissions = perms
}

// SetOverwrite sets whether to overwrite existing config files
func (cc *ConfigComponent) SetOverwrite(overwrite bool) {
	cc.Overwrite = overwrite
}

package components

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// BinaryOptions provides options for binary installation
type BinaryOptions struct {
	ExecutableName string
	Permissions    os.FileMode
	StripDebug     bool
}

// BinaryComponent installs a binary executable
type BinaryComponent struct {
	core.Component
	SourcePath string
	DestDir    string
	Options    BinaryOptions
	assets     fs.FS
}

// NewBinaryComponent creates a new binary installation component
func NewBinaryComponent(sourcePath string, destDir string, opts BinaryOptions) *BinaryComponent {
	bc := &BinaryComponent{
		SourcePath: sourcePath,
		DestDir:    destDir,
		Options:    opts,
	}

	if opts.Permissions == 0 {
		bc.Options.Permissions = 0755
	}

	if opts.ExecutableName == "" {
		bc.Options.ExecutableName = filepath.Base(sourcePath)
		if runtime.GOOS == "windows" && filepath.Ext(bc.Options.ExecutableName) == "" {
			bc.Options.ExecutableName += ".exe"
		}
	}

	bc.Component = core.Component{
		ID:          "binary-" + strings.ReplaceAll(bc.Options.ExecutableName, ".", "_"),
		Name:        "Binary: " + bc.Options.ExecutableName,
		Description: fmt.Sprintf("Install %s to %s", bc.Options.ExecutableName, destDir),
		Required:    true,
		Selected:    true,
		Validator:   bc.validate,
		Installer:   bc.install,
		Uninstaller: bc.uninstall,
	}

	return bc
}

func (bc *BinaryComponent) validate() error {
	if bc.SourcePath == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	if bc.DestDir == "" {
		return fmt.Errorf("destination directory cannot be empty")
	}
	return nil
}

func (bc *BinaryComponent) install(ctx context.Context) error {
	logger, _ := ctx.Value("logger").(core.Logger)
	if logger != nil {
		logger.Verbose("Installing binary",
			"source", bc.SourcePath,
			"dest", bc.DestDir,
			"name", bc.Options.ExecutableName)
	}

	// Create destination directory
	if err := os.MkdirAll(bc.DestDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read binary data
	var data []byte
	var err error

	// Check if we have assets FS from context
	if assets, ok := ctx.Value("assets").(fs.FS); ok {
		bc.assets = assets
	}

	if bc.assets != nil {
		// Try to read from embedded assets
		file, err := bc.assets.Open(bc.SourcePath)
		if err == nil {
			defer file.Close()
			data, err = io.ReadAll(file)
			if err != nil {
				return fmt.Errorf("failed to read from assets: %w", err)
			}
		}
	}

	if data == nil {
		// Try to read from filesystem
		data, err = os.ReadFile(bc.SourcePath)
		if err != nil {
			return fmt.Errorf("failed to read binary: %w", err)
		}
	}

	// Write to destination
	destPath := filepath.Join(bc.DestDir, bc.Options.ExecutableName)
	if err := os.WriteFile(destPath, data, bc.Options.Permissions); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}

	if logger != nil {
		logger.Info("Binary installed", "path", destPath)
	}

	return nil
}

func (bc *BinaryComponent) uninstall(ctx context.Context) error {
	destPath := filepath.Join(bc.DestDir, bc.Options.ExecutableName)

	logger, _ := ctx.Value("logger").(core.Logger)
	if logger != nil {
		logger.Verbose("Uninstalling binary", "path", destPath)
	}

	if err := os.Remove(destPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
	}

	return nil
}

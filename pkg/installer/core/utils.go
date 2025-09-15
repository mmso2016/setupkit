package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// CheckDiskSpace checks if there is enough disk space available
func CheckDiskSpace(path string, required int64) error {
	// Create path if it doesn't exist (temporarily, for space check)
	dir := filepath.Dir(path)

	// Get disk usage statistics
	availableSpace, err := getAvailableSpace(dir)
	if err != nil {
		return fmt.Errorf("failed to check disk space: %w", err)
	}

	if availableSpace < required {
		return fmt.Errorf("insufficient disk space: required %d bytes, available %d bytes",
			required, availableSpace)
	}

	return nil
}

// ExtractAssets extracts embedded assets to the target directory
func ExtractAssets(assets fs.FS, targetDir string) error {
	return fs.WalkDir(assets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == "." {
			return nil
		}

		targetPath := filepath.Join(targetDir, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read file from embedded FS
		data, err := fs.ReadFile(assets, path)
		if err != nil {
			return err
		}

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Write file
		return os.WriteFile(targetPath, data, 0644)
	})
}

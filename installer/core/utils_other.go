//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package core

import (
	"fmt"
)

// getAvailableSpace returns a default implementation for unsupported platforms
func getAvailableSpace(path string) (int64, error) {
	// Return a large number to allow installation to proceed
	// on platforms where we can't determine disk space
	return 1 << 40, fmt.Errorf("disk space check not supported on this platform")
}

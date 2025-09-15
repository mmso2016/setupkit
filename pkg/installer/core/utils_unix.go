//go:build linux || darwin
// +build linux darwin

package core

import (
	"syscall"
)

// getAvailableSpace returns available disk space on Unix-like systems
func getAvailableSpace(path string) (int64, error) {
	var stat syscall.Statfs_t
	
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	
	// Available space = block size * available blocks
	available := int64(stat.Bavail) * int64(stat.Bsize)
	
	return available, nil
}

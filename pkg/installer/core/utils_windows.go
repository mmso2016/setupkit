//go:build windows
// +build windows

package core

import (
	"syscall"
	"unsafe"
)

// getAvailableSpace returns available disk space on Windows
func getAvailableSpace(path string) (int64, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")
	
	var freeBytesAvailable int64
	var totalNumberOfBytes int64
	var totalNumberOfFreeBytes int64
	
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	
	ret, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	
	if ret == 0 {
		return 0, err
	}
	
	return freeBytesAvailable, nil
}

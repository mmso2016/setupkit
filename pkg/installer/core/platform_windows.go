//go:build windows
// +build windows

package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// WindowsPlatformInstaller implements PlatformInstaller for Windows
type WindowsPlatformInstaller struct {
	config *Config
}

// createWindowsPlatformInstaller is the internal factory function
func createWindowsPlatformInstaller(config *Config) PlatformInstaller {
	return &WindowsPlatformInstaller{
		config: config,
	}
}

// NewWindowsPlatformInstaller creates a Windows platform installer
func NewWindowsPlatformInstaller(config *Config) PlatformInstaller {
	return &WindowsPlatformInstaller{
		config: config,
	}
}

func (w *WindowsPlatformInstaller) Initialize() error {
	// Windows-specific initialization
	return nil
}

func (w *WindowsPlatformInstaller) CheckRequirements() error {
	// Check Windows version
	if !isWindows10OrLater() {
		return fmt.Errorf("only for Windows 10 or later (is required)")
	}

	// Check .NET Framework if needed
	// Check Visual C++ Redistributables if needed

	return nil
}

func (w *WindowsPlatformInstaller) IsElevated() bool {
	// Check if running as administrator
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}

func (w *WindowsPlatformInstaller) RequiresElevation() bool {
	// Check if installation path requires admin rights
	installPath := w.config.InstallDir

	// Program Files requires admin
	if strings.HasPrefix(strings.ToLower(installPath), "c:\\program files") {
		return true
	}

	// System directories require admin
	systemDirs := []string{
		"c:\\windows",
		"c:\\programdata",
	}

	for _, dir := range systemDirs {
		if strings.HasPrefix(strings.ToLower(installPath), dir) {
			return true
		}
	}

	// Check if we can write to the directory
	testFile := filepath.Join(installPath, ".write_test")
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return true
	}

	file, err := os.Create(testFile)
	if err != nil {
		return true
	}
	file.Close()
	os.Remove(testFile)

	return false
}

func (w *WindowsPlatformInstaller) RequestElevation() error {
	if w.IsElevated() {
		return nil
	}

	// Get current executable path
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// Prepare ShellExecute parameters
	verb := "runas"
	args := strings.Join(os.Args[1:], " ")

	// Use ShellExecute to restart with elevation
	err = windows.ShellExecute(
		0,
		syscall.StringToUTF16Ptr(verb),
		syscall.StringToUTF16Ptr(exe),
		syscall.StringToUTF16Ptr(args),
		nil,
		windows.SW_NORMAL,
	)

	if err != nil {
		return fmt.Errorf("failed to request elevation: %w", err)
	}

	// Exit current process
	os.Exit(0)
	return nil
}

func (w *WindowsPlatformInstaller) RegisterWithOS() error {
	// Add to Windows registry for Add/Remove Programs
	keyPath := fmt.Sprintf(`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\%s`, w.config.AppName)

	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, keyPath, registry.ALL_ACCESS)
	if err != nil {
		// Try current user if local machine fails
		key, _, err = registry.CreateKey(registry.CURRENT_USER, keyPath, registry.ALL_ACCESS)
		if err != nil {
			return fmt.Errorf("failed to create registry key: %w", err)
		}
	}
	defer key.Close()

	// Set registry values
	key.SetStringValue("DisplayName", w.config.AppName)
	key.SetStringValue("DisplayVersion", w.config.Version)
	key.SetStringValue("Publisher", w.config.Publisher)
	key.SetStringValue("InstallLocation", w.config.InstallDir)
	key.SetStringValue("UninstallString", filepath.Join(w.config.InstallDir, "uninstall.exe"))

	// Set install date
	key.SetStringValue("InstallDate", time.Now().Format("20060102"))

	// Estimate size (in KB)
	var totalSize int64
	for _, comp := range w.config.Components {
		if comp.Selected || comp.Required {
			totalSize += comp.Size
		}
	}
	key.SetDWordValue("EstimatedSize", uint32(totalSize/1024))

	return nil
}

func (w *WindowsPlatformInstaller) CreateShortcuts() error {
	// Create Start Menu shortcuts
	startMenuPath := os.Getenv("APPDATA")
	if startMenuPath == "" {
		return fmt.Errorf("could not determine Start Menu path")
	}

	startMenuDir := filepath.Join(startMenuPath, "Microsoft", "Windows", "Start Menu", "Programs", w.config.AppName)
	if err := os.MkdirAll(startMenuDir, 0755); err != nil {
		return fmt.Errorf("failed to create Start Menu directory: %w", err)
	}

	// Create shortcut using PowerShell
	shortcutPath := filepath.Join(startMenuDir, w.config.AppName+".lnk")
	targetPath := filepath.Join(w.config.InstallDir, w.config.AppName+".exe")

	psScript := fmt.Sprintf(`
		$WshShell = New-Object -comObject WScript.Shell
		$Shortcut = $WshShell.CreateShortcut("%s")
		$Shortcut.TargetPath = "%s"
		$Shortcut.WorkingDirectory = "%s"
		$Shortcut.IconLocation = "%s,0"
		$Shortcut.Save()
	`, shortcutPath, targetPath, w.config.InstallDir, targetPath)

	cmd := exec.Command("powershell", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create shortcut: %w", err)
	}

	// Optionally create Desktop shortcut
	desktopPath := filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
	if desktopPath != "" {
		desktopShortcut := filepath.Join(desktopPath, w.config.AppName+".lnk")
		psScript = fmt.Sprintf(`
			$WshShell = New-Object -comObject WScript.Shell
			$Shortcut = $WshShell.CreateShortcut("%s")
			$Shortcut.TargetPath = "%s"
			$Shortcut.WorkingDirectory = "%s"
			$Shortcut.IconLocation = "%s,0"
			$Shortcut.Save()
		`, desktopShortcut, targetPath, w.config.InstallDir, targetPath)

		cmd = exec.Command("powershell", "-Command", psScript)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.Run() // Ignore error for desktop shortcut
	}

	return nil
}

func (w *WindowsPlatformInstaller) RegisterUninstaller() error {
	// Create uninstaller executable
	// uninstallerPath := filepath.Join(w.config.InstallDir, "uninstall.exe")
	// TODO: Implement uninstaller registration

	// For now, we'll create a simple batch file
	// In production, you'd want to create a proper uninstaller
	uninstallBat := filepath.Join(w.config.InstallDir, "uninstall.bat")

	content := fmt.Sprintf(`@echo off
echo Uninstalling %s...
rmdir /s /q "%s"
reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\%s" /f 2>nul
reg delete "HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\%s" /f 2>nul
echo Uninstallation complete.
pause
`, w.config.AppName, w.config.InstallDir, w.config.AppName, w.config.AppName)

	return os.WriteFile(uninstallBat, []byte(content), 0755)
}

func (w *WindowsPlatformInstaller) UpdatePath(dirs []string, system bool) error {
	// Determine registry root
	var root registry.Key
	var subkey string

	if system {
		root = registry.LOCAL_MACHINE
		subkey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	} else {
		root = registry.CURRENT_USER
		subkey = `Environment`
	}

	// Open registry key
	key, err := registry.OpenKey(root, subkey, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Get current PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		currentPath = ""
	}

	// Parse existing paths
	existingPaths := strings.Split(currentPath, ";")
	pathMap := make(map[string]bool)
	for _, p := range existingPaths {
		if p != "" {
			pathMap[strings.ToLower(p)] = true
		}
	}

	// Add new directories
	for _, dir := range dirs {
		fullPath := filepath.Join(w.config.InstallDir, dir)
		if !pathMap[strings.ToLower(fullPath)] {
			if currentPath != "" && !strings.HasSuffix(currentPath, ";") {
				currentPath += ";"
			}
			currentPath += fullPath
		}
	}

	// Set new PATH
	if err := key.SetStringValue("Path", currentPath); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}

	// Broadcast environment change
	broadcastEnvironmentChange()

	return nil
}

// Helper functions

func isWindows10OrLater() bool {
	// Check Windows version
	ver := windows.RtlGetVersion()
	// Windows 10 is version 10.0 (major version 10)
	return ver.MajorVersion >= 10
}

func broadcastEnvironmentChange() {
	// Notify all windows about environment change
	// Note: SendMessage is not available in syscall package
	// We need to use Windows API directly

	// Define constants
	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
	)

	// Load user32.dll
	user32 := windows.NewLazySystemDLL("user32.dll")
	sendMessageW := user32.NewProc("SendMessageW")

	// Send the message
	envPtr, _ := windows.UTF16PtrFromString("Environment")
	sendMessageW.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(envPtr)),
	)
}

// AddToPath adds a directory to the PATH environment variable
func (w *WindowsPlatformInstaller) AddToPath(dir string, system bool) error {
	// Determine registry root
	var root registry.Key
	var subkey string

	if system {
		root = registry.LOCAL_MACHINE
		subkey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	} else {
		root = registry.CURRENT_USER
		subkey = `Environment`
	}

	// Open registry key
	key, err := registry.OpenKey(root, subkey, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Get current PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to read PATH: %w", err)
	}

	// Check if already in PATH
	if w.isInPathString(currentPath, dir) {
		return nil // Already in PATH
	}

	// Add to PATH
	if currentPath != "" && !strings.HasSuffix(currentPath, ";") {
		currentPath += ";"
	}
	currentPath += dir

	// Set new PATH
	if err := key.SetStringValue("Path", currentPath); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}

	// Broadcast environment change
	broadcastEnvironmentChange()

	return nil
}

// RemoveFromPath removes a directory from the PATH environment variable
func (w *WindowsPlatformInstaller) RemoveFromPath(dir string, system bool) error {
	// Determine registry root
	var root registry.Key
	var subkey string

	if system {
		root = registry.LOCAL_MACHINE
		subkey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	} else {
		root = registry.CURRENT_USER
		subkey = `Environment`
	}

	// Open registry key
	key, err := registry.OpenKey(root, subkey, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Get current PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		return fmt.Errorf("failed to read PATH: %w", err)
	}

	// Remove from PATH
	paths := strings.Split(currentPath, ";")
	var newPaths []string
	for _, p := range paths {
		cleanPath := strings.TrimSpace(p)
		if !strings.EqualFold(cleanPath, dir) && cleanPath != "" {
			newPaths = append(newPaths, cleanPath)
		}
	}
	newPath := strings.Join(newPaths, ";")

	// Set new PATH
	if err := key.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}

	// Broadcast environment change
	broadcastEnvironmentChange()

	return nil
}

// IsInPath checks if a directory is in the PATH environment variable
func (w *WindowsPlatformInstaller) IsInPath(dir string, system bool) bool {
	// Determine registry root
	var root registry.Key
	var subkey string

	if system {
		root = registry.LOCAL_MACHINE
		subkey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	} else {
		root = registry.CURRENT_USER
		subkey = `Environment`
	}

	// Open registry key
	key, err := registry.OpenKey(root, subkey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	// Get current PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		return false
	}

	return w.isInPathString(currentPath, dir)
}

// isInPathString checks if a directory is in a PATH string
func (w *WindowsPlatformInstaller) isInPathString(pathStr, dir string) bool {
	paths := strings.Split(pathStr, ";")
	for _, p := range paths {
		cleanPath := strings.TrimSpace(p)
		if strings.EqualFold(cleanPath, dir) {
			return true
		}
	}
	return false
}

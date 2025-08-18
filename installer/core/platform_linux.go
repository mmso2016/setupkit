//go:build linux
// +build linux

package core

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LinuxPlatformInstaller implements PlatformInstaller for Linux
type LinuxPlatformInstaller struct {
	config *Config
}

// createLinuxPlatformInstaller is the internal factory function
func createLinuxPlatformInstaller(config *Config) PlatformInstaller {
	return &LinuxPlatformInstaller{
		config: config,
	}
}

// NewLinuxPlatformInstaller creates a Linux platform installer
func NewLinuxPlatformInstaller(config *Config) PlatformInstaller {
	return &LinuxPlatformInstaller{
		config: config,
	}
}

func (l *LinuxPlatformInstaller) Initialize() error {
	return nil
}

func (l *LinuxPlatformInstaller) CheckRequirements() error {
	// Check for required tools
	requiredCommands := []string{"chmod", "ln"}
	
	for _, cmd := range requiredCommands {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("required command '%s' not found", cmd)
		}
	}
	
	// Check glibc version if needed
	// Check other dependencies
	
	return nil
}

func (l *LinuxPlatformInstaller) IsElevated() bool {
	return os.Geteuid() == 0
}

func (l *LinuxPlatformInstaller) RequiresElevation() bool {
	// Check if installation path requires root
	systemPaths := []string{
		"/usr",
		"/opt",
		"/etc",
		"/var",
	}
	
	for _, path := range systemPaths {
		if strings.HasPrefix(l.config.InstallDir, path) {
			return true
		}
	}
	
	// Check if we can write to the directory
	if err := os.MkdirAll(l.config.InstallDir, 0755); err != nil {
		// Check if it's a permission error
		if os.IsPermission(err) {
			return true
		}
	}
	
	// Test write permission
	testFile := filepath.Join(l.config.InstallDir, ".write_test")
	file, err := os.Create(testFile)
	if err != nil {
		if os.IsPermission(err) {
			return true
		}
	} else {
		file.Close()
		os.Remove(testFile)
	}
	
	return false
}

func (l *LinuxPlatformInstaller) RequestElevation() error {
	if l.IsElevated() {
		return nil
	}
	
	// Try to use sudo
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Check if sudo is available
	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		// Try pkexec as fallback
		pkexecPath, err := exec.LookPath("pkexec")
		if err != nil {
			return fmt.Errorf("neither sudo nor pkexec found, cannot elevate privileges")
		}
		
		// Use pkexec
		cmd := exec.Command(pkexecPath, exe)
		cmd.Args = append(cmd.Args, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("elevation failed: %w", err)
		}
		os.Exit(0)
	}
	
	// Use sudo
	cmd := exec.Command(sudoPath, exe)
	cmd.Args = append(cmd.Args, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("elevation failed: %w", err)
	}
	
	os.Exit(0)
	return nil
}

func (l *LinuxPlatformInstaller) RegisterWithOS() error {
	// Create .desktop file for desktop environments
	desktopFile := fmt.Sprintf(`[Desktop Entry]
Version=1.0
Type=Application
Name=%s
Comment=%s Application
Exec=%s
Icon=%s
Terminal=false
Categories=Application;
`, l.config.AppName, l.config.AppName, 
   filepath.Join(l.config.InstallDir, l.config.AppName),
   filepath.Join(l.config.InstallDir, l.config.AppName+".png"))
	
	// Determine desktop file location
	var desktopPath string
	
	// System-wide installation
	if l.IsElevated() {
		desktopPath = "/usr/share/applications"
	} else {
		// User installation
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		desktopPath = filepath.Join(home, ".local", "share", "applications")
	}
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(desktopPath, 0755); err != nil {
		return fmt.Errorf("failed to create applications directory: %w", err)
	}
	
	// Write desktop file
	desktopFilePath := filepath.Join(desktopPath, l.config.AppName+".desktop")
	if err := os.WriteFile(desktopFilePath, []byte(desktopFile), 0644); err != nil {
		return fmt.Errorf("failed to create desktop file: %w", err)
	}
	
	// Update desktop database if available
	if _, err := exec.LookPath("update-desktop-database"); err == nil {
		exec.Command("update-desktop-database", desktopPath).Run()
	}
	
	return nil
}

func (l *LinuxPlatformInstaller) CreateShortcuts() error {
	// Create symbolic link in /usr/local/bin or ~/bin
	binPath := "/usr/local/bin"
	if !l.IsElevated() {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		binPath = filepath.Join(home, "bin")
		
		// Create ~/bin if it doesn't exist
		if err := os.MkdirAll(binPath, 0755); err != nil {
			return err
		}
	}
	
	// Create symlink
	source := filepath.Join(l.config.InstallDir, l.config.AppName)
	target := filepath.Join(binPath, l.config.AppName)
	
	// Remove existing symlink if it exists
	os.Remove(target)
	
	if err := os.Symlink(source, target); err != nil {
		// If symlink fails, try to copy the executable
		if err := copyFile(source, target); err != nil {
			return fmt.Errorf("failed to create shortcut: %w", err)
		}
		os.Chmod(target, 0755)
	}
	
	return nil
}

func (l *LinuxPlatformInstaller) RegisterUninstaller() error {
	// Create uninstall script
	uninstallScript := fmt.Sprintf(`#!/bin/bash
echo "Uninstalling %s..."

# Remove installation directory
rm -rf "%s"

# Remove desktop file
rm -f /usr/share/applications/%s.desktop
rm -f ~/.local/share/applications/%s.desktop

# Remove symlinks
rm -f /usr/local/bin/%s
rm -f ~/bin/%s

# Remove from PATH if added
sed -i '/%s/d' ~/.bashrc 2>/dev/null
sed -i '/%s/d' ~/.zshrc 2>/dev/null

echo "Uninstallation complete."
`, l.config.AppName, l.config.InstallDir, 
   l.config.AppName, l.config.AppName,
   l.config.AppName, l.config.AppName,
   strings.ReplaceAll(l.config.InstallDir, "/", "\\/"),
   strings.ReplaceAll(l.config.InstallDir, "/", "\\/"))
	
	uninstallPath := filepath.Join(l.config.InstallDir, "uninstall.sh")
	if err := os.WriteFile(uninstallPath, []byte(uninstallScript), 0755); err != nil {
		return fmt.Errorf("failed to create uninstaller: %w", err)
	}
	
	return nil
}

func (l *LinuxPlatformInstaller) UpdatePath(dirs []string, system bool) error {
	var rcFiles []string
	
	if system {
		// System-wide PATH update
		rcFiles = []string{"/etc/profile", "/etc/bash.bashrc"}
	} else {
		// User PATH update
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		
		// Check which shell config files exist
		possibleFiles := []string{".bashrc", ".zshrc", ".profile"}
		for _, file := range possibleFiles {
			path := filepath.Join(home, file)
			if _, err := os.Stat(path); err == nil {
				rcFiles = append(rcFiles, path)
			}
		}
	}
	
	// Build PATH export line
	var pathDirs []string
	for _, dir := range dirs {
		if filepath.IsAbs(dir) {
			pathDirs = append(pathDirs, dir)
		} else {
			pathDirs = append(pathDirs, filepath.Join(l.config.InstallDir, dir))
		}
	}
	
	pathLine := fmt.Sprintf("\n# Added by %s installer\nexport PATH=\"%s:$PATH\"\n",
		l.config.AppName, strings.Join(pathDirs, ":"))
	
	// Update each RC file
	for _, rcFile := range rcFiles {
		if err := l.updateRCFile(rcFile, pathLine); err != nil {
			// Log warning - logger might not be available here
			// l.config.Logger.Warn("Failed to update PATH in file", "file", rcFile, "error", err)
		}
	}
	
	return nil
}

func (l *LinuxPlatformInstaller) updateRCFile(filename, pathLine string) error {
	// Read existing file
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Create file if it doesn't exist
			return os.WriteFile(filename, []byte(pathLine), 0644)
		}
		return err
	}
	
	// Check if already contains our PATH
	if strings.Contains(string(content), l.config.AppName+" installer") {
		return nil // Already added
	}
	
	// Append PATH line
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = file.WriteString(pathLine)
	return err
}

// Helper function to copy file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}

// AddToPath adds a directory to the PATH environment variable
func (l *LinuxPlatformInstaller) AddToPath(dir string, system bool) error {
	if system {
		// Create profile.d script for system-wide PATH
		if !l.IsElevated() {
			return fmt.Errorf("root privileges required for system PATH")
		}
		
		scriptName := fmt.Sprintf("%s.sh", strings.ToLower(l.config.AppName))
		if scriptName == ".sh" {
			scriptName = "installer.sh"
		}
		
		scriptPath := filepath.Join("/etc/profile.d", scriptName)
		scriptContent := fmt.Sprintf("#!/bin/sh\nexport PATH=\"%s:$PATH\"\n", dir)
		
		return os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	} else {
		// Add to user's shell config
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		
		// Try common shell config files
		configFiles := []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".profile"),
		}
		
		pathLine := fmt.Sprintf("\n# Added by %s installer\nexport PATH=\"%s:$PATH\"\n", l.config.AppName, dir)
		
		for _, configFile := range configFiles {
			if _, err := os.Stat(configFile); err == nil {
				// File exists, append to it
				file, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					defer file.Close()
					_, err = file.WriteString(pathLine)
					return err
				}
			}
		}
		
		// If no config file found, create .profile
		profileFile := filepath.Join(home, ".profile")
		return os.WriteFile(profileFile, []byte(pathLine), 0644)
	}
}

// RemoveFromPath removes a directory from the PATH environment variable
func (l *LinuxPlatformInstaller) RemoveFromPath(dir string, system bool) error {
	if system {
		// Remove profile.d script
		if !l.IsElevated() {
			return fmt.Errorf("root privileges required for system PATH")
		}
		
		scriptName := fmt.Sprintf("%s.sh", strings.ToLower(l.config.AppName))
		if scriptName == ".sh" {
			scriptName = "installer.sh"
		}
		
		scriptPath := filepath.Join("/etc/profile.d", scriptName)
		return os.Remove(scriptPath)
	} else {
		// Remove from user's shell config
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		
		configFiles := []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".profile"),
		}
		
		for _, configFile := range configFiles {
			content, err := os.ReadFile(configFile)
			if err != nil {
				continue
			}
			
			pathLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", dir)
			if strings.Contains(string(content), pathLine) {
				// Remove the line
				lines := strings.Split(string(content), "\n")
				var newLines []string
				for _, line := range lines {
					if !strings.Contains(line, pathLine) && !strings.Contains(line, fmt.Sprintf("# Added by %s installer", l.config.AppName)) {
						newLines = append(newLines, line)
					}
				}
				return os.WriteFile(configFile, []byte(strings.Join(newLines, "\n")), 0644)
			}
		}
	}
	
	return nil
}

// IsInPath checks if a directory is in the PATH environment variable
func (l *LinuxPlatformInstaller) IsInPath(dir string, system bool) bool {
	if system {
		// Check profile.d
		scriptName := fmt.Sprintf("%s.sh", strings.ToLower(l.config.AppName))
		if scriptName == ".sh" {
			scriptName = "installer.sh"
		}
		
		scriptPath := filepath.Join("/etc/profile.d", scriptName)
		if content, err := os.ReadFile(scriptPath); err == nil {
			return strings.Contains(string(content), dir)
		}
	} else {
		// Check current PATH
		currentPath := os.Getenv("PATH")
		paths := strings.Split(currentPath, ":")
		for _, p := range paths {
			if p == dir {
				return true
			}
		}
		
		// Check shell config files
		home, _ := os.UserHomeDir()
		configFiles := []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".profile"),
		}
		
		pathLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", dir)
		for _, configFile := range configFiles {
			if content, err := os.ReadFile(configFile); err == nil {
				if strings.Contains(string(content), pathLine) {
					return true
				}
			}
		}
	}
	
	return false
}

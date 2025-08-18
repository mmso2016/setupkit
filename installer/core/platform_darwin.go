//go:build darwin
// +build darwin

package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DarwinPlatformInstaller implements PlatformInstaller for macOS
type DarwinPlatformInstaller struct {
	config *Config
}

// createDarwinPlatformInstaller is the internal factory function
func createDarwinPlatformInstaller(config *Config) PlatformInstaller {
	return &DarwinPlatformInstaller{
		config: config,
	}
}

// NewDarwinPlatformInstaller creates a macOS platform installer
func NewDarwinPlatformInstaller(config *Config) PlatformInstaller {
	return &DarwinPlatformInstaller{
		config: config,
	}
}

func (d *DarwinPlatformInstaller) Initialize() error {
	return nil
}

func (d *DarwinPlatformInstaller) CheckRequirements() error {
	// Check macOS version if needed
	// Could use sw_vers command to check version
	
	// Check for required tools
	requiredCommands := []string{"chmod", "ln", "defaults"}
	
	for _, cmd := range requiredCommands {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("required command '%s' not found", cmd)
		}
	}
	
	return nil
}

func (d *DarwinPlatformInstaller) IsElevated() bool {
	return os.Geteuid() == 0
}

func (d *DarwinPlatformInstaller) RequiresElevation() bool {
	// Check if installation path requires root
	systemPaths := []string{
		"/Applications",
		"/System",
		"/Library",
		"/usr/local",
	}
	
	for _, path := range systemPaths {
		if strings.HasPrefix(d.config.InstallDir, path) {
			return true
		}
	}
	
	// Check write permission
	if err := os.MkdirAll(d.config.InstallDir, 0755); err != nil {
		if os.IsPermission(err) {
			return true
		}
	}
	
	testFile := filepath.Join(d.config.InstallDir, ".write_test")
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

func (d *DarwinPlatformInstaller) RequestElevation() error {
	if d.IsElevated() {
		return nil
	}
	
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Use osascript to request admin privileges
	script := fmt.Sprintf(`do shell script "%s %s" with administrator privileges`,
		exe, strings.Join(os.Args[1:], " "))
	
	cmd := exec.Command("osascript", "-e", script)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("elevation failed: %w", err)
	}
	
	os.Exit(0)
	return nil
}

func (d *DarwinPlatformInstaller) RegisterWithOS() error {
	// Create .app bundle structure if appropriate
	if strings.HasSuffix(d.config.AppName, ".app") || strings.Contains(d.config.InstallDir, "/Applications") {
		return d.createAppBundle()
	}
	
	// Register with Launch Services
	if appPath := filepath.Join(d.config.InstallDir, d.config.AppName+".app"); fileExists(appPath) {
		cmd := exec.Command("/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister",
			"-f", appPath)
		cmd.Run() // Ignore errors
	}
	
	return nil
}

func (d *DarwinPlatformInstaller) CreateShortcuts() error {
	// Create symlink in /usr/local/bin
	binPath := "/usr/local/bin"
	
	// Create /usr/local/bin if it doesn't exist
	if err := os.MkdirAll(binPath, 0755); err != nil {
		// Try user's local bin
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		binPath = filepath.Join(home, ".local", "bin")
		os.MkdirAll(binPath, 0755)
	}
	
	// Create symlink
	source := filepath.Join(d.config.InstallDir, d.config.AppName)
	target := filepath.Join(binPath, d.config.AppName)
	
	// Remove existing symlink
	os.Remove(target)
	
	if err := os.Symlink(source, target); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}
	
	// Add to Dock if it's an app
	if strings.Contains(d.config.InstallDir, "/Applications") {
		d.addToDock()
	}
	
	return nil
}

func (d *DarwinPlatformInstaller) RegisterUninstaller() error {
	// Create uninstall script
	uninstallScript := fmt.Sprintf(`#!/bin/bash
echo "Uninstalling %s..."

# Remove app bundle or installation directory
rm -rf "%s"

# Remove symlinks
rm -f /usr/local/bin/%s
rm -f ~/.local/bin/%s

# Remove from Dock
defaults delete com.apple.dock persistent-apps | grep -v "%s" || true
killall Dock

# Remove from LaunchServices database
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -u "%s"

echo "Uninstallation complete."
`, d.config.AppName, d.config.InstallDir,
   d.config.AppName, d.config.AppName,
   d.config.AppName, 
   filepath.Join(d.config.InstallDir, d.config.AppName+".app"))
	
	uninstallPath := filepath.Join(d.config.InstallDir, "uninstall.sh")
	if err := os.WriteFile(uninstallPath, []byte(uninstallScript), 0755); err != nil {
		return fmt.Errorf("failed to create uninstaller: %w", err)
	}
	
	return nil
}

func (d *DarwinPlatformInstaller) UpdatePath(dirs []string, system bool) error {
	// Update shell configuration files
	var rcFiles []string
	
	if system {
		// System-wide PATH update
		rcFiles = []string{"/etc/paths.d/" + d.config.AppName}
		
		// Create paths.d file
		if len(rcFiles) > 0 && len(dirs) > 0 {
			pathContent := ""
			for _, dir := range dirs {
				if filepath.IsAbs(dir) {
					pathContent += dir + "\n"
				} else {
					pathContent += filepath.Join(d.config.InstallDir, dir) + "\n"
				}
			}
			return os.WriteFile(rcFiles[0], []byte(pathContent), 0644)
		}
	} else {
		// User PATH update
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		
		// Check which shell config files exist
		possibleFiles := []string{".zshrc", ".bash_profile", ".bashrc", ".profile"}
		for _, file := range possibleFiles {
			path := filepath.Join(home, file)
			if _, err := os.Stat(path); err == nil {
				rcFiles = append(rcFiles, path)
			}
		}
		
		// macOS uses zsh by default since Catalina
		if len(rcFiles) == 0 {
			rcFiles = []string{filepath.Join(home, ".zshrc")}
		}
	}
	
	// Build PATH export line
	var pathDirs []string
	for _, dir := range dirs {
		if filepath.IsAbs(dir) {
			pathDirs = append(pathDirs, dir)
		} else {
			pathDirs = append(pathDirs, filepath.Join(d.config.InstallDir, dir))
		}
	}
	
	pathLine := fmt.Sprintf("\n# Added by %s installer\nexport PATH=\"%s:$PATH\"\n",
		d.config.AppName, strings.Join(pathDirs, ":"))
	
	// Update each RC file
	for _, rcFile := range rcFiles {
		d.updateRCFile(rcFile, pathLine)
	}
	
	return nil
}

// Helper methods

func (d *DarwinPlatformInstaller) createAppBundle() error {
	appPath := filepath.Join(d.config.InstallDir, d.config.AppName+".app")
	
	// Create app bundle structure
	dirs := []string{
		filepath.Join(appPath, "Contents"),
		filepath.Join(appPath, "Contents", "MacOS"),
		filepath.Join(appPath, "Contents", "Resources"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	
	// Create Info.plist
	infoPlist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleExecutable</key>
	<string>%s</string>
	<key>CFBundleIdentifier</key>
	<string>com.%s.%s</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>%s</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>%s</string>
	<key>CFBundleVersion</key>
	<string>1</string>
	<key>LSMinimumSystemVersion</key>
	<string>10.12</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>`, d.config.AppName, 
		strings.ReplaceAll(strings.ToLower(d.config.Publisher), " ", ""),
		strings.ReplaceAll(strings.ToLower(d.config.AppName), " ", ""),
		d.config.AppName,
		d.config.Version)
	
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	if err := os.WriteFile(plistPath, []byte(infoPlist), 0644); err != nil {
		return err
	}
	
	// Copy executable to MacOS directory
	source := filepath.Join(d.config.InstallDir, d.config.AppName)
	target := filepath.Join(appPath, "Contents", "MacOS", d.config.AppName)
	
	// If executable exists, copy it
	if fileExists(source) {
		if err := copyFile(source, target); err != nil {
			return err
		}
		os.Chmod(target, 0755)
	}
	
	return nil
}

func (d *DarwinPlatformInstaller) addToDock() error {
	appPath := filepath.Join(d.config.InstallDir, d.config.AppName+".app")
	
	// Use defaults command to add to Dock
	script := fmt.Sprintf(`
		defaults write com.apple.dock persistent-apps -array-add '<dict><key>tile-data</key><dict><key>file-data</key><dict><key>_CFURLString</key><string>%s</string><key>_CFURLStringType</key><integer>0</integer></dict></dict></dict>'
		killall Dock
	`, appPath)
	
	cmd := exec.Command("bash", "-c", script)
	cmd.Run() // Ignore errors
	
	return nil
}

func (d *DarwinPlatformInstaller) updateRCFile(filename, pathLine string) error {
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
	if strings.Contains(string(content), d.config.AppName+" installer") {
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	err = os.WriteFile(dst, input, 0644)
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
func (d *DarwinPlatformInstaller) AddToPath(dir string, system bool) error {
	if system {
		// On macOS, system-wide PATH is managed differently
		// We'll use /etc/paths.d/ for system-wide PATH additions
		if !d.IsElevated() {
			return fmt.Errorf("root privileges required for system PATH")
		}
		
		pathsFile := fmt.Sprintf("/etc/paths.d/%s", strings.ToLower(d.config.AppName))
		if pathsFile == "/etc/paths.d/" {
			pathsFile = "/etc/paths.d/installer"
		}
		
		return os.WriteFile(pathsFile, []byte(dir+"\n"), 0644)
	} else {
		// Add to user's shell config
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		
		// Check for zsh (default on macOS Catalina+)
		configFile := filepath.Join(home, ".zshrc")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			// Fall back to bash
			configFile = filepath.Join(home, ".bash_profile")
		}
		
		pathLine := fmt.Sprintf("\n# Added by %s installer\nexport PATH=\"%s:$PATH\"\n", d.config.AppName, dir)
		
		// Check if already exists
		content, err := os.ReadFile(configFile)
		if err == nil && strings.Contains(string(content), pathLine) {
			return nil // Already added
		}
		
		file, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		
		_, err = file.WriteString(pathLine)
		return err
	}
}

// RemoveFromPath removes a directory from the PATH environment variable
func (d *DarwinPlatformInstaller) RemoveFromPath(dir string, system bool) error {
	if system {
		// Remove from /etc/paths.d/
		if !d.IsElevated() {
			return fmt.Errorf("root privileges required for system PATH")
		}
		
		pathsFile := fmt.Sprintf("/etc/paths.d/%s", strings.ToLower(d.config.AppName))
		if pathsFile == "/etc/paths.d/" {
			pathsFile = "/etc/paths.d/installer"
		}
		
		return os.Remove(pathsFile)
	} else {
		// Remove from user's shell config
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		
		configFiles := []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".bashrc"),
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
					if !strings.Contains(line, pathLine) && !strings.Contains(line, fmt.Sprintf("# Added by %s installer", d.config.AppName)) {
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
func (d *DarwinPlatformInstaller) IsInPath(dir string, system bool) bool {
	if system {
		// Check /etc/paths.d/
		pathsFile := fmt.Sprintf("/etc/paths.d/%s", strings.ToLower(d.config.AppName))
		if pathsFile == "/etc/paths.d/" {
			pathsFile = "/etc/paths.d/installer"
		}
		
		if content, err := os.ReadFile(pathsFile); err == nil {
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
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".bashrc"),
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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mmso2016/setupkit/installer/core"
)

func main() {
	// Command line flags
	var (
		install     = flag.Bool("install", false, "Install the application")
		uninstall   = flag.Bool("uninstall", false, "Uninstall the application")
		service     = flag.Bool("service", false, "Install as service")
		serviceName = flag.String("service-name", "goinstaller-example", "Service name")
		installDir  = flag.String("dir", getDefaultInstallDir(), "Installation directory")
		addPath     = flag.Bool("path", true, "Add to PATH")
		_           = flag.Bool("v", false, "Verbose output")
	)

	flag.Parse()

	// Create installer configuration
	config := &core.Config{
		AppName:    "GoSetupKitExample",
		Version:    "1.0.0",
		Publisher:  "Go SetupKit Team",
		InstallDir: *installDir,
		Mode:       core.ModeCLI,
	}

	// Get platform-specific installer
	platform, err := core.GetPlatformInstaller()
	if err != nil {
		log.Fatalf("Failed to get platform installer: %v", err)
	}

	if *install {
		if err := runInstallation(config, platform, *addPath, *service, *serviceName); err != nil {
			log.Fatalf("Installation failed: %v", err)
		}
		fmt.Println("Installation completed successfully!")
	} else if *uninstall {
		if err := runUninstallation(config, platform, *service, *serviceName); err != nil {
			log.Fatalf("Uninstallation failed: %v", err)
		}
		fmt.Println("Uninstallation completed successfully!")
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func getDefaultInstallDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("PROGRAMFILES"), "GoSetupKitExample")
	case "darwin":
		return "/Applications/GoSetupKitExample"
	default:
		return "/opt/setupkit-example"
	}
}

func runInstallation(config *core.Config, platform core.PlatformInstaller, addPath, installService bool, serviceName string) error {
	fmt.Println("Starting installation...")

	// Check if we need elevation
	if platform.RequiresElevation() && !platform.IsElevated() {
		// Try to detect if platform supports elevation
		if extended, ok := platform.(core.ExtendedPlatformInstaller); ok && extended.CanElevate() {
			fmt.Println("Requesting elevated privileges...")
			if err := platform.RequestElevation(); err != nil {
				return fmt.Errorf("failed to elevate: %w", err)
			}
		} else {
			fmt.Println("WARNING: This installation may require administrator privileges.")
			fmt.Println("If the installation fails, please run as administrator/root.")
		}
	}

	// Create installation directory
	if err := os.MkdirAll(config.InstallDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Copy files (simplified example)
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	destPath := filepath.Join(config.InstallDir, filepath.Base(exePath))
	if err := copyFile(exePath, destPath); err != nil {
		return fmt.Errorf("failed to copy executable: %w", err)
	}

	// Add to PATH if requested
	if addPath {
		fmt.Println("Adding to PATH...")
		if err := platform.AddToPath(config.InstallDir, true); err != nil {
			fmt.Printf("WARNING: Failed to add to PATH: %v\n", err)
		}
	}

	// Install service if requested
	if installService {
		fmt.Printf("Installing service '%s'...\n", serviceName)
		svcMgr, err := core.GetServiceManager()
		if err != nil {
			return fmt.Errorf("failed to get service manager: %w", err)
		}

		svcConfig := &core.ServiceConfig{
			Name:          serviceName,
			DisplayName:   "Go SetupKit Example Service",
			Description:   "Example service installed by Go SetupKit",
			Executable:    destPath,
			Arguments:     []string{"--run-service"},
			StartType:     core.ServiceStartAutomatic,
			RestartPolicy: core.RestartOnFailure,
		}

		if err := svcMgr.Install(svcConfig); err != nil {
			return fmt.Errorf("failed to install service: %w", err)
		}

		fmt.Println("Service installed successfully!")
	}

	// Register with OS
	fmt.Println("Registering with OS...")
	if err := platform.RegisterWithOS(); err != nil {
		fmt.Printf("WARNING: Failed to register with OS: %v\n", err)
	}

	// Create shortcuts
	fmt.Println("Creating shortcuts...")
	if err := platform.CreateShortcuts(); err != nil {
		fmt.Printf("WARNING: Failed to create shortcuts: %v\n", err)
	}

	// Register uninstaller
	fmt.Println("Registering uninstaller...")
	if err := platform.RegisterUninstaller(); err != nil {
		fmt.Printf("WARNING: Failed to register uninstaller: %v\n", err)
	}

	// Platform-specific post-installation
	if extended, ok := platform.(core.ExtendedPlatformInstaller); ok {
		// Set environment variables (example)
		if err := extended.SetEnv("GOSETUPKIT_HOME", config.InstallDir, true); err != nil {
			fmt.Printf("WARNING: Failed to set environment variable: %v\n", err)
		}

		// Write registry entries on Windows
		if runtime.GOOS == "windows" {
			regKey := `SOFTWARE\GoSetupKitExample`
			if err := extended.WriteRegistryString(regKey, "InstallPath", config.InstallDir); err != nil {
				fmt.Printf("WARNING: Failed to write registry: %v\n", err)
			}
			if err := extended.WriteRegistryString(regKey, "Version", config.Version); err != nil {
				fmt.Printf("WARNING: Failed to write registry: %v\n", err)
			}
		}
	}

	return nil
}

func runUninstallation(config *core.Config, platform core.PlatformInstaller, uninstallService bool, serviceName string) error {
	fmt.Println("Starting uninstallation...")

	// Check if we need elevation
	if platform.RequiresElevation() && !platform.IsElevated() {
		if extended, ok := platform.(core.ExtendedPlatformInstaller); ok && extended.CanElevate() {
			fmt.Println("Requesting elevated privileges...")
			if err := platform.RequestElevation(); err != nil {
				return fmt.Errorf("failed to elevate: %w", err)
			}
		} else {
			fmt.Println("WARNING: This uninstallation may require administrator privileges.")
		}
	}

	// Uninstall service if it was installed
	if uninstallService {
		fmt.Printf("Uninstalling service '%s'...\n", serviceName)
		svcMgr, err := core.GetServiceManager()
		if err != nil {
			fmt.Printf("WARNING: Failed to get service manager: %v\n", err)
		} else {
			// Stop service first
			if err := svcMgr.Stop(serviceName); err != nil {
				fmt.Printf("WARNING: Failed to stop service: %v\n", err)
			}

			// Uninstall service
			if err := svcMgr.Uninstall(serviceName); err != nil {
				fmt.Printf("WARNING: Failed to uninstall service: %v\n", err)
			} else {
				fmt.Println("Service uninstalled successfully!")
			}
		}
	}

	// Remove from PATH
	fmt.Println("Removing from PATH...")
	if err := platform.RemoveFromPath(config.InstallDir, true); err != nil {
		fmt.Printf("WARNING: Failed to remove from PATH: %v\n", err)
	}

	// Platform-specific cleanup
	if extended, ok := platform.(core.ExtendedPlatformInstaller); ok {
		// Remove environment variables
		if err := extended.UnsetEnv("GOSETUPKIT_HOME", true); err != nil {
			fmt.Printf("WARNING: Failed to unset environment variable: %v\n", err)
		}

		// Remove registry entries on Windows
		if runtime.GOOS == "windows" {
			regKey := `SOFTWARE\GoSetupKitExample`
			if err := extended.DeleteRegistryValue(regKey, "InstallPath"); err != nil {
				fmt.Printf("WARNING: Failed to delete registry value: %v\n", err)
			}
			if err := extended.DeleteRegistryValue(regKey, "Version"); err != nil {
				fmt.Printf("WARNING: Failed to delete registry value: %v\n", err)
			}
		}
	}

	// Remove installation directory
	fmt.Println("Removing installation directory...")
	if err := os.RemoveAll(config.InstallDir); err != nil {
		return fmt.Errorf("failed to remove installation directory: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	// This is a simplified file copy - in production use proper copy with permissions
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0755)
}

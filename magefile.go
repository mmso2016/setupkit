//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	// Build variables
	version   = getVersion()
	buildDate = time.Now().UTC().Format(time.RFC3339)

	// Directories
	binDir      = "bin"
	coverageDir = "coverage"

	// Binary extension
	binExt = getBinaryExt()
)

// Default target
var Default = All

// All runs clean, test and build
func All() {
	mg.Deps(Clean)
	mg.Deps(Test)
	mg.Deps(Build)
}

// Build builds the embedded installer demo
func Build() error {
	fmt.Println("Building SetupKit embedded installer...")
	fmt.Println("All configuration and assets are embedded in the executable.")
	fmt.Println()

	// Create bin directory
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	// Build ldflags for embedded version
	ldflags := fmt.Sprintf("-s -w -X main.Version=%s -X main.BuildDate=%s", version, buildDate)

	// Add Windows-specific ldflags for webview
	if runtime.GOOS == "windows" {
		ldflags += " -H windowsgui"
	}

	// Build embedded installer
	output := filepath.Join(binDir, "setupkit-installer-demo"+binExt)
	args := []string{"build", "-v", "-ldflags", ldflags, "-o", output, "./examples/installer-demo"}

	// Set environment variables for CGO (required for webview)
	env := map[string]string{
		"CGO_ENABLED": "1",
	}

	if err := sh.RunWithV(env, "go", args...); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✅ Embedded installer built: %s\n", output)
	fmt.Println("✅ Single-file installer ready - no external dependencies needed!")

	return nil
}

// BuildCustomStateDemo builds the custom state demo
func BuildCustomStateDemo() error {
	fmt.Println("Building SetupKit custom state demo...")
	fmt.Println("Demonstrates database configuration custom state.")
	fmt.Println()

	// Create bin directory
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	// Build custom state demo
	output := filepath.Join(binDir, "setupkit-custom-state-demo"+binExt)
	args := []string{"build", "-v", "-o", output, "./examples/custom-state-demo"}

	if err := sh.RunV("go", args...); err != nil {
		return fmt.Errorf("custom state demo build failed: %w", err)
	}

	fmt.Printf("✅ Custom state demo built: %s\n", output)
	fmt.Println("✅ Database configuration demo ready!")

	return nil
}

// BuildAll builds both demo applications
func BuildAll() error {
	mg.Deps(Build)
	mg.Deps(BuildCustomStateDemo)
	return nil
}

// Test runs all tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "-short", "./pkg/...")
}

// TestVerbose runs tests with verbose output
func TestVerbose() error {
	fmt.Println("Running tests (verbose)...")
	return sh.RunV("go", "test", "-v", "./pkg/...")
}

// TestRace runs tests with race detector
func TestRace() error {
	fmt.Println("Running tests with race detector...")
	return sh.RunV("go", "test", "-race", "-short", "./pkg/...")
}

// Coverage generates test coverage report
func Coverage() error {
	fmt.Println("Generating coverage report...")

	// Create coverage directory
	if err := os.MkdirAll(coverageDir, 0755); err != nil {
		return err
	}

	// Generate coverage data
	coverFile := filepath.Join(coverageDir, "coverage.out")
	if err := sh.Run("go", "test", "-coverprofile="+coverFile, "./pkg/..."); err != nil {
		return err
	}

	// Generate HTML report
	htmlFile := filepath.Join(coverageDir, "coverage.html")
	if err := sh.Run("go", "tool", "cover", "-html="+coverFile, "-o", htmlFile); err != nil {
		return err
	}

	fmt.Printf("Coverage report generated: %s\n", htmlFile)
	return nil
}

// Fmt formats the code
func Fmt() error {
	fmt.Println("Formatting code...")
	return sh.RunV("go", "fmt", "./...")
}

// Vet runs go vet
func Vet() error {
	fmt.Println("Running go vet...")
	return sh.RunV("go", "vet", "./...")
}

// Lint runs golangci-lint
func Lint() error {
	fmt.Println("Running linters...")
	return sh.RunV("golangci-lint", "run")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")

	// Remove build directories
	dirs := []string{binDir, coverageDir}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("Warning: Failed to remove %s: %v\n", dir, err)
		}
	}

	// Remove test files
	if err := removeFiles(".", "*.test"); err != nil {
		fmt.Printf("Warning: Failed to remove test files: %v\n", err)
	}

	if err := removeFiles(".", "*.out"); err != nil {
		fmt.Printf("Warning: Failed to remove .out files: %v\n", err)
	}

	fmt.Println("✅ Cleanup complete!")
	return nil
}


// Deps downloads dependencies
func Deps() error {
	fmt.Println("Downloading dependencies...")
	return sh.RunV("go", "mod", "download")
}

// Tidy runs go mod tidy
func Tidy() error {
	fmt.Println("Tidying go.mod...")
	return sh.RunV("go", "mod", "tidy")
}

// Verify verifies dependencies
func Verify() error {
	fmt.Println("Verifying dependencies...")
	return sh.RunV("go", "mod", "verify")
}

// Run runs the embedded installer in auto mode
func Run() error {
	mg.Deps(Build)
	binary := filepath.Join(binDir, "setupkit-installer-demo"+binExt)
	fmt.Println("Running embedded installer (auto mode)...")
	fmt.Println("Uses embedded configuration and assets")
	fmt.Println()
	return sh.RunV(binary)
}

// RunGUI runs the installer in GUI mode
func RunGUI() error {
	mg.Deps(Build)
	binary := filepath.Join(binDir, "setupkit-installer-demo"+binExt)
	fmt.Println("Starting embedded installer (GUI mode)...")
	fmt.Println("Opens browser-based interface with embedded assets")
	fmt.Println()
	return sh.RunV(binary, "-mode=gui")
}

// RunCLI runs the installer in CLI mode
func RunCLI() error {
	mg.Deps(Build)
	binary := filepath.Join(binDir, "setupkit-installer-demo"+binExt)
	fmt.Println("Starting embedded installer (CLI mode)...")
	fmt.Println("Interactive CLI with embedded configuration")
	fmt.Println()
	return sh.RunV(binary, "-mode=cli")
}

// RunSilent runs the installer in silent mode
func RunSilent() error {
	mg.Deps(Build)
	binary := filepath.Join(binDir, "setupkit-installer-demo"+binExt)
	fmt.Println("Starting embedded installer (silent mode)...")
	fmt.Println("Unattended installation with embedded assets")
	fmt.Println()
	return sh.RunV(binary, "-silent", "-profile=minimal")
}

// RunCustomStateDemo runs the custom state demo in silent mode
func RunCustomStateDemo() error {
	mg.Deps(BuildCustomStateDemo)
	binary := filepath.Join(binDir, "setupkit-custom-state-demo"+binExt)
	fmt.Println("Starting custom state demo (database configuration)...")
	fmt.Println("Demonstrates: Welcome → License → Components → Install Path → DB Config → Summary → Complete")
	fmt.Println()
	return sh.RunV(binary, "-mode=silent")
}

// RunCustomStateDemoCLI runs the custom state demo in CLI mode
func RunCustomStateDemoCLI() error {
	mg.Deps(BuildCustomStateDemo)
	binary := filepath.Join(binDir, "setupkit-custom-state-demo"+binExt)
	fmt.Println("Starting custom state demo (CLI mode)...")
	fmt.Println("Interactive CLI with database configuration state")
	fmt.Println()
	return sh.RunV(binary, "-mode=cli")
}

// RunCustomStateDemoAuto runs the custom state demo in auto mode
func RunCustomStateDemoAuto() error {
	mg.Deps(BuildCustomStateDemo)
	binary := filepath.Join(binDir, "setupkit-custom-state-demo"+binExt)
	fmt.Println("Starting custom state demo (auto mode)...")
	fmt.Println("Auto-selects best UI mode for database configuration")
	fmt.Println()
	return sh.RunV(binary, "-mode=auto")
}

// HelpCustomStateDemo shows custom state demo help
func HelpCustomStateDemo() error {
	mg.Deps(BuildCustomStateDemo)
	binary := filepath.Join(binDir, "setupkit-custom-state-demo"+binExt)
	fmt.Println("Custom State Demo Help:")
	fmt.Println()
	return sh.RunV(binary, "--help")
}

// CleanInstall cleans test installation directories
func CleanInstall() error {
	fmt.Println("Cleaning test installation directories...")

	testDirs := []string{
		"/tmp/DemoApp",
		"C:\\Program Files\\DemoApp",
		filepath.Join(os.Getenv("HOME"), "Applications", "DemoApp"),
		"/opt/demoapp",
	}

	for _, dir := range testDirs {
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: Failed to remove %s: %v\n", dir, err)
		}
	}

	fmt.Println("✅ Test installations cleaned")
	return nil
}

// Version shows version information
func Version() {
	fmt.Printf("SetupKit Framework\n")
	fmt.Printf("==================\n")
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Build Date: %s\n", buildDate)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// Helper functions

func getBinaryExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func getVersion() string {
	// Try to get version from git
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(output))
}

func removeFiles(dir, pattern string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			os.Remove(path)
		}

		return nil
	})
}


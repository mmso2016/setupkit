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
	exampleDir  = "./examples/minimal"
	coverageDir = "coverage"

	// Binary name and extension
	binaryName = "setupkit-example"
	binExt     = getBinaryExt()
)

// Default target
var Default = All

// All runs clean, test and build
func All() {
	mg.Deps(Clean)
	mg.Deps(Test)
	mg.Deps(Build)
}

// Build builds the example installer using the framework
func Build() error {
	fmt.Println("Building SetupKit example installer...")
	fmt.Println("Note: The SetupKit framework provides all UI components!")
	fmt.Println()

	// Create bin directory
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	output := filepath.Join(binDir, binaryName+binExt)

	// Build ldflags
	ldflags := fmt.Sprintf("-s -w -X main.Version=%s -X main.BuildDate=%s", version, buildDate)

	// Build the example - the framework handles everything
	fmt.Println("Building with embedded framework UI...")
	args := []string{"build", "-tags", "desktop,production", "-v", "-ldflags", ldflags, "-o", output, exampleDir}

	if err := sh.RunV("go", args...); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Build successful!")
	fmt.Printf("📦 Output: %s\n", output)
	fmt.Println()
	fmt.Println("The framework has embedded the complete UI.")
	fmt.Println("No additional files or folders needed!")

	return nil
}

// Test runs all tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "-short", "./pkg/...", "./internal/...")
}

// TestVerbose runs tests with verbose output
func TestVerbose() error {
	fmt.Println("Running tests (verbose)...")
	return sh.RunV("go", "test", "-v", ".")
}

// TestRace runs tests with race detector
func TestRace() error {
	fmt.Println("Running tests with race detector...")
	return sh.RunV("go", "test", "-race", "-short", "./pkg/...", "./internal/...")
}

// Bench runs benchmarks
func Bench() error {
	fmt.Println("Running benchmarks...")
	return sh.RunV("go", "test", "-bench=.", "-benchmem", "./pkg/...", "./internal/...")
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
	if err := sh.Run("go", "test", "-coverprofile="+coverFile, "./pkg/...", "./internal/..."); err != nil {
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

// Clean removes ALL build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")

	// Remove root build directories
	dirs := []string{binDir, coverageDir}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("Warning: Failed to remove %s: %v\n", dir, err)
		}
	}

	// Clean up example directory - remove any generated folders
	fmt.Println("Cleaning example directory...")

	// Remove unwanted directories from example
	unwantedDirs := []string{
		filepath.Join(exampleDir, "build"),
		filepath.Join(exampleDir, "frontend"),
		filepath.Join(exampleDir, "bin"),
	}

	for _, dir := range unwantedDirs {
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("  Removing %s...\n", dir)
			if err := os.RemoveAll(dir); err != nil {
				fmt.Printf("  Warning: Failed to remove %s: %v\n", dir, err)
			}
		}
	}

	// Remove old/backup files
	oldFiles := []string{
		filepath.Join(exampleDir, "wails.json.old"),
		filepath.Join(exampleDir, "main.go.old"),
		filepath.Join(exampleDir, "main_simple.go.old"),
		filepath.Join(exampleDir, "app.go"),
	}

	for _, file := range oldFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("  Removing %s...\n", file)
			os.Remove(file)
		}
	}

	// Remove test files
	if err := removeFiles(".", "*.test"); err != nil {
		fmt.Printf("Warning: Failed to remove test files: %v\n", err)
	}

	if err := removeFiles(".", "*.out"); err != nil {
		fmt.Printf("Warning: Failed to remove .out files: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✅ Cleanup complete!")
	fmt.Println("The example directory now contains only the essential files.")

	return nil
}

// CleanExample removes all generated files from the example directory
func CleanExample() error {
	fmt.Println("Cleaning example directory to framework-only state...")

	// Remove ALL generated directories
	unwantedDirs := []string{
		filepath.Join(exampleDir, "build"),
		filepath.Join(exampleDir, "frontend"),
		filepath.Join(exampleDir, "bin"),
	}

	for _, dir := range unwantedDirs {
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("Removing %s\n", dir)
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("failed to remove %s: %w", dir, err)
			}
		}
	}

	// Remove wails.json if it exists
	wailsFile := filepath.Join(exampleDir, "wails.json")
	if _, err := os.Stat(wailsFile); err == nil {
		fmt.Printf("Removing %s\n", wailsFile)
		os.Remove(wailsFile)
	}

	fmt.Println()
	fmt.Println("✅ Example directory cleaned!")
	fmt.Println("Only main.go and README.md remain - as it should be in a framework!")

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

// Run runs the example installer
func Run() error {
	mg.Deps(Build)
	binary := filepath.Join(binDir, binaryName+binExt)
	fmt.Println("Running example installer...")
	fmt.Println()
	return sh.RunV(binary)
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

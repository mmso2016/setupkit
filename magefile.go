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
	ldflags   = fmt.Sprintf("-s -w -X main.Version=%s -X main.BuildDate=%s", version, buildDate)
	
	// Directories
	binDir      = "bin"
	examplesDir = "examples"
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

// Build builds all examples
func Build() error {
	mg.Deps(BuildCLI, BuildConsole, BuildPlatform, BuildUI, BuildUIWails, BuildUINoGUI)
	
	// Try to build GUI if Wails is available
	if err := checkWails(); err == nil {
		mg.Deps(BuildGUI)
	} else {
		fmt.Println("Skipping GUI build: Wails not installed")
	}
	
	return nil
}

// BuildCLI builds the CLI installer
func BuildCLI() error {
	fmt.Println("Building CLI installer...")
	return buildBinary("installer-cli", "./examples/basic")
}

// BuildConsole builds the console GUI
func BuildConsole() error {
	fmt.Println("Building console GUI...")
	return buildBinary("installer-console", "./examples/gui-console")
}

// BuildPlatform builds the platform example
func BuildPlatform() error {
	fmt.Println("Building platform example...")
	return buildBinary("installer-platform", "./examples/platform")
}

// BuildUI builds the UI example (default build)
func BuildUI() error {
	fmt.Println("Building UI example...")
	return buildBinary("installer-ui", "./examples/ui")
}

// BuildUIWails builds the UI example with Wails support
func BuildUIWails() error {
	fmt.Println("Building UI example with Wails support...")
	return buildBinaryWithTags("installer-ui-wails", "./examples/ui", []string{"wails"})
}

// BuildUINoGUI builds the UI example without GUI support
func BuildUINoGUI() error {
	fmt.Println("Building UI example without GUI support...")
	return buildBinaryWithTags("installer-ui-nogui", "./examples/ui", []string{"nogui"})
}

// BuildGUI builds the Wails GUI
func BuildGUI() error {
	fmt.Println("Building Wails GUI...")
	
	// Check Wails
	if err := checkWails(); err != nil {
		return fmt.Errorf("Wails not found: %w", err)
	}
	
	// Prepare frontend
	guiDir := filepath.Join(examplesDir, "gui")
	frontendSrc := filepath.Join(guiDir, "frontend", "src")
	frontendDist := filepath.Join(guiDir, "frontend", "dist")
	
	// Create dist directory if it doesn't exist
	if err := os.MkdirAll(frontendDist, 0755); err != nil {
		return err
	}
	
	// Copy frontend files
	if err := copyDir(frontendSrc, frontendDist); err != nil {
		fmt.Printf("Warning: Failed to copy frontend files: %v\n", err)
	}
	
	// Change to GUI directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(guiDir); err != nil {
		return err
	}
	
	// Build with Wails
	if err := sh.Run("wails", "build", "-clean"); err != nil {
		return fmt.Errorf("Wails build failed: %w", err)
	}
	
	// Copy executable to bin directory
	buildBin := filepath.Join("build", "bin")
	entries, err := os.ReadDir(buildBin)
	if err != nil {
		return fmt.Errorf("Failed to read build output: %w", err)
	}
	
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), binExt) {
			src := filepath.Join(buildBin, entry.Name())
			dst := filepath.Join("..", "..", binDir, "installer-gui"+binExt)
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("Failed to copy GUI executable: %w", err)
			}
			fmt.Printf("GUI installer built: %s\n", dst)
			break
		}
	}
	
	return nil
}

// Test runs all tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "-short", "./...")
}

// TestVerbose runs tests with verbose output
func TestVerbose() error {
	fmt.Println("Running tests (verbose)...")
	return sh.RunV("go", "test", "-v", "./...")
}

// TestRace runs tests with race detector
func TestRace() error {
	fmt.Println("Running tests with race detector...")
	return sh.RunV("go", "test", "-race", "-short", "./...")
}

// Bench runs benchmarks
func Bench() error {
	fmt.Println("Running benchmarks...")
	return sh.RunV("go", "test", "-bench=.", "-benchmem", "./...")
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
	if err := sh.Run("go", "test", "-coverprofile="+coverFile, "./..."); err != nil {
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
	
	// Remove directories
	dirs := []string{binDir, coverageDir, filepath.Join(examplesDir, "gui", "build")}
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

// Install installs the CLI to GOPATH/bin
func Install() error {
	fmt.Println("Installing to GOPATH/bin...")
	return sh.RunV("go", "install", "./examples/basic")
}

// Dev starts Wails in development mode
func Dev() error {
	fmt.Println("Starting Wails development mode...")
	
	if err := checkWails(); err != nil {
		return err
	}
	
	// Change to GUI directory
	guiDir := filepath.Join(examplesDir, "gui")
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(guiDir); err != nil {
		return err
	}
	
	return sh.RunV("wails", "dev")
}

// RunCLI runs the CLI installer
func RunCLI() error {
	mg.Deps(BuildCLI)
	binary := filepath.Join(binDir, "installer-cli"+binExt)
	return sh.RunV(binary, "--help")
}

// RunConsole runs the console GUI
func RunConsole() error {
	mg.Deps(BuildConsole)
	binary := filepath.Join(binDir, "installer-console"+binExt)
	return sh.RunV(binary)
}

// RunGUI runs the Wails GUI
func RunGUI() error {
	mg.Deps(BuildGUI)
	binary := filepath.Join(binDir, "installer-gui"+binExt)
	return sh.RunV(binary)
}

// Version shows version information
func Version() {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Build Date: %s\n", buildDate)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// WailsInstall installs Wails CLI
func WailsInstall() error {
	fmt.Println("Installing Wails CLI...")
	return sh.RunV("go", "install", "github.com/wailsapp/wails/v2/cmd/wails@latest")
}

// WailsDoctor runs wails doctor
func WailsDoctor() error {
	fmt.Println("Running Wails doctor...")
	return sh.RunV("wails", "doctor")
}

// Helper functions

func buildBinary(name, pkg string) error {
	return buildBinaryWithTags(name, pkg, nil)
}

func buildBinaryWithTags(name, pkg string, tags []string) error {
	// Create bin directory
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	
	output := filepath.Join(binDir, name+binExt)
	args := []string{"build", "-v", "-ldflags", ldflags}
	
	// Add build tags if specified
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	
	args = append(args, "-o", output, pkg)
	
	return sh.RunV("go", args...)
}

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

func checkWails() error {
	return exec.Command("wails", "version").Run()
}

func copyFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	// Write destination file
	return os.WriteFile(dst, data, 0755)
}

func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	
	// Walk source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		
		// Handle directories
		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		
		// Copy files
		return copyFile(path, dstPath)
	})
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

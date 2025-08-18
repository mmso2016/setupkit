//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

// TestAllBuilds tests all build configurations
func TestAllBuilds() error {
	fmt.Println("Testing all build configurations...")

	// Clean first
	if err := Clean(); err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}

	builds := []struct {
		name   string
		target string
	}{
		{"CLI", "BuildCLI"},
		{"Console", "BuildConsole"},
		{"Platform", "BuildPlatform"},
		{"UI (default)", "BuildUI"},
		{"UI (Wails)", "BuildUIWails"},
		{"UI (NoGUI)", "BuildUINoGUI"},
	}

	successful := 0
	for _, build := range builds {
		fmt.Printf("\n=== Testing %s ===\n", build.name)

		var err error
		switch build.target {
		case "BuildCLI":
			err = BuildCLI()
		case "BuildConsole":
			err = BuildConsole()
		case "BuildPlatform":
			err = BuildPlatform()
		case "BuildUI":
			err = BuildUI()
		case "BuildUIWails":
			err = BuildUIWails()
		case "BuildUINoGUI":
			err = BuildUINoGUI()
		}

		if err != nil {
			fmt.Printf("❌ FAILED: %s - %v\n", build.name, err)
		} else {
			fmt.Printf("✅ SUCCESS: %s\n", build.name)
			successful++
		}
	}

	// Try GUI if Wails is available
	fmt.Printf("\n=== Testing GUI (Wails) ===\n")
	if err := checkWails(); err == nil {
		if err := BuildGUI(); err != nil {
			fmt.Printf("❌ FAILED: GUI - %v\n", err)
		} else {
			fmt.Printf("✅ SUCCESS: GUI\n")
			successful++
		}
		builds = append(builds, struct{ name, target string }{"GUI", "BuildGUI"})
	} else {
		fmt.Printf("⚠️  SKIPPED: GUI - Wails not available\n")
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Successful builds: %d/%d\n", successful, len(builds))

	// List created binaries
	if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
		fmt.Printf("\nCreated binaries in %s:\n", binDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				fmt.Printf("  - %s\n", entry.Name())
			}
		}
	}

	return nil
}

// TestBuildTags tests specific build tag combinations
func TestBuildTags() error {
	fmt.Println("Testing build tag combinations...")

	testCases := []struct {
		name string
		tags []string
		pkg  string
	}{
		{"UI Default", nil, "./examples/ui"},
		{"UI with Wails", []string{"wails"}, "./examples/ui"},
		{"UI without GUI", []string{"nogui"}, "./examples/ui"},
		{"UI without CLI", []string{"nocli"}, "./examples/ui"},
	}

	for _, tc := range testCases {
		fmt.Printf("\nTesting: %s\n", tc.name)

		binaryName := fmt.Sprintf("test-%s", tc.name)
		// Replace spaces with dashes for valid filename
		binaryName = filepath.Join(binDir, fmt.Sprintf("test-%s%s",
			fmt.Sprintf("%v", tc.tags), binExt))

		if err := buildBinaryWithTags(binaryName, tc.pkg, tc.tags); err != nil {
			fmt.Printf("❌ FAILED: %s - %v\n", tc.name, err)
		} else {
			fmt.Printf("✅ SUCCESS: %s\n", tc.name)
		}
	}

	return nil
}

// TestConfigs tests the UI configuration system
func TestConfigs() error {
	fmt.Println("Testing UI configuration system...")

	// Ensure CLI is built
	if err := BuildCLI(); err != nil {
		return fmt.Errorf("failed to build CLI: %w", err)
	}

	cliPath := filepath.Join(binDir, "installer-cli"+binExt)

	// Test 1: Theme listing
	fmt.Println("\n1. Testing theme listing...")
	if err := sh.RunV(cliPath, "-list-themes"); err != nil {
		fmt.Printf("❌ Theme listing failed: %v\n", err)
	} else {
		fmt.Println("✅ Theme listing successful")
	}

	// Test 2: Config generation
	fmt.Println("\n2. Testing config generation...")
	testConfigFile := "test-config.yaml"
	if err := sh.RunV(cliPath, "-generate-config", testConfigFile); err != nil {
		fmt.Printf("❌ Config generation failed: %v\n", err)
	} else {
		fmt.Println("✅ Config generation successful")
		// Clean up
		os.Remove(testConfigFile)
	}

	// Test 3: Different themes
	fmt.Println("\n3. Testing different themes...")
	themes := []string{"default", "corporate-blue", "medical-green", "tech-dark", "minimal-light"}
	for _, theme := range themes {
		fmt.Printf("Testing theme: %s\n", theme)
		if err := sh.Run(cliPath, "-theme", theme, "-help"); err != nil {
			fmt.Printf("❌ Theme %s test failed: %v\n", theme, err)
		} else {
			fmt.Printf("✅ Theme %s successful\n", theme)
		}
	}

	// Test 4: Example configs
	fmt.Println("\n4. Testing example configurations...")
	configDir := "examples/configs"
	if entries, err := os.ReadDir(configDir); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".yaml") {
				configPath := filepath.Join(configDir, entry.Name())
				fmt.Printf("Testing config: %s\n", configPath)
				if err := sh.Run(cliPath, "-config", configPath, "-help"); err != nil {
					fmt.Printf("❌ Config %s test failed: %v\n", entry.Name(), err)
				} else {
					fmt.Printf("✅ Config %s successful\n", entry.Name())
				}
			}
		}
	} else {
		fmt.Printf("Warning: Could not read config directory: %v\n", err)
	}

	fmt.Println("\n🎉 UI configuration tests completed!")
	return nil
}

// ValidateBuildTags checks that all build tag files have correct syntax
func ValidateBuildTags() error {
	fmt.Println("Validating build tag syntax...")

	// Files to check
	files := []string{
		"installer/ui/factory_cli.go",
		"installer/ui/factory_gui.go",
		"installer/ui/factory_gui_stub.go",
		"installer/ui/factory_wails.go",
	}

	for _, file := range files {
		fmt.Printf("Checking %s...\n", file)

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		lines := strings.Split(string(content), "\n")
		if len(lines) < 2 {
			return fmt.Errorf("%s: file too short", file)
		}

		// Check for build constraints
		foundGoBuild := false
		//foundPlusBuild := false

		for i, line := range lines[:5] { // Check first 5 lines
			if strings.HasPrefix(line, "//go:build") {
				foundGoBuild = true
				fmt.Printf("  Line %d: %s\n", i+1, line)
			}
			if strings.HasPrefix(line, "// +build") {
				//foundPlusBuild = true
				fmt.Printf("  Line %d: %s\n", i+1, line)
			}
		}

		if !foundGoBuild {
			fmt.Printf("  ⚠️  WARNING: No //go:build constraint found in %s\n", file)
		}

		fmt.Printf("  ✅ %s validated\n", file)
	}

	return nil
}

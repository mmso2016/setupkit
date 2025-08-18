package installer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mmso2016/setupkit/installer"
)

// TestNewInstaller tests installer creation with various options
func TestNewInstaller(t *testing.T) {
	tests := []struct {
		name    string
		opts    []installer.Option
		wantErr bool
	}{
		{
			name: "minimal config",
			opts: []installer.Option{
				installer.WithAppName("TestApp"),
				installer.WithVersion("1.0.0"),
			},
			wantErr: false,
		},
		{
			name: "full config",
			opts: []installer.Option{
				installer.WithAppName("TestApp"),
				installer.WithVersion("1.0.0"),
				installer.WithPublisher("Test Publisher"),
				installer.WithWebsite("https://example.com"),
				installer.WithInstallDir("/tmp/testapp"),
				installer.WithMode(installer.ModeSilent),
			},
			wantErr: false,
		},
		{
			name:    "empty config",
			opts:    []installer.Option{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, err := installer.New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && inst == nil {
				t.Error("New() returned nil installer without error")
			}
		})
	}
}

// TestInstallerConfig tests configuration management
func TestInstallerConfig(t *testing.T) {
	appName := "TestApp"
	version := "1.2.3"
	installDir := "/opt/testapp"

	inst, err := installer.New(
		installer.WithAppName(appName),
		installer.WithVersion(version),
		installer.WithInstallDir(installDir),
	)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	config := inst.GetConfig()
	if config.AppName != appName {
		t.Errorf("AppName = %v, want %v", config.AppName, appName)
	}
	if config.Version != version {
		t.Errorf("Version = %v, want %v", config.Version, version)
	}
	if config.InstallDir != installDir {
		t.Errorf("InstallDir = %v, want %v", config.InstallDir, installDir)
	}
}

// TestComponentSelection tests component selection functionality
func TestComponentSelection(t *testing.T) {
	components := []installer.Component{
		{
			ID:       "core",
			Name:     "Core Files",
			Required: true,
			Selected: true,
		},
		{
			ID:       "docs",
			Name:     "Documentation",
			Required: false,
			Selected: false,
		},
		{
			ID:       "examples",
			Name:     "Examples",
			Required: false,
			Selected: true,
		},
	}

	inst, err := installer.New(
		installer.WithAppName("TestApp"),
		installer.WithComponents(components...),
	)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	gotComponents := inst.GetComponents()
	if len(gotComponents) != len(components) {
		t.Errorf("GetComponents() returned %d components, want %d",
			len(gotComponents), len(components))
	}

	// Test SetSelectedComponents
	selected := []installer.Component{components[0], components[2]}
	inst.SetSelectedComponents(selected)

	// Verify selection
	gotComponents = inst.GetComponents()
	for _, comp := range gotComponents {
		switch comp.ID {
		case "core":
			if !comp.Selected {
				t.Error("Core component should be selected")
			}
		case "docs":
			if comp.Selected {
				t.Error("Docs component should not be selected")
			}
		case "examples":
			if !comp.Selected {
				t.Error("Examples component should be selected")
			}
		}
	}
}

// TestExitCodes tests exit code functionality
func TestExitCodes(t *testing.T) {
	tests := []struct {
		code int
		desc string
	}{
		{installer.ExitSuccess, "Installation completed successfully"},
		{installer.ExitGeneralError, "General error occurred"},
		{installer.ExitPermissionError, "Permission denied"},
		{installer.ExitUserCancelled, "Installation cancelled by user"},
		{999, "Unknown exit code: 999"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := installer.ExitCodeDescription(tt.code)
			if got != tt.desc {
				t.Errorf("ExitCodeDescription(%d) = %v, want %v",
					tt.code, got, tt.desc)
			}
		})
	}
}

// TestInstallError tests custom error type
func TestInstallError(t *testing.T) {
	cause := os.ErrPermission
	err := installer.NewError(
		installer.ExitPermissionError,
		"Cannot write to directory",
		cause,
	)

	if err.ExitCode() != installer.ExitPermissionError {
		t.Errorf("ExitCode() = %v, want %v",
			err.ExitCode(), installer.ExitPermissionError)
	}

	expectedMsg := "Cannot write to directory: " + cause.Error()
	if err.Error() != expectedMsg {
		t.Errorf("Error() = %v, want %v", err.Error(), expectedMsg)
	}

	// Test GetExitCodeForError
	code := installer.GetExitCodeForError(err)
	if code != installer.ExitPermissionError {
		t.Errorf("GetExitCodeForError() = %v, want %v",
			code, installer.ExitPermissionError)
	}

	// Test with nil error
	code = installer.GetExitCodeForError(nil)
	if code != installer.ExitSuccess {
		t.Errorf("GetExitCodeForError(nil) = %v, want %v",
			code, installer.ExitSuccess)
	}

	// Test with regular error
	code = installer.GetExitCodeForError(os.ErrNotExist)
	if code != installer.ExitGeneralError {
		t.Errorf("GetExitCodeForError(regular) = %v, want %v",
			code, installer.ExitGeneralError)
	}
}

// TestLogger tests logger functionality
func TestLogger(t *testing.T) {
	// Create temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	logger := installer.NewLogger("info", logFile)

	// Ensure logger is closed after test
	defer func() {
		// Type assertion to check if logger has Close method
		if closer, ok := logger.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}()

	// Test logging methods (shouldn't panic)
	logger.Debug("Debug message", "key", "value")
	logger.Info("Info message", "count", 42)
	logger.Warn("Warning message")
	logger.Error("Error message", "error", "test error")
	logger.Verbose("Verbose message")

	// Test verbose mode
	logger.SetVerbose(true)
	logger.VerboseSection("Test Section")

	// Verify log file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestPathConfiguration tests PATH configuration
func TestPathConfiguration(t *testing.T) {
	pathConfig := &installer.PathConfiguration{
		Enabled: true,
		System:  false,
		Dirs:    []string{"bin", "scripts"},
	}

	inst, err := installer.New(
		installer.WithAppName("TestApp"),
		installer.WithPathConfig(pathConfig),
	)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	config := inst.GetConfig()
	if config.PathConfig == nil {
		t.Fatal("PathConfig is nil")
	}

	if !config.PathConfig.Enabled {
		t.Error("PathConfig should be enabled")
	}

	if config.PathConfig.System {
		t.Error("PathConfig should use user PATH")
	}

	if len(config.PathConfig.Dirs) != 2 {
		t.Errorf("PathConfig.Dirs length = %d, want 2",
			len(config.PathConfig.Dirs))
	}
}

// TestComponentValidation tests component validation
func TestComponentValidation(t *testing.T) {
	validatorCalled := false
	component := installer.Component{
		ID:          "test",
		Name:        "Test Component",
		Description: "Test component for validation",
		Validator: func() error {
			validatorCalled = true
			return nil
		},
		Installer: func(ctx context.Context) error {
			return nil
		},
	}

	// Verify component fields
	if component.ID != "test" {
		t.Errorf("ID = %v, want test", component.ID)
	}

	if component.Name != "Test Component" {
		t.Errorf("Name = %v, want Test Component", component.Name)
	}

	if component.Description != "Test component for validation" {
		t.Errorf("Description = %v, want Test component for validation", component.Description)
	}

	// Test Installer is set
	if component.Installer == nil {
		t.Error("Installer should not be nil")
	}

	// Call validator
	err := component.Validator()
	if err != nil {
		t.Errorf("Validator returned error: %v", err)
	}

	if !validatorCalled {
		t.Error("Validator was not called")
	}
}

// TestModeConstants tests installation mode constants
func TestModeConstants(t *testing.T) {
	modes := []installer.Mode{
		installer.ModeAuto,
		installer.ModeGUI,
		installer.ModeCLI,
		installer.ModeSilent,
	}

	// Ensure modes are distinct
	seen := make(map[installer.Mode]bool)
	for _, mode := range modes {
		if seen[mode] {
			t.Errorf("Duplicate mode value: %v", mode)
		}
		seen[mode] = true
	}
}

// TestRollbackStrategy tests rollback strategy constants
func TestRollbackStrategy(t *testing.T) {
	strategies := []installer.RollbackStrategy{
		installer.RollbackNone,
		installer.RollbackPartial,
		installer.RollbackFull,
	}

	// Ensure strategies are distinct
	seen := make(map[installer.RollbackStrategy]bool)
	for _, strategy := range strategies {
		if seen[strategy] {
			t.Errorf("Duplicate strategy value: %v", strategy)
		}
		seen[strategy] = true
	}
}

package core_test

import (
	"context"
	"testing"
	"time"

	"github.com/mmso2016/setupkit/pkg/installer/core"
)

type contextKey string

func (c contextKey) String() string {
	return "core context key " + string(c)
}

// TestConfig tests configuration structures
func TestConfig(t *testing.T) {
	config := &core.Config{
		AppName:       "TestApp",
		Version:       "1.0.0",
		Publisher:     "TestCorp",
		Website:       "https://test.com",
		InstallDir:    "/opt/test",
		Mode:          core.ModeAuto,
		Rollback:      core.RollbackFull,
		DryRun:        false,
		Force:         true,
		Unattended:    false,
		AcceptLicense: true,
		LogFile:       "/tmp/test.log",
		LogLevel:      "debug",
		Verbose:       true,
	}

	// Test all fields to avoid unused warnings
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"AppName", config.AppName, "TestApp"},
		{"Version", config.Version, "1.0.0"},
		{"Publisher", config.Publisher, "TestCorp"},
		{"Website", config.Website, "https://test.com"},
		{"InstallDir", config.InstallDir, "/opt/test"},
		{"Mode", config.Mode, core.ModeAuto},
		{"Rollback", config.Rollback, core.RollbackFull},
		{"DryRun", config.DryRun, false},
		{"Force", config.Force, true},
		{"Unattended", config.Unattended, false},
		{"AcceptLicense", config.AcceptLicense, true},
		{"LogFile", config.LogFile, "/tmp/test.log"},
		{"LogLevel", config.LogLevel, "debug"},
		{"Verbose", config.Verbose, true},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestContext tests installer context
func TestContext(t *testing.T) {
	config := &core.Config{
		AppName: "TestApp",
	}

	logger := core.NewLogger("info", "")
	defer logger.Close() // Close logger at the end

	ctx := &core.Context{
		Config:      config,
		Logger:      logger,
		StartTime:   time.Now(),
		Checkpoints: []core.Checkpoint{},
		Metadata:    make(map[string]interface{}),
	}

	// Verify all fields are properly set
	if ctx.Config != config {
		t.Error("Config not properly set")
	}

	if ctx.Logger != logger {
		t.Error("Logger not properly set")
	}

	if ctx.StartTime.IsZero() {
		t.Error("StartTime not set")
	}

	// Test metadata storage
	ctx.Metadata["test_key"] = "test_value"

	if val, ok := ctx.Metadata["test_key"]; !ok || val != "test_value" {
		t.Error("Metadata storage failed")
	}

	// Test checkpoint
	checkpoint := core.Checkpoint{
		ID:        "test_checkpoint",
		Timestamp: time.Now(),
		State:     make(map[string]interface{}),
		Rollback: func() error {
			return nil
		},
	}

	ctx.Checkpoints = append(ctx.Checkpoints, checkpoint)
	if len(ctx.Checkpoints) != 1 {
		t.Error("Checkpoint not added")
	}
}

// TestPlatformInstaller tests platform installer creation
func TestPlatformInstaller(t *testing.T) {
	config := &core.Config{
		AppName:    "TestApp",
		InstallDir: "/tmp/test",
	}

	platform := core.CreatePlatformInstaller(config)
	if platform == nil {
		t.Fatal("CreatePlatformInstaller returned nil")
	}

	// Test basic methods (should not panic)
	_ = platform.Initialize()
	_ = platform.IsElevated()
	_ = platform.RequiresElevation()
}

// TestDefaultPlatformInstaller tests the default/fallback installer
func TestDefaultPlatformInstaller(t *testing.T) {
	config := &core.Config{
		AppName: "TestApp",
	}

	platform := core.NewDefaultPlatformInstaller(config)
	if platform == nil {
		t.Fatal("NewDefaultPlatformInstaller returned nil")
	}

	// Test all methods return expected defaults
	if err := platform.Initialize(); err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	if err := platform.CheckRequirements(); err != nil {
		t.Errorf("CheckRequirements() error = %v", err)
	}

	if platform.IsElevated() {
		t.Error("IsElevated() should return false for default")
	}

	if platform.RequiresElevation() {
		t.Error("RequiresElevation() should return false for default")
	}

	if err := platform.RequestElevation(); err != nil {
		t.Errorf("RequestElevation() error = %v", err)
	}

	if err := platform.RegisterWithOS(); err != nil {
		t.Errorf("RegisterWithOS() error = %v", err)
	}

	if err := platform.CreateShortcuts(); err != nil {
		t.Errorf("CreateShortcuts() error = %v", err)
	}

	if err := platform.RegisterUninstaller(); err != nil {
		t.Errorf("RegisterUninstaller() error = %v", err)
	}

	dirs := []string{"/test/bin"}
	if err := platform.UpdatePath(dirs, false); err != nil {
		t.Errorf("UpdatePath() error = %v", err)
	}

	if err := platform.AddToPath("/test/bin", false); err != nil {
		t.Errorf("AddToPath() error = %v", err)
	}

	if err := platform.RemoveFromPath("/test/bin", false); err != nil {
		t.Errorf("RemoveFromPath() error = %v", err)
	}

	if platform.IsInPath("/test/bin", false) {
		t.Error("IsInPath() should return false for default")
	}
}

// TestRollbackManager tests rollback functionality
func TestRollbackManager(t *testing.T) {
	rollback := core.NewRollbackManager(core.RollbackFull)

	if rollback == nil {
		t.Fatal("NewRollbackManager returned nil")
	}

	// Test adding checkpoints
	rollbackCalled := false
	rollback.AddCheckpoint("test1", func(ctx context.Context) error {
		rollbackCalled = true
		return nil
	})

	if rollback.Count() != 1 {
		t.Errorf("Count() = %d, want 1", rollback.Count())
	}

	// Add another checkpoint
	rollback.AddCheckpoint("test2", func(ctx context.Context) error {
		return nil
	})

	if rollback.Count() != 2 {
		t.Errorf("Count() = %d, want 2", rollback.Count())
	}

	// Test execute to verify rollbackCalled
	config := &core.Config{AppName: "Test"}
	logger := core.NewLogger("info", "")
	defer logger.Close() // Close logger at the end
	testCtx := &core.Context{
		Config: config,
		Logger: logger,
	}

	// Execute rollback to trigger the callback
	_ = rollback.Execute(testCtx)

	if !rollbackCalled {
		t.Error("Rollback callback was not called")
	}

	// Test clear
	rollback.Clear()
	if rollback.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", rollback.Count())
	}
}

// TestProgress tests progress reporting
func TestProgress(t *testing.T) {
	progress := &core.Progress{
		TotalComponents:   3,
		CurrentComponent:  1,
		ComponentName:     "Core Files",
		ComponentProgress: 50.0,
		OverallProgress:   16.67,
		Message:           "Installing core files...",
		IsError:           false,
	}

	// Test all fields
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"TotalComponents", progress.TotalComponents, 3},
		{"CurrentComponent", progress.CurrentComponent, 1},
		{"ComponentName", progress.ComponentName, "Core Files"},
		{"ComponentProgress", progress.ComponentProgress, 50.0},
		{"OverallProgress", progress.OverallProgress, 16.67},
		{"Message", progress.Message, "Installing core files..."},
		{"IsError", progress.IsError, false},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestInstallSummary tests installation summary
func TestInstallSummary(t *testing.T) {
	summary := &core.InstallSummary{
		Success:             true,
		Duration:            5 * time.Minute,
		ComponentsInstalled: []string{"Core", "Docs", "Examples"},
		InstallPath:         "/opt/testapp",
		Warnings:            []string{"PATH not updated"},
		NextSteps:           []string{"Run 'testapp' to start"},
	}

	if !summary.Success {
		t.Error("Success should be true")
	}

	if summary.Duration != 5*time.Minute {
		t.Errorf("Duration = %v, want 5m", summary.Duration)
	}

	if len(summary.ComponentsInstalled) != 3 {
		t.Errorf("ComponentsInstalled length = %d, want 3",
			len(summary.ComponentsInstalled))
	}

	if summary.InstallPath != "/opt/testapp" {
		t.Errorf("InstallPath = %s, want /opt/testapp", summary.InstallPath)
	}

	// Verify warnings and next steps are set
	if len(summary.Warnings) != 1 || summary.Warnings[0] != "PATH not updated" {
		t.Error("Warnings not properly set")
	}

	if len(summary.NextSteps) != 1 || summary.NextSteps[0] != "Run 'testapp' to start" {
		t.Error("NextSteps not properly set")
	}
}

// TestCheckDiskSpace tests disk space checking
func TestCheckDiskSpace(t *testing.T) {
	// This is a basic test - actual implementation depends on platform
	tempDir := t.TempDir()
	requiredSpace := int64(1024) // 1KB should always be available

	err := core.CheckDiskSpace(tempDir, requiredSpace)
	if err != nil {
		t.Errorf("CheckDiskSpace() error = %v for 1KB", err)
	}

	// Test with huge requirement (should fail)
	hugeSpace := int64(1 << 50) // 1PB - should fail
	err = core.CheckDiskSpace(tempDir, hugeSpace)
	if err == nil {
		t.Error("CheckDiskSpace() should fail for 1PB requirement")
	}
}

// TestSimpleLogger tests the simple logger implementation
func TestSimpleLogger(t *testing.T) {
	// Test logger without file
	logger := core.NewLogger("debug", "")
	defer logger.Close() // Close logger at the end

	// These should not panic
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")
	logger.Verbose("verbose message")
	logger.VerboseSection("Test Section")

	// Test verbose toggle
	logger.SetVerbose(true)
	logger.Verbose("verbose enabled")

	logger.SetVerbose(false)
	logger.Verbose("verbose disabled")

	// Test with log file
	tmpDir := t.TempDir()
	logFile := tmpDir + "/test.log"
	fileLogger := core.NewLogger("info", logFile)
	defer fileLogger.Close() // Important: Close file logger to release file handle

	fileLogger.Info("test log entry")

	// Logger should create the file
	// Note: Implementation might buffer, so we just check it doesn't panic
}

// TestComponentExecution tests component installer/uninstaller execution
func TestComponentExecution(t *testing.T) {
	installerCalled := false
	uninstallerCalled := false

	component := core.Component{
		ID:          "test",
		Name:        "Test Component",
		Description: "Test component",
		Required:    false,
		Selected:    true,
		Validator: func() error {
			return nil
		},
		Installer: func(ctx context.Context) error {
			installerCalled = true
			// Check context values
			if ctx.Value(contextKey("test_key")) != "test_value" {
				t.Error("Context value not found")
			}
			return nil
		},
		Uninstaller: func(ctx context.Context) error {
			uninstallerCalled = true
			return nil
		},
	}

	// Verify component fields are set correctly
	if component.ID != "test" {
		t.Errorf("ID = %v, want test", component.ID)
	}

	if component.Name != "Test Component" {
		t.Errorf("Name = %v, want Test Component", component.Name)
	}

	if component.Description != "Test component" {
		t.Errorf("Description = %v, want Test component", component.Description)
	}

	if component.Required {
		t.Error("Required should be false")
	}

	if !component.Selected {
		t.Error("Selected should be true")
	}

	// Test validator
	if err := component.Validator(); err != nil {
		t.Errorf("Validator() error = %v", err)
	}

	// Create context with test value
	ctx := context.WithValue(context.Background(), contextKey("test_key"), "test_value")

	// Execute installer
	err := component.Installer(ctx)
	if err != nil {
		t.Errorf("Installer() error = %v", err)
	}
	if !installerCalled {
		t.Error("Installer was not called")
	}

	// Execute uninstaller
	err = component.Uninstaller(ctx)
	if err != nil {
		t.Errorf("Uninstaller() error = %v", err)
	}
	if !uninstallerCalled {
		t.Error("Uninstaller was not called")
	}
}

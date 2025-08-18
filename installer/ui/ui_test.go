package ui_test

import (
	"fmt"
	"testing"

	"github.com/mmso2016/setupkit/installer/core"
	"github.com/mmso2016/setupkit/installer/ui"
)

// TestSilentUI tests the silent UI implementation
func TestSilentUI(t *testing.T) {
	silentUI := ui.NewSilentUI()

	if silentUI == nil {
		t.Fatal("NewSilentUI() returned nil")
	}

	// Create test context
	config := &core.Config{
		AppName:       "TestApp",
		Version:       "1.0.0",
		InstallDir:    "/opt/test",
		AcceptLicense: true,
	}

	logger := core.NewLogger("info", "")

	ctx := &core.Context{
		Config:   config,
		Logger:   logger,
		Metadata: make(map[string]interface{}),
	}

	// Test Initialize
	err := silentUI.Initialize(ctx)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Test ShowWelcome
	err = silentUI.ShowWelcome()
	if err != nil {
		t.Errorf("ShowWelcome() error = %v", err)
	}

	// Test ShowLicense with accepted license
	accepted, err := silentUI.ShowLicense("Test License")
	if err != nil {
		t.Errorf("ShowLicense() error = %v", err)
	}
	if !accepted {
		t.Error("ShowLicense() should return true when AcceptLicense is true")
	}

	// Test with declined license
	config.AcceptLicense = false
	accepted, err = silentUI.ShowLicense("Test License")
	if err != nil {
		t.Errorf("ShowLicense() error = %v", err)
	}
	if accepted {
		t.Error("ShowLicense() should return false when AcceptLicense is false")
	}
	config.AcceptLicense = true // Reset

	// Test SelectComponents
	components := []core.Component{
		{ID: "core", Name: "Core", Required: true, Selected: true},
		{ID: "docs", Name: "Docs", Required: false, Selected: false},
		{ID: "examples", Name: "Examples", Required: false, Selected: true},
	}

	selected, err := silentUI.SelectComponents(components)
	if err != nil {
		t.Errorf("SelectComponents() error = %v", err)
	}

	// Should return only selected or required components
	if len(selected) != 2 {
		t.Errorf("SelectComponents() returned %d components, want 2", len(selected))
	}

	// Test SelectInstallPath
	path, err := silentUI.SelectInstallPath("/default/path")
	if err != nil {
		t.Errorf("SelectInstallPath() error = %v", err)
	}
	if path != config.InstallDir {
		t.Errorf("SelectInstallPath() = %v, want %v", path, config.InstallDir)
	}

	// Test with empty config path
	config.InstallDir = ""
	path, err = silentUI.SelectInstallPath("/default/path")
	if err != nil {
		t.Errorf("SelectInstallPath() error = %v", err)
	}
	if path != "/default/path" {
		t.Errorf("SelectInstallPath() = %v, want /default/path", path)
	}

	// Test ShowProgress
	progress := &core.Progress{
		ComponentName:   "Core",
		OverallProgress: 50.0,
	}
	err = silentUI.ShowProgress(progress)
	if err != nil {
		t.Errorf("ShowProgress() error = %v", err)
	}

	// Test ShowError
	retry, err := silentUI.ShowError(fmt.Errorf("test error"), true)
	if err != nil {
		t.Errorf("ShowError() error = %v", err)
	}
	if retry {
		t.Error("ShowError() should not retry in silent mode")
	}

	// Test ShowSuccess
	summary := &core.InstallSummary{
		Success:     true,
		InstallPath: "/opt/test",
	}
	err = silentUI.ShowSuccess(summary)
	if err != nil {
		t.Errorf("ShowSuccess() error = %v", err)
	}

	// Test RequestElevation
	granted, err := silentUI.RequestElevation("test reason")
	if err != nil {
		t.Errorf("RequestElevation() error = %v", err)
	}
	if !granted {
		t.Error("RequestElevation() should return true in silent mode")
	}

	// Test Shutdown
	err = silentUI.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

// TestUIFactory tests UI factory creation
func TestUIFactory(t *testing.T) {
	tests := []struct {
		name    string
		mode    core.Mode
		wantErr bool
	}{
		{
			name:    "silent mode",
			mode:    core.ModeSilent,
			wantErr: false,
		},
		{
			name:    "CLI mode",
			mode:    core.ModeCLI,
			wantErr: false,
		},
		{
			name:    "auto mode",
			mode:    core.ModeAuto,
			wantErr: false,
		},
		// GUI mode will fail unless compiled with wails tag
		{
			name:    "GUI mode",
			mode:    core.ModeGUI,
			wantErr: true, // Expected to fail without wails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui, err := ui.CreateUI(tt.mode)
			if tt.wantErr {
				if err == nil {
					t.Error("CreateUI() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("CreateUI() unexpected error = %v", err)
				}
				if ui == nil {
					t.Error("CreateUI() returned nil without error")
				}
			}
		})
	}
}

// TestHasDisplay tests display detection
func TestHasDisplay(t *testing.T) {
	// This is environment-dependent, just ensure it doesn't panic
	// and returns a boolean
	_ = ui.HasDisplay()
}

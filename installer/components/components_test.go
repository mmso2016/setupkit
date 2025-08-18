package components_test

import (
	"context"
	"testing"
	
	"github.com/mmso2016/setupkit/installer"
	"github.com/mmso2016/setupkit/installer/components"
	"github.com/mmso2016/setupkit/installer/core"
)

// MockPlatformInstaller implements a mock platform installer for testing
type MockPlatformInstaller struct {
	elevated       bool
	pathEntries    map[string]bool
	addToPathErr   error
	removeFromPathErr error
}

func NewMockPlatformInstaller() *MockPlatformInstaller {
	return &MockPlatformInstaller{
		elevated:    false,
		pathEntries: make(map[string]bool),
	}
}

func (m *MockPlatformInstaller) Initialize() error                          { return nil }
func (m *MockPlatformInstaller) CheckRequirements() error                   { return nil }
func (m *MockPlatformInstaller) IsElevated() bool                          { return m.elevated }
func (m *MockPlatformInstaller) RequiresElevation() bool                   { return false }
func (m *MockPlatformInstaller) RequestElevation() error                   { return nil }
func (m *MockPlatformInstaller) RegisterWithOS() error                     { return nil }
func (m *MockPlatformInstaller) CreateShortcuts() error                    { return nil }
func (m *MockPlatformInstaller) RegisterUninstaller() error                { return nil }
func (m *MockPlatformInstaller) UpdatePath(dirs []string, system bool) error { return nil }

func (m *MockPlatformInstaller) AddToPath(dir string, system bool) error {
	if m.addToPathErr != nil {
		return m.addToPathErr
	}
	key := dir
	if system {
		key = "system:" + dir
	}
	m.pathEntries[key] = true
	return nil
}

func (m *MockPlatformInstaller) RemoveFromPath(dir string, system bool) error {
	if m.removeFromPathErr != nil {
		return m.removeFromPathErr
	}
	key := dir
	if system {
		key = "system:" + dir
	}
	delete(m.pathEntries, key)
	return nil
}

func (m *MockPlatformInstaller) IsInPath(dir string, system bool) bool {
	key := dir
	if system {
		key = "system:" + dir
	}
	return m.pathEntries[key]
}

// TestPathComponent tests PATH component functionality
func TestPathComponent(t *testing.T) {
	tests := []struct {
		name      string
		directory string
		scope     components.PathScope
		wantErr   bool
	}{
		{
			name:      "user path",
			directory: "/usr/local/bin",
			scope:     components.PathScopeUser,
			wantErr:   false,
		},
		{
			name:      "system path",
			directory: "/opt/app/bin",
			scope:     components.PathScopeSystem,
			wantErr:   false,
		},
		{
			name:      "auto scope",
			directory: "/usr/local/bin",
			scope:     components.PathScopeAuto,
			wantErr:   false,
		},
		{
			name:      "empty directory",
			directory: "",
			scope:     components.PathScopeUser,
			wantErr:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := components.NewPathComponent(tt.directory, tt.scope)
			
			// Test validation
			err := pc.Validator()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validator() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if tt.wantErr {
				return // Skip further tests for invalid configs
			}
			
			// Test component metadata
			if pc.ID != "path-configuration" {
				t.Errorf("ID = %v, want path-configuration", pc.ID)
			}
			
			if pc.Name != "PATH Environment Variable" {
				t.Errorf("Name = %v, want PATH Environment Variable", pc.Name)
			}
		})
	}
}

// TestPathComponentInstall tests PATH installation
func TestPathComponentInstall(t *testing.T) {
	mockPlatform := NewMockPlatformInstaller()
	logger := core.NewLogger("info", "")
	
	pc := components.NewPathComponent("/test/bin", components.PathScopeUser)
	
	// Create context with platform and logger
	ctx := context.Background()
	ctx = context.WithValue(ctx, "platform", mockPlatform)
	ctx = context.WithValue(ctx, "logger", logger)
	
	// Install
	err := pc.Installer(ctx)
	if err != nil {
		t.Errorf("Installer() error = %v", err)
	}
	
	// Verify path was added
	if !mockPlatform.IsInPath("/test/bin", false) {
		t.Error("Path was not added")
	}
	
	// Test duplicate prevention
	pc.SkipDuplicate = true
	err = pc.Installer(ctx)
	if err != nil {
		t.Errorf("Installer() with duplicate error = %v", err)
	}
}

// TestPathComponentUninstall tests PATH removal
func TestPathComponentUninstall(t *testing.T) {
	mockPlatform := NewMockPlatformInstaller()
	logger := core.NewLogger("info", "")
	
	// Pre-add path
	mockPlatform.AddToPath("/test/bin", false)
	
	pc := components.NewPathComponent("/test/bin", components.PathScopeUser)
	
	// Create context
	ctx := context.Background()
	ctx = context.WithValue(ctx, "platform", mockPlatform)
	ctx = context.WithValue(ctx, "logger", logger)
	
	// Uninstall
	err := pc.Uninstaller(ctx)
	if err != nil {
		t.Errorf("Uninstaller() error = %v", err)
	}
	
	// Verify path was removed
	if mockPlatform.IsInPath("/test/bin", false) {
		t.Error("Path was not removed")
	}
}

// TestBinaryComponent tests binary component
func TestBinaryComponent(t *testing.T) {
	bc := components.NewBinaryComponent(
		"myapp",
		"/usr/local/bin",
		components.BinaryOptions{
			ExecutableName: "myapp",
			Permissions:    0755,
		},
	)
	
	// Test metadata
	if bc.ID != "binary-myapp" {
		t.Errorf("ID = %v, want binary-myapp", bc.ID)
	}
	
	if bc.Name != "Binary: myapp" {
		t.Errorf("Name = %v, want Binary: myapp", bc.Name)
	}
	
	// Test validation
	err := bc.Validator()
	if err != nil {
		t.Errorf("Validator() error = %v", err)
	}
	
	// Test empty source validation
	bcEmpty := components.NewBinaryComponent(
		"",
		"/usr/local/bin",
		components.BinaryOptions{},
	)
	
	err = bcEmpty.Validator()
	if err == nil {
		t.Error("Validator() should fail for empty source")
	}
}

// TestConfigComponent tests configuration component
func TestConfigComponent(t *testing.T) {
	configContent := `# Configuration
app.name=TestApp
app.port=8080
`
	
	cc := components.NewConfigComponent(
		"app.conf",
		"/etc/myapp",
		configContent,
	)
	
	// Test metadata
	if cc.ID != "config-app.conf" {
		t.Errorf("ID = %v, want config-app.conf", cc.ID)
	}
	
	if cc.Name != "Configuration: app.conf" {
		t.Errorf("Name = %v, want Configuration: app.conf", cc.Name)
	}
	
	// Test properties
	if cc.FileName != "app.conf" {
		t.Errorf("FileName = %v, want app.conf", cc.FileName)
	}
	
	if cc.DestDir != "/etc/myapp" {
		t.Errorf("DestDir = %v, want /etc/myapp", cc.DestDir)
	}
	
	if cc.Content != configContent {
		t.Errorf("Content = %v, want %v", cc.Content, configContent)
	}
	
	if cc.Overwrite {
		t.Error("Overwrite should be false by default")
	}
	
	if cc.Permissions != 0644 {
		t.Errorf("Permissions = %v, want 0644", cc.Permissions)
	}
}

// TestShortcutComponent tests shortcut component
func TestShortcutComponent(t *testing.T) {
	sc := components.NewShortcutComponent(
		"MyApp",
		"/opt/myapp/bin/myapp",
		components.ShortcutOptions{
			CreateDesktop:     true,
			CreateStartMenu:   true,
			CreateQuickLaunch: false,
			IconPath:         "/opt/myapp/icon.png",
			Arguments:        "--start",
			Description:      "My Application",
		},
	)
	
	// Test component metadata
	if sc.ID != "shortcut-MyApp" {
		t.Errorf("ID = %v, want shortcut-MyApp", sc.ID)
	}
	
	if sc.Name != "Shortcuts: MyApp" {
		t.Errorf("Component Name = %v, want Shortcuts: MyApp", sc.Name)
	}
	
	// Test shortcut properties
	if sc.ShortcutName != "MyApp" {
		t.Errorf("ShortcutName = %v, want MyApp", sc.ShortcutName)
	}
	
	if sc.TargetPath != "/opt/myapp/bin/myapp" {
		t.Errorf("TargetPath = %v, want /opt/myapp/bin/myapp", sc.TargetPath)
	}
	
	if !sc.Options.CreateDesktop {
		t.Error("CreateDesktop should be true")
	}
	
	if !sc.Options.CreateStartMenu {
		t.Error("CreateStartMenu should be true")
	}
	
	if sc.Options.CreateQuickLaunch {
		t.Error("CreateQuickLaunch should be false")
	}
	
	if sc.Options.IconPath != "/opt/myapp/icon.png" {
		t.Errorf("IconPath = %v, want /opt/myapp/icon.png", sc.Options.IconPath)
	}
	
	if sc.Options.Arguments != "--start" {
		t.Errorf("Arguments = %v, want --start", sc.Options.Arguments)
	}
	
	if sc.Options.Description != "My Application" {
		t.Errorf("Description = %v, want My Application", sc.Options.Description)
	}
}

// TestAdvancedPathComponent tests advanced PATH component
func TestAdvancedPathComponent(t *testing.T) {
	apc := components.NewAdvancedPathComponent(
		"/opt/app/bin",
		components.PathScopeSystem,
		components.PathOptions{
			PrependPath:    true,
			CreateBackup:   true,
			NotifyUser:     true,
			RequireRestart: true,
		},
	)
	
	// Test that description includes restart notice
	if apc.Description != "Add /opt/app/bin to PATH (restart required)" {
		t.Errorf("Description = %v, want restart notice", apc.Description)
	}
	
	// Test basic properties inherited
	if apc.Directory != "/opt/app/bin" {
		t.Errorf("Directory = %v, want /opt/app/bin", apc.Directory)
	}
	
	if apc.Scope != components.PathScopeSystem {
		t.Errorf("Scope = %v, want PathScopeSystem", apc.Scope)
	}
}

// TestComponentIntegration tests component integration with installer
func TestComponentIntegration(t *testing.T) {
	// Create various components
	pathComp := components.NewPathComponent("/app/bin", components.PathScopeUser)
	binComp := components.NewBinaryComponent(
		"app",
		"/app/bin",
		components.BinaryOptions{
			ExecutableName: "app",
		},
	)
	configComp := components.NewConfigComponent(
		"app.conf",
		"/app/etc",
		"config=value",
	)
	
	// Create installer with components
	inst, err := installer.New(
		installer.WithAppName("TestApp"),
		installer.WithComponents(
			pathComp.Component,
			binComp.Component,
			configComp.Component,
		),
	)
	
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}
	
	// Verify components were added
	components := inst.GetComponents()
	if len(components) != 3 {
		t.Errorf("GetComponents() returned %d components, want 3", len(components))
	}
	
	// Verify component IDs
	expectedIDs := map[string]bool{
		"path-configuration": true,
		"binary-app":        true,
		"config-app.conf":   true,
	}
	
	for _, comp := range components {
		if !expectedIDs[comp.ID] {
			t.Errorf("Unexpected component ID: %v", comp.ID)
		}
		delete(expectedIDs, comp.ID)
	}
	
	if len(expectedIDs) > 0 {
		t.Errorf("Missing components: %v", expectedIDs)
	}
}

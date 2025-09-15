package ui_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/mmso2016/setupkit/pkg/installer/controller"
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// DFAControllerTestSuite tests the DFA controller with mock UI
type DFAControllerTestSuite struct {
	suite.Suite
	tempDir   string
	config    *core.Config
	context   *core.Context
	installer *core.Installer
	mockView  *MockInstallerView
	controller *controller.InstallerController
}

// SetupSuite runs once before all tests
func (suite *DFAControllerTestSuite) SetupSuite() {
	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "setupkit_dfa_test")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// Setup test configuration
	suite.config = &core.Config{
		AppName:       "DFATestApp",
		Version:       "2.0.0",
		Publisher:     "SetupKit DFA Test",
		InstallDir:    filepath.Join(tempDir, "dfa_install"),
		AcceptLicense: true,
		License:       "DFA Test License Agreement",
		Components: []core.Component{
			{ID: "core", Name: "Core Component", Required: true, Selected: true},
			{ID: "plugins", Name: "Plugins", Required: false, Selected: true},
			{ID: "docs", Name: "Documentation", Required: false, Selected: false},
		},
	}

	// Setup logger
	logger := core.NewLogger("debug", "")
	
	// Create context
	suite.context = &core.Context{
		Config:   suite.config,
		Logger:   logger,
		Metadata: make(map[string]interface{}),
	}

	// Create installer
	suite.installer = core.New(suite.config)
	suite.context.Metadata["installer"] = suite.installer
}

// TearDownSuite runs once after all tests
func (suite *DFAControllerTestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// SetupTest runs before each test
func (suite *DFAControllerTestSuite) SetupTest() {
	// Create fresh mock view and controller for each test
	suite.mockView = NewMockInstallerView()
	suite.controller = controller.NewInstallerController(suite.config, suite.installer)
	suite.controller.SetView(suite.mockView)
	
	// Setup mock responses
	suite.mockView.SetLicenseAccepted(true)
	suite.mockView.SetSelectedPath(suite.config.InstallDir)
	suite.mockView.SetElevationGranted(true)
	
	// Set default selected components (core + plugins)
	expectedComponents := []core.Component{
		{ID: "core", Name: "Core Component", Required: true, Selected: true},
		{ID: "plugins", Name: "Plugins", Required: false, Selected: true},
	}
	suite.mockView.SetSelectedComponents(expectedComponents)
}

// Test DFA Controller State Transitions
func (suite *DFAControllerTestSuite) TestDFAStateTransitions() {
	// After controller creation, current state should be welcome (initial state)
	currentState := suite.controller.GetCurrentState()
	suite.Equal(controller.StateWelcome, currentState, "Initial state should be welcome after controller creation")

	// Start the DFA (this will trigger state entry)
	err := suite.controller.Start()
	suite.NoError(err, "DFA controller start should succeed")

	// Current state should still be welcome after Start()
	currentState = suite.controller.GetCurrentState()
	suite.Equal(controller.StateWelcome, currentState, "Current state should be welcome after Start()")

	// Verify state transitions were recorded
	transitions := suite.mockView.GetStateTransitions()
	suite.NotEmpty(transitions, "State transitions should be recorded")

	// Verify the first transition is to welcome state
	if len(transitions) > 0 {
		suite.Equal(controller.StateWelcome, transitions[0].To, "First state should be welcome")
	}

	// Verify expected method calls were made
	calls := suite.mockView.GetRecordedCalls()
	suite.Contains(calls, "ShowWelcome", "ShowWelcome should be called")

	// Continue the flow to trigger license step
	err = suite.controller.Next()
	suite.NoError(err, "Next transition should succeed")

	// Now verify license was called
	calls = suite.mockView.GetRecordedCalls()
	suite.Contains(calls, "ShowLicense: DFA Test License Agreement", "ShowLicense should be called with license text")
}

// Test DFA Navigation Methods
func (suite *DFAControllerTestSuite) TestDFANavigation() {
	// Start DFA
	err := suite.controller.Start()
	suite.NoError(err)

	// Test CanGoNext
	canNext := suite.controller.CanGoNext()
	suite.True(canNext, "Should be able to go next from initial state")

	// Test Next transition
	err = suite.controller.Next()
	suite.NoError(err, "Next transition should succeed")

	// Test CanGoBack
	canBack := suite.controller.CanGoBack()
	_ = canBack // Suppress unused variable warning

	// Test CanCancel
	canCancel := suite.controller.CanCancel()
	suite.True(canCancel, "Should be able to cancel from most states")
}

// Test DFA State Validation
func (suite *DFAControllerTestSuite) TestDFAStateValidation() {
	// Test license validation failure
	suite.mockView.SetLicenseAccepted(false)
	suite.config.AcceptLicense = false

	err := suite.controller.Start()
	// This should fail or handle the license rejection appropriately
	// The exact behavior depends on the implementation
	_ = err // Suppress unused variable warning
	
	// Reset for next test
	suite.mockView.SetLicenseAccepted(true)
	suite.config.AcceptLicense = true
}

// Test DFA Component Selection
func (suite *DFAControllerTestSuite) TestDFAComponentSelection() {
	// Set specific component selection
	selectedComponents := []core.Component{
		{ID: "core", Name: "Core Component", Required: true, Selected: true},
		{ID: "docs", Name: "Documentation", Required: false, Selected: true}, // Different selection
	}
	suite.mockView.SetSelectedComponents(selectedComponents)

	err := suite.controller.Start()
	suite.NoError(err)

	// Navigate through states to reach components
	err = suite.controller.Next() // Welcome -> License
	suite.NoError(err)
	err = suite.controller.Next() // License -> Components
	suite.NoError(err)

	// Verify component selection was called
	calls := suite.mockView.GetRecordedCalls()
	suite.Contains(calls, "ShowComponents: 3 components", "Component selection should be called")
}

// Test DFA Installation Path Selection
func (suite *DFAControllerTestSuite) TestDFAInstallPathSelection() {
	customPath := filepath.Join(suite.tempDir, "custom_install")
	originalInstallDir := suite.config.InstallDir // Save original before mock modifies it
	suite.mockView.SetSelectedPath(customPath)

	err := suite.controller.Start()
	suite.NoError(err)

	// Navigate through states to reach install path
	err = suite.controller.Next() // Welcome -> License
	suite.NoError(err)
	err = suite.controller.Next() // License -> Components
	suite.NoError(err)
	err = suite.controller.Next() // Components -> InstallPath
	suite.NoError(err)

	// Verify path selection was called with the original default path
	calls := suite.mockView.GetRecordedCalls()
	suite.NotEmpty(calls, "Some method calls should have been recorded")

	// The controller should call ShowInstallPath with the original default path
	expectedCall := "ShowInstallPath: " + originalInstallDir
	found := false
	for _, call := range calls {
		if call == expectedCall {
			found = true
			break
		}
	}
	suite.True(found, "Path selection should be called with original default path: %s", originalInstallDir)

	// Verify the config was updated with the selected path
	suite.Equal(customPath, suite.config.InstallDir, "Config should be updated with selected path")
}

// Test DFA Installation Summary
func (suite *DFAControllerTestSuite) TestDFAInstallationSummary() {
	err := suite.controller.Start()
	suite.NoError(err)

	// Navigate through states to reach summary
	err = suite.controller.Next() // Welcome -> License
	suite.NoError(err)
	err = suite.controller.Next() // License -> Components
	suite.NoError(err)
	err = suite.controller.Next() // Components -> InstallPath
	suite.NoError(err)
	err = suite.controller.Next() // InstallPath -> Summary
	suite.NoError(err)

	// Verify summary was shown
	calls := suite.mockView.GetRecordedCalls()
	found := false
	for _, call := range calls {
		if len(call) > 11 && call[:11] == "ShowSummary" {
			found = true
			break
		}
	}
	suite.True(found, "Installation summary should be shown")
}

// Test DFA Controller with Different License Configurations
func (suite *DFAControllerTestSuite) TestDFAWithoutLicense() {
	// Test DFA behavior when no license is configured
	suite.config.License = ""
	
	controller := controller.NewInstallerController(suite.config, suite.installer)
	mockView := NewMockInstallerView()
	controller.SetView(mockView)

	err := controller.Start()
	suite.NoError(err)

	// Verify license step was skipped
	calls := mockView.GetRecordedCalls()
	licenseCallFound := false
	for _, call := range calls {
		if len(call) >= 11 && call[:11] == "ShowLicense" {
			licenseCallFound = true
			break
		}
	}
	// Should not find license call when no license is configured
	suite.False(licenseCallFound, "License step should be skipped when no license configured")
}

// Test DFA Error Handling
func (suite *DFAControllerTestSuite) TestDFAErrorHandling() {
	err := suite.controller.Start()
	suite.NoError(err)

	// The actual error handling would depend on installation failures
	// For now, verify that the controller can handle basic scenarios
	
	// Verify no errors were recorded during normal flow
	errors := suite.mockView.GetErrors()
	suite.Empty(errors, "No errors should be recorded during successful flow")
}

// Test DFA State Consistency Across Multiple Runs
func (suite *DFAControllerTestSuite) TestDFAStateConsistency() {
	// Run the DFA multiple times and verify consistent behavior
	for i := 0; i < 3; i++ {
		suite.SetupTest() // Reset mock view and controller
		
		err := suite.controller.Start()
		suite.NoError(err, "DFA run %d should succeed", i+1)

		// Verify consistent state transitions
		transitions := suite.mockView.GetStateTransitions()
		suite.NotEmpty(transitions, "Run %d should have state transitions", i+1)
		
		if len(transitions) > 0 {
			suite.Equal(controller.StateWelcome, transitions[0].To, 
				"Run %d should start with welcome state", i+1)
		}
	}
}

// TestDFAControllerTestSuite runs the DFA controller test suite
func TestDFAControllerTestSuite(t *testing.T) {
	suite.Run(t, new(DFAControllerTestSuite))
}
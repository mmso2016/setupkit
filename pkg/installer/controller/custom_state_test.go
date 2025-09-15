package controller

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// CustomStateTestSuite tests custom state functionality
type CustomStateTestSuite struct {
	suite.Suite
	tempDir   string
	config    *core.Config
	context   *core.Context
	installer *core.Installer
	controller *InstallerController
	mockView  *MockExtendedInstallerView
}

// MockExtendedInstallerView implements both InstallerView and ExtendedInstallerView for testing
type MockExtendedInstallerView struct {
	recordedCalls    []string
	customStateData  map[wizard.State]CustomStateData
	shouldReturnData CustomStateData
}

func NewMockExtendedInstallerView() *MockExtendedInstallerView {
	return &MockExtendedInstallerView{
		recordedCalls:   make([]string, 0),
		customStateData: make(map[wizard.State]CustomStateData),
	}
}

// InstallerView interface methods (simplified)
func (m *MockExtendedInstallerView) ShowWelcome() error {
	m.recordedCalls = append(m.recordedCalls, "ShowWelcome")
	return nil
}

func (m *MockExtendedInstallerView) ShowLicense(license string) (accepted bool, err error) {
	m.recordedCalls = append(m.recordedCalls, "ShowLicense")
	return true, nil
}

func (m *MockExtendedInstallerView) ShowComponents(components []core.Component) (selected []core.Component, err error) {
	m.recordedCalls = append(m.recordedCalls, "ShowComponents")
	return components, nil
}

func (m *MockExtendedInstallerView) ShowInstallPath(defaultPath string) (path string, err error) {
	m.recordedCalls = append(m.recordedCalls, "ShowInstallPath")
	return defaultPath, nil
}

func (m *MockExtendedInstallerView) ShowSummary(config *core.Config, selectedComponents []core.Component, installPath string) (proceed bool, err error) {
	m.recordedCalls = append(m.recordedCalls, "ShowSummary")
	return true, nil
}

func (m *MockExtendedInstallerView) ShowProgress(progress *core.Progress) error {
	m.recordedCalls = append(m.recordedCalls, "ShowProgress")
	return nil
}

func (m *MockExtendedInstallerView) ShowComplete(summary *core.InstallSummary) error {
	m.recordedCalls = append(m.recordedCalls, "ShowComplete")
	return nil
}

func (m *MockExtendedInstallerView) ShowErrorMessage(err error) error {
	m.recordedCalls = append(m.recordedCalls, "ShowErrorMessage")
	return nil
}

func (m *MockExtendedInstallerView) OnStateChanged(oldState, newState wizard.State) error {
	m.recordedCalls = append(m.recordedCalls, "OnStateChanged")
	return nil
}

// ExtendedInstallerView interface method
func (m *MockExtendedInstallerView) ShowCustomState(stateID wizard.State, data CustomStateData) (CustomStateData, error) {
	m.recordedCalls = append(m.recordedCalls, "ShowCustomState:"+string(stateID))
	m.customStateData[stateID] = data

	// Return configured data or default behavior
	if m.shouldReturnData != nil {
		return m.shouldReturnData, nil
	}

	// Default behavior for database config
	if stateID == StateDBConfig {
		defaultConfig := DefaultDatabaseConfig()
		return CustomStateData{"config": defaultConfig}, nil
	}

	return data, nil
}

// Test helper methods
func (m *MockExtendedInstallerView) GetRecordedCalls() []string {
	return m.recordedCalls
}

func (m *MockExtendedInstallerView) GetCustomStateData(stateID wizard.State) (CustomStateData, bool) {
	data, exists := m.customStateData[stateID]
	return data, exists
}

func (m *MockExtendedInstallerView) SetReturnData(data CustomStateData) {
	m.shouldReturnData = data
}

// SetupSuite runs once before all tests
func (suite *CustomStateTestSuite) SetupSuite() {
	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "custom_state_test")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// Setup test configuration
	suite.config = &core.Config{
		AppName:       "CustomStateTestApp",
		Version:       "1.0.0",
		Publisher:     "CustomState Test",
		InstallDir:    filepath.Join(tempDir, "install"),
		AcceptLicense: true,
		License:       "Test License Agreement",
		Components: []core.Component{
			{ID: "core", Name: "Core Component", Required: true, Selected: true},
			{ID: "db", Name: "Database Component", Required: false, Selected: true},
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
func (suite *CustomStateTestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// SetupTest runs before each test
func (suite *CustomStateTestSuite) SetupTest() {
	suite.controller = NewInstallerController(suite.config, suite.installer)
	suite.mockView = NewMockExtendedInstallerView()
	suite.controller.SetView(suite.mockView)
}

// TestCustomStateRegistry tests the custom state registry functionality
func (suite *CustomStateTestSuite) TestCustomStateRegistry() {
	registry := NewCustomStateRegistry()

	// Test registration
	handler := NewDatabaseConfigHandler()
	err := registry.Register(handler)
	suite.NoError(err, "Should be able to register custom state")

	// Test duplicate registration
	err = registry.Register(handler)
	suite.Error(err, "Should not allow duplicate registration")

	// Test retrieval
	retrieved, exists := registry.GetHandler(StateDBConfig)
	suite.True(exists, "Should find registered handler")
	suite.Equal(handler, retrieved, "Should return the same handler")

	// Test non-existent state
	_, exists = registry.GetHandler("non-existent")
	suite.False(exists, "Should not find non-existent handler")

	// Test GetAll
	all := registry.GetAll()
	suite.Len(all, 1, "Should return all registered handlers")
	suite.Equal(handler, all[0], "Should return registered handler")
}

// TestDatabaseConfigHandler tests the database configuration handler
func (suite *CustomStateTestSuite) TestDatabaseConfigHandler() {
	handler := NewDatabaseConfigHandler()

	// Test basic properties
	suite.Equal(StateDBConfig, handler.GetStateID())
	suite.Equal("Database Configuration", handler.GetConfig().Name)
	suite.Equal(InsertAfterInstallPath, handler.GetInsertionPoint())

	// Test validation with valid config
	validConfig := &DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
		UseSSL:   false,
	}

	data := map[string]interface{}{
		"db_config": validConfig,
	}

	err := handler.Validate(suite.controller, data)
	suite.NoError(err, "Valid config should pass validation")

	// Test validation with invalid config
	invalidConfig := &DatabaseConfig{
		Type:     "unsupported",
		Host:     "",
		Port:     -1,
		Database: "",
	}

	data["db_config"] = invalidConfig
	err = handler.Validate(suite.controller, data)
	suite.Error(err, "Invalid config should fail validation")
}

// TestCustomStateIntegration tests the full custom state integration
func (suite *CustomStateTestSuite) TestCustomStateIntegration() {
	// Register database config handler
	handler := NewDatabaseConfigHandler()
	err := suite.controller.RegisterCustomState(handler)
	suite.NoError(err, "Should register custom state successfully")

	// Verify registration
	customStates := suite.controller.GetCustomStates()
	suite.Len(customStates, 1, "Should have one custom state registered")
	suite.Equal(handler, customStates[0], "Should return the registered handler")

	// Start the controller (this should trigger DFA setup with custom states)
	err = suite.controller.Start()
	suite.NoError(err, "Controller should start successfully with custom states")

	// Navigate through the states to reach the custom state
	err = suite.controller.Next() // Welcome -> License
	suite.NoError(err, "Should transition to license")

	err = suite.controller.Next() // License -> Components
	suite.NoError(err, "Should transition to components")

	err = suite.controller.Next() // Components -> InstallPath
	suite.NoError(err, "Should transition to install path")

	err = suite.controller.Next() // InstallPath -> DB Config (custom state)
	suite.NoError(err, "Should transition to custom database config state")

	// Verify the custom state was called
	calls := suite.mockView.GetRecordedCalls()
	suite.Contains(calls, "ShowCustomState:"+string(StateDBConfig), "Custom state should have been called")

	// Verify custom state data was passed
	customData, exists := suite.mockView.GetCustomStateData(StateDBConfig)
	suite.True(exists, "Custom state data should exist")
	suite.NotEmpty(customData, "Custom state data should not be empty")

	// Continue to summary
	err = suite.controller.Next() // DB Config -> Summary
	suite.NoError(err, "Should transition from custom state to summary")

	// Verify summary was called
	suite.Contains(calls, "ShowSummary", "Summary should be called after custom state")
}

// TestCustomStateFlow tests the complete flow with custom states
func (suite *CustomStateTestSuite) TestCustomStateFlow() {
	// Register custom handler
	handler := NewDatabaseConfigHandler()
	err := suite.controller.RegisterCustomState(handler)
	suite.NoError(err, "Should register custom state")

	// Set up mock to return specific database configuration
	testConfig := &DatabaseConfig{
		Type:     "postgresql",
		Host:     "testhost",
		Port:     5432,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
		UseSSL:   true,
	}
	suite.mockView.SetReturnData(CustomStateData{"config": testConfig})

	// Run through the complete flow
	err = suite.controller.Start()
	suite.NoError(err, "Should start successfully")

	// Navigate through all states
	states := []string{"license", "components", "install-path", "db-config", "summary"}
	for _, expectedState := range states {
		err = suite.controller.Next()
		suite.NoError(err, "Should transition successfully through state: %s", expectedState)
	}

	// Check that state data contains the database configuration
	stateData := suite.controller.GetStateData()
	suite.NotEmpty(stateData, "State data should not be empty")

	// The database config should be stored in state data
	if dbConfig, exists := stateData["database_config"]; exists {
		config, ok := dbConfig.(*DatabaseConfig)
		suite.True(ok, "Database config should be of correct type")
		suite.Equal("postgresql", config.Type, "Database type should match")
		suite.Equal("testhost", config.Host, "Database host should match")
		suite.Equal(5432, config.Port, "Database port should match")
	}
}

// TestCustomStateValidation tests custom state validation
func (suite *CustomStateTestSuite) TestCustomStateValidation() {
	handler := NewDatabaseConfigHandler()

	// Test various validation scenarios
	testCases := []struct {
		name        string
		config      *DatabaseConfig
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid MySQL config",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "root",
				Password: "password",
				UseSSL:   false,
			},
			shouldError: false,
		},
		{
			name: "valid PostgreSQL config",
			config: &DatabaseConfig{
				Type:     "postgresql",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "postgres",
				Password: "password",
				UseSSL:   true,
			},
			shouldError: false,
		},
		{
			name: "valid SQLite config",
			config: &DatabaseConfig{
				Type:     "sqlite",
				Host:     "", // SQLite doesn't need host
				Port:     0,  // SQLite doesn't need port
				Database: "/path/to/database.db",
				Username: "", // SQLite doesn't require username
				Password: "", // SQLite doesn't require password
				UseSSL:   false,
			},
			shouldError: false,
		},
		{
			name: "empty host",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "",
				Port:     3306,
				Database: "testdb",
			},
			shouldError: true,
		},
		{
			name: "invalid port",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     -1,
				Database: "testdb",
			},
			shouldError: true,
		},
		{
			name: "empty database name",
			config: &DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "",
			},
			shouldError: true,
		},
		{
			name: "unsupported database type",
			config: &DatabaseConfig{
				Type:     "unsupported",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			data := map[string]interface{}{
				"db_config": tc.config,
			}

			err := handler.Validate(suite.controller, data)

			if tc.shouldError {
				suite.Error(err, "Should return validation error for case: %s", tc.name)
			} else {
				suite.NoError(err, "Should pass validation for case: %s", tc.name)
			}
		})
	}
}

// TestCustomStateTestSuite runs the custom state test suite
func TestCustomStateTestSuite(t *testing.T) {
	suite.Run(t, new(CustomStateTestSuite))
}
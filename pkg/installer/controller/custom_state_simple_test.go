package controller

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// Simple test for custom state registration and basic functionality
func TestCustomStateRegistrationSimple(t *testing.T) {
	// Create basic test setup
	tempDir, err := os.MkdirTemp("", "simple_custom_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := &core.Config{
		AppName:       "TestApp",
		Version:       "1.0.0",
		Publisher:     "Test",
		InstallDir:    filepath.Join(tempDir, "install"),
		AcceptLicense: true,
		Components: []core.Component{
			{ID: "core", Name: "Core", Required: true, Selected: true},
		},
	}

	installer := core.New(config)
	controller := NewInstallerController(config, installer)

	// Test 1: Basic registry functionality
	registry := NewCustomStateRegistry()
	handler := NewDatabaseConfigHandler()

	err = registry.Register(handler)
	assert.NoError(t, err, "Should register handler successfully")

	retrieved, exists := registry.GetHandler(StateDBConfig)
	assert.True(t, exists, "Should find registered handler")
	assert.Equal(t, handler, retrieved, "Should return same handler")

	// Test 2: Controller registration
	err = controller.RegisterCustomState(handler)
	assert.NoError(t, err, "Should register custom state with controller")

	customStates := controller.GetCustomStates()
	assert.Len(t, customStates, 1, "Should have one registered custom state")

	// Test 3: Basic database config functionality
	dbConfig := DefaultDatabaseConfig()
	assert.Equal(t, "mysql", dbConfig.Type, "Default should be MySQL")
	assert.Equal(t, "localhost", dbConfig.Host, "Default host should be localhost")
	assert.Equal(t, 3306, dbConfig.Port, "Default port should be 3306")

	// Test 4: Connection string generation
	connectionString := dbConfig.GetConnectionString()
	assert.Contains(t, connectionString, "localhost", "Connection string should contain host")
	assert.Contains(t, connectionString, "3306", "Connection string should contain port")

	// Test 5: PostgreSQL connection string
	pgConfig := &DatabaseConfig{
		Type:     "postgresql",
		Host:     "pghost",
		Port:     5432,
		Database: "pgdb",
		Username: "pguser",
		Password: "pgpass",
		UseSSL:   true,
	}
	pgConnString := pgConfig.GetConnectionString()
	assert.Contains(t, pgConnString, "pghost", "PostgreSQL connection string should contain host")
	assert.Contains(t, pgConnString, "5432", "PostgreSQL connection string should contain port")
	assert.Contains(t, pgConnString, "sslmode=require", "PostgreSQL connection string should require SSL")

	// Test 6: SQLite connection string
	sqliteConfig := &DatabaseConfig{
		Type:     "sqlite",
		Database: "/path/to/db.sqlite",
	}
	sqliteConnString := sqliteConfig.GetConnectionString()
	assert.Equal(t, "/path/to/db.sqlite", sqliteConnString, "SQLite connection string should be the database path")
}

// Simple test for database config validation without network calls
func TestDatabaseConfigValidationSimple(t *testing.T) {
	handler := NewDatabaseConfigHandler()
	controller := &InstallerController{}

	// Test valid MySQL config (no network validation in tests)
	validMySQL := &DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		UseSSL:   false,
	}
	data := map[string]interface{}{"db_config": validMySQL}
	err := handler.Validate(controller, data)
	assert.NoError(t, err, "Valid MySQL config should pass validation")

	// Test valid SQLite config
	validSQLite := &DatabaseConfig{
		Type:     "sqlite",
		Database: "/path/to/database.db",
	}
	data = map[string]interface{}{"db_config": validSQLite}
	err = handler.Validate(controller, data)
	assert.NoError(t, err, "Valid SQLite config should pass validation")

	// Test invalid config - empty database name
	invalidConfig := &DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "", // Empty database name
		Username: "user",
		Password: "pass",
	}
	data = map[string]interface{}{"db_config": invalidConfig}
	err = handler.Validate(controller, data)
	assert.Error(t, err, "Empty database name should fail validation")
	assert.Contains(t, err.Error(), "database name cannot be empty")

	// Test invalid config - unsupported database type
	unsupportedConfig := &DatabaseConfig{
		Type:     "unsupported_db",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
	}
	data = map[string]interface{}{"db_config": unsupportedConfig}
	err = handler.Validate(controller, data)
	assert.Error(t, err, "Unsupported database type should fail validation")
	assert.Contains(t, err.Error(), "unsupported database type")
}

// Simple test for custom state handler properties
func TestCustomStateHandlerProperties(t *testing.T) {
	handler := NewDatabaseConfigHandler()

	// Test state ID
	assert.Equal(t, StateDBConfig, handler.GetStateID(), "Should return correct state ID")

	// Test config
	config := handler.GetConfig()
	assert.Equal(t, "Database Configuration", config.Name, "Should have correct name")
	assert.Equal(t, "Configure database connection settings", config.Description, "Should have correct description")
	assert.True(t, config.CanGoNext, "Should allow going next")
	assert.True(t, config.CanGoBack, "Should allow going back")
	assert.True(t, config.CanCancel, "Should allow canceling")

	// Test insertion point
	insertPoint := handler.GetInsertionPoint()
	assert.Equal(t, StateInstallPath, insertPoint.After, "Should insert after install path")
	assert.Equal(t, StateSummary, insertPoint.Before, "Should insert before summary")

	// Test default database config
	defaultDB := DefaultDatabaseConfig()
	assert.NotNil(t, defaultDB, "Should return default config")
	assert.Equal(t, "mysql", defaultDB.Type, "Default type should be mysql")
	assert.Equal(t, "localhost", defaultDB.Host, "Default host should be localhost")
	assert.Equal(t, 3306, defaultDB.Port, "Default port should be 3306")
	assert.Equal(t, "myapp", defaultDB.Database, "Default database should be myapp")
}
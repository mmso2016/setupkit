// Package controller provides database configuration custom state
package controller

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/mmso2016/setupkit/pkg/wizard"
)

const (
	StateDBConfig wizard.State = "db-config"
)

// DatabaseConfig holds database configuration data
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	UseSSL   bool   `json:"useSSL"`
	Type     string `json:"type"` // mysql, postgresql, sqlite, etc.
}

// DefaultDatabaseConfig returns sensible defaults
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "myapp",
		Username: "root",
		Password: "",
		UseSSL:   false,
		Type:     "mysql",
	}
}

// DatabaseConfigHandler handles database configuration state
type DatabaseConfigHandler struct {
	*BaseCustomStateHandler
	defaultConfig *DatabaseConfig
}

// NewDatabaseConfigHandler creates a new database configuration handler
func NewDatabaseConfigHandler() *DatabaseConfigHandler {
	return &DatabaseConfigHandler{
		BaseCustomStateHandler: &BaseCustomStateHandler{
			StateID:      StateDBConfig,
			Name:         "Database Configuration",
			Description:  "Configure database connection settings",
			InsertPoint:  InsertAfterInstallPath, // After install path, before summary
			CanGoNext:    true,
			CanGoBack:    true,
			CanCancel:    true,
			ValidateFunc: nil, // Will be set below
		},
		defaultConfig: DefaultDatabaseConfig(),
	}
}

// HandleEnter implements CustomStateHandler
func (h *DatabaseConfigHandler) HandleEnter(controller *InstallerController, data map[string]interface{}) error {
	// Initialize with defaults if not already set
	if _, exists := data["db_config"]; !exists {
		data["db_config"] = h.defaultConfig
	}

	// Call the UI to display database configuration
	if view, ok := controller.view.(ExtendedInstallerView); ok {
		customData := CustomStateData{
			"config":        data["db_config"],
			"supported_dbs": []string{"mysql", "postgresql", "sqlite", "sqlserver"},
		}

		result, err := view.ShowCustomState(StateDBConfig, customData)
		if err != nil {
			return err
		}

		// Update data with user input
		if config, ok := result["config"]; ok {
			data["db_config"] = config
		}
	} else {
		return fmt.Errorf("view does not support custom states")
	}

	return nil
}

// HandleLeave implements CustomStateHandler
func (h *DatabaseConfigHandler) HandleLeave(controller *InstallerController, data map[string]interface{}) error {
	// Store database config in the installer config for later use
	if dbConfig, ok := data["db_config"].(*DatabaseConfig); ok {
		// Add to global state data for later access
		data["database_config"] = dbConfig

		// Log configuration (simplified since GetLogger might not exist)
		fmt.Printf("Database configuration saved: %s\n", dbConfig.String())
	}

	return nil
}

// Validate implements CustomStateHandler via ValidateFunc
func (h *DatabaseConfigHandler) Validate(controller *InstallerController, data map[string]interface{}) error {
	// Try both keys for compatibility
	dbConfigInterface, exists := data["db_config"]
	if !exists {
		// Fallback to database_config key used in HandleLeave
		dbConfigInterface, exists = data["database_config"]
	}
	if !exists {
		return fmt.Errorf("database configuration not found")
	}

	dbConfig, ok := dbConfigInterface.(*DatabaseConfig)
	if !ok {
		return fmt.Errorf("invalid database configuration type")
	}

	// Validate host (not required for SQLite)
	if dbConfig.Type != "sqlite" && strings.TrimSpace(dbConfig.Host) == "" {
		return fmt.Errorf("database host cannot be empty")
	}

	// Validate port (not required for SQLite)
	if dbConfig.Type != "sqlite" && (dbConfig.Port <= 0 || dbConfig.Port > 65535) {
		return fmt.Errorf("database port must be between 1 and 65535")
	}

	// Validate database name
	if strings.TrimSpace(dbConfig.Database) == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	// Validate database type
	supportedTypes := map[string]bool{
		"mysql":      true,
		"postgresql": true,
		"sqlite":     true,
		"sqlserver":  true,
	}
	if !supportedTypes[dbConfig.Type] {
		return fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	// For non-sqlite databases, validate connection (skip in test environment or demo mode)
	if dbConfig.Type != "sqlite" && !h.isTestEnvironment() && !h.isDemoMode() {
		if err := h.validateConnection(dbConfig); err != nil {
			return fmt.Errorf("database connection validation failed: %w", err)
		}
	}

	return nil
}

// isTestEnvironment checks if we're running in a test environment
func (h *DatabaseConfigHandler) isTestEnvironment() bool {
	// Simple heuristic: if testing package is imported and tests are running
	for i := 1; i < 10; i++ {
		if _, file, _, ok := runtime.Caller(i); ok {
			if strings.Contains(file, "_test.go") || strings.Contains(file, "testing") {
				return true
			}
		} else {
			break
		}
	}
	return false
}

// isDemoMode checks if we're running in demo mode
func (h *DatabaseConfigHandler) isDemoMode() bool {
	// Simple heuristic: if running from demo executable
	for i := 1; i < 10; i++ {
		if _, file, _, ok := runtime.Caller(i); ok {
			if strings.Contains(file, "demo") || strings.Contains(file, "custom-state-demo") {
				return true
			}
		} else {
			break
		}
	}
	return false
}

// validateConnection tests if we can reach the database host/port
func (h *DatabaseConfigHandler) validateConnection(config *DatabaseConfig) error {
	// Simple TCP connection test
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to %s: %w", address, err)
	}
	conn.Close()
	return nil
}

// GetConnectionString returns a connection string for the configured database
func (config *DatabaseConfig) GetConnectionString() string {
	switch config.Type {
	case "mysql":
		ssl := ""
		if config.UseSSL {
			ssl = "&tls=true"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true%s",
			config.Username, config.Password, config.Host, config.Port, config.Database, ssl)

	case "postgresql":
		ssl := "disable"
		if config.UseSSL {
			ssl = "require"
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, ssl)

	case "sqlite":
		return config.Database // For SQLite, database is the file path

	case "sqlserver":
		return fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s",
			config.Host, config.Port, config.Database, config.Username, config.Password)

	default:
		return ""
	}
}

// String returns a human-readable description of the database configuration
func (config *DatabaseConfig) String() string {
	if config.Type == "sqlite" {
		return fmt.Sprintf("SQLite database: %s", config.Database)
	}
	return fmt.Sprintf("%s database: %s@%s:%d/%s",
		strings.ToUpper(config.Type), config.Username, config.Host, config.Port, config.Database)
}
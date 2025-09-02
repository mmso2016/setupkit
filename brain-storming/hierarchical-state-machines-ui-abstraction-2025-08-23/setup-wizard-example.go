// examples/complete-setup-wizard/main.go
package main

import (
    "bufio"
    "errors"
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"
    "time"
    
    "github.com/setupkit/pkg/wizard"
    "github.com/setupkit/pkg/ui"
    "github.com/setupkit/pkg/ui/controls"
)

// Vollständiger Setup-Wizard mit allen Sub-Flows
type CompleteSetupWizard struct {
    *wizard.SetupWizard
    config *SetupConfig
}

// Konfiguration die während des Setup gesammelt wird
type SetupConfig struct {
    // Welcome Phase
    Language        string
    LicenseAccepted bool
    SystemChecked   bool
    
    // Configuration Phase
    DatabaseConfig  DatabaseConfig
    NetworkConfig   NetworkConfig
    AdminUser       UserConfig
    
    // Installation Phase
    InstallPath     string
    Components      []string
    ServiceConfig   ServiceConfig
    
    // Completion
    TestResults     map[string]bool
    Summary         string
}

type DatabaseConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    Database string
    Engine   string
}

type NetworkConfig struct {
    Port        int
    EnableHTTPS bool
    CertPath    string
    KeyPath     string
}

type UserConfig struct {
    Username string
    Email    string
    Password string
}

type ServiceConfig struct {
    StartWithSystem bool
    ServiceName     string
    LogLevel        string
}

func main() {
    fmt.Println("SetupKit Demo - Hierarchical State Machine with Multi-UI")
    
    // Command line arguments für UI-Mode
    uiMode := "cli" // default
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "--web", "--html":
            uiMode = "web"
        case "--cli":
            uiMode = "cli"
        default:
            fmt.Printf("Unknown UI mode: %s. Using CLI.\n", os.Args[1])
        }
    }
    
    fmt.Printf("Starting setup wizard in %s mode...\n\n", uiMode)
    
    // UI Interface erstellen
    var userInterface ui.UserInterface
    var err error
    
    switch uiMode {
    case "cli":
        userInterface = ui.NewCLIInterface(os.Stdin, os.Stdout)
    case "web":
        userInterface = ui.NewHTMLInterface(":8080")
        fmt.Println("Web interface starting at http://localhost:8080")
        go func() {
            if htmlInterface, ok := userInterface.(*ui.HTMLInterface); ok {
                if err := htmlInterface.Start(); err != nil {
                    fmt.Printf("Web server error: %v\n", err)
                }
            }
        }()
        time.Sleep(1 * time.Second) // Kurz warten bis Server startet
    default:
        fmt.Printf("Unsupported UI mode: %s\n", uiMode)
        return
    }
    
    // Setup-Wizard mit Konfiguration initialisieren
    config := &SetupConfig{}
    setupWizard := &CompleteSetupWizard{
        SetupWizard: wizard.NewSetupWizard(userInterface),
        config:      config,
    }
    
    // Erweiterte Sub-Flows registrieren
    setupWizard.registerCustomFlows()
    
    // Wizard starten
    fmt.Println("Starting setup process...")
    if err = setupWizard.Start(); err != nil {
        fmt.Printf("Setup failed: %v\n", err)
        os.Exit(1)
    }
    
    // Setup abgeschlossen
    fmt.Println("\n🎉 Setup completed successfully!")
    setupWizard.printSummary()
}

// Erweiterte Sub-Flows registrieren
func (w *CompleteSetupWizard) registerCustomFlows() {
    // Diese Methode würde die vollständigen Sub-Flows konfigurieren
    // Hier zeigen wir nur die Struktur
}

func (w *CompleteSetupWizard) printSummary() {
    fmt.Println("\n=== Setup Summary ===")
    fmt.Printf("Language: %s\n", w.config.Language)
    fmt.Printf("Database: %s@%s:%d\n", w.config.DatabaseConfig.Engine, 
        w.config.DatabaseConfig.Host, w.config.DatabaseConfig.Port)
    fmt.Printf("Admin User: %s (%s)\n", w.config.AdminUser.Username, w.config.AdminUser.Email)
    fmt.Printf("Service Port: %d\n", w.config.NetworkConfig.Port)
    fmt.Printf("Install Path: %s\n", w.config.InstallPath)
    fmt.Println("================")
}

// Erweiterte Configuration Flow mit vollständiger Implementierung
type EnhancedConfigurationFlow struct {
    wizard *wizard.SetupWizard
    config *SetupConfig
    states map[string]wizard.WizardState
    current string
}

func NewEnhancedConfigurationFlow(wizard *wizard.SetupWizard, config *SetupConfig) *EnhancedConfigurationFlow {
    cf := &EnhancedConfigurationFlow{
        wizard: wizard,
        config: config,
        states: make(map[string]wizard.WizardState),
    }
    
    // States mit Konfiguration initialisieren
    cf.states["database"] = &DatabaseConfigurationState{flow: cf}
    cf.states["network"] = &NetworkConfigurationState{flow: cf}
    cf.states["admin"] = &AdminUserState{flow: cf}
    cf.states["review"] = &ConfigurationReviewState{flow: cf}
    
    cf.current = "database"
    return cf
}

func (cf *EnhancedConfigurationFlow) GetInitialState() wizard.WizardState {
    return cf.states[cf.current]
}

func (cf *EnhancedConfigurationFlow) GetNextState(current string) wizard.WizardState {
    switch current {
    case "database":
        return cf.states["network"]
    case "network":
        return cf.states["admin"]
    case "admin":
        return cf.states["review"]
    case "review":
        // Übergang zur Installation Phase
        return nil // Würde InstallationFlow zurückgeben
    }
    return nil
}

// DatabaseConfigurationState mit Control-Abstraktion
type DatabaseConfigurationState struct {
    flow *EnhancedConfigurationFlow
}

func (s *DatabaseConfigurationState) Enter() error {
    // Form mit Controls erstellen
    form := ui.NewForm("Database Configuration", 
        "Configure your database connection settings")
    
    // Database Engine Select
    engineOptions := []controls.Option{
        {Value: "postgresql", Label: "PostgreSQL"},
        {Value: "mysql", Label: "MySQL"},
        {Value: "sqlite", Label: "SQLite"},
    }
    
    engineControl := controls.NewSelect("db_engine", "Database Engine", engineOptions).
        SetRequired(true)
    
    // Host Input
    hostControl := controls.NewTextInput("db_host", "Database Host").
        SetPlaceholder("localhost").
        SetRequired(true).
        SetValidator(func(value string) error {
            if value == "" {
                return errors.New("hostname cannot be empty")
            }
            // Validate hostname/IP
            if net.ParseIP(value) == nil && !isValidHostname(value) {
                return errors.New("invalid hostname or IP address")
            }
            return nil
        })
    
    // Port Input
    portControl := controls.NewNumberInput("db_port", "Database Port").
        SetMin(1).
        SetMax(65535).
        SetRequired(true)
    
    // Set default ports based on engine
    if s.flow.config.DatabaseConfig.Engine == "postgresql" {
        portControl.SetValue(5432)
    } else if s.flow.config.DatabaseConfig.Engine == "mysql" {
        portControl.SetValue(3306)
    }
    
    // Username Input
    userControl := controls.NewTextInput("db_user", "Database Username").
        SetRequired(true)
    
    // Password Input
    passControl := controls.NewPasswordInput("db_password", "Database Password").
        SetRequired(true).
        SetMinLength(6)
    
    // Database Name
    dbControl := controls.NewTextInput("db_name", "Database Name").
        SetPlaceholder("setupkit").
        SetRequired(true).
        SetValidator(func(value string) error {
            if !isValidDatabaseName(value) {
                return errors.New("database name contains invalid characters")
            }
            return nil
        })
    
    // Controls zur Form hinzufügen
    form.AddControls([]controls.Control{
        engineControl,
        hostControl, 
        portControl,
        userControl,
        passControl,
        dbControl,
    })
    
    // Test Connection Button
    testButton := controls.NewButton("test_connection", "Test Connection", "testDbConnection()").
        SetVariant("secondary")
    form.AddControl(testButton)
    
    return s.flow.wizard.GetUI().ShowForm(form)
}

func (s *DatabaseConfigurationState) Exit() error {
    return nil
}

func (s *DatabaseConfigurationState) Handle(event wizard.Event) (wizard.WizardState, error) {
    switch event.Type {
    case wizard.NextStep:
        // Input validieren und in Config speichern
        input, ok := event.Data.(ui.Input)
        if !ok {
            return s, errors.New("invalid input data")
        }
        
        // Konfiguration aktualisieren
        s.flow.config.DatabaseConfig = DatabaseConfig{
            Engine:   input["db_engine"].(string),
            Host:     input["db_host"].(string),
            Port:     input["db_port"].(int),
            Username: input["db_user"].(string),
            Password: input["db_password"].(string),
            Database: input["db_name"].(string),
        }
        
        // Verbindung testen
        if err := s.testConnection(); err != nil {
            s.flow.wizard.GetUI().ShowError(fmt.Errorf("database connection failed: %v", err))
            return s, nil
        }
        
        return s.flow.GetNextState("database"), nil
        
    case wizard.PreviousStep:
        // Zurück zur System Check
        return nil, nil // Würde zu vorherigem Flow zurückgehen
        
    case wizard.Cancel:
        return nil, errors.New("setup cancelled")
        
    default:
        return s, nil
    }
}

func (s *DatabaseConfigurationState) GetName() string {
    return "DatabaseConfiguration"
}

func (s *DatabaseConfigurationState) Validate() error {
    cfg := s.flow.config.DatabaseConfig
    if cfg.Host == "" || cfg.Username == "" || cfg.Password == "" {
        return errors.New("all database fields are required")
    }
    return nil
}

func (s *DatabaseConfigurationState) testConnection() error {
    // Hier würde echte Datenbankverbindung getestet
    cfg := s.flow.config.DatabaseConfig
    
    // Simulate connection test
    fmt.Printf("Testing connection to %s@%s:%d...\n", 
        cfg.Engine, cfg.Host, cfg.Port)
    
    time.Sleep(1 * time.Second) // Simulate network delay
    
    // Mock successful connection
    if cfg.Host == "localhost" || cfg.Host == "127.0.0.1" {
        return nil
    }
    
    return fmt.Errorf("could not connect to %s:%d", cfg.Host, cfg.Port)
}

// NetworkConfigurationState
type NetworkConfigurationState struct {
    flow *EnhancedConfigurationFlow
}

func (s *NetworkConfigurationState) Enter() error {
    form := ui.NewForm("Network Configuration", 
        "Configure network settings for your application")
    
    // Application Port
    portControl := controls.NewNumberInput("app_port", "Application Port").
        SetMin(1024).
        SetMax(65535).
        SetRequired(true)
    portControl.SetValue(8080)
    
    // HTTPS Enable
    httpsControl := controls.NewCheckbox("enable_https", "Enable HTTPS").
        SetRequired(false)
    
    // SSL Certificate Path (conditional)
    certControl := controls.NewTextInput("cert_path", "SSL Certificate Path").
        SetPlaceholder("/etc/ssl/certs/app.crt")
    
    // SSL Key Path (conditional) 
    keyControl := controls.NewTextInput("key_path", "SSL Private Key Path").
        SetPlaceholder("/etc/ssl/private/app.key")
    
    form.AddControls([]controls.Control{
        portControl,
        httpsControl,
        certControl,
        keyControl,
    })
    
    return s.flow.wizard.GetUI().ShowForm(form)
}

func (s *NetworkConfigurationState) Exit() error { return nil }

func (s *NetworkConfigurationState) Handle(event wizard.Event) (wizard.WizardState, error) {
    switch event.Type {
    case wizard.NextStep:
        input := event.Data.(ui.Input)
        
        s.flow.config.NetworkConfig = NetworkConfig{
            Port:        input["app_port"].(int),
            EnableHTTPS: input["enable_https"].(bool),
            CertPath:    input["cert_path"].(string),
            KeyPath:     input["key_path"].(string),
        }
        
        return s.flow.GetNextState("network"), nil
        
    case wizard.PreviousStep:
        return s.flow.states["database"], nil
        
    default:
        return s, nil
    }
}

func (s *NetworkConfigurationState) GetName() string {
    return "NetworkConfiguration"
}

func (s *NetworkConfigurationState) Validate() error {
    return nil
}

// AdminUserState
type AdminUserState struct {
    flow *EnhancedConfigurationFlow
}

func (s *AdminUserState) Enter() error {
    form := ui.NewForm("Administrator Account", 
        "Create the main administrator account")
    
    usernameControl := controls.NewTextInput("admin_username", "Username").
        SetRequired(true).
        SetValidator(func(value string) error {
            if len(value) < 3 {
                return errors.New("username must be at least 3 characters")
            }
            if !isValidUsername(value) {
                return errors.New("username contains invalid characters")
            }
            return nil
        })
    
    emailControl := controls.NewTextInput("admin_email", "Email Address").
        SetRequired(true).
        SetValidator(func(value string) error {
            if !strings.Contains(value, "@") {
                return errors.New("invalid email address")
            }
            return nil
        })
    
    passwordControl := controls.NewPasswordInput("admin_password", "Password").
        SetRequired(true).
        SetMinLength(8).
        SetValidator(func(value string) error {
            if len(value) < 8 {
                return errors.New("password must be at least 8 characters")
            }
            return nil
        })
    
    confirmControl := controls.NewPasswordInput("admin_password_confirm", "Confirm Password").
        SetRequired(true)
    
    form.AddControls([]controls.Control{
        usernameControl,
        emailControl,
        passwordControl,
        confirmControl,
    })
    
    return s.flow.wizard.GetUI().ShowForm(form)
}

func (s *AdminUserState) Exit() error { return nil }

func (s *AdminUserState) Handle(event wizard.Event) (wizard.WizardState, error) {
    switch event.Type {
    case wizard.NextStep:
        input := event.Data.(ui.Input)
        
        // Password confirmation check
        if input["admin_password"] != input["admin_password_confirm"] {
            s.flow.wizard.GetUI().ShowError(errors.New("passwords do not match"))
            return s, nil
        }
        
        s.flow.config.AdminUser = UserConfig{
            Username: input["admin_username"].(string),
            Email:    input["admin_email"].(string),
            Password: input["admin_password"].(string),
        }
        
        return s.flow.GetNextState("admin"), nil
        
    case wizard.PreviousStep:
        return s.flow.states["network"], nil
        
    default:
        return s, nil
    }
}

func (s *AdminUserState) GetName() string {
    return "AdminUser"
}

func (s *AdminUserState) Validate() error {
    return nil
}

// ConfigurationReviewState
type ConfigurationReviewState struct {
    flow *EnhancedConfigurationFlow
}

func (s *ConfigurationReviewState) Enter() error {
    form := ui.NewForm("Configuration Review", 
        "Please review your configuration before proceeding")
    
    // Summary als TextArea (read-only)
    summary := s.generateSummary()
    summaryControl := controls.NewTextArea("config_summary", "Configuration Summary").
        SetRows(15)
    summaryControl.SetValue(summary)
    
    // Confirm checkbox
    confirmControl := controls.NewCheckbox("confirm_config", "I confirm this configuration is correct").
        SetRequired(true)
    
    form.AddControls([]controls.Control{
        summaryControl,
        confirmControl,
    })
    
    return s.flow.wizard.GetUI().ShowForm(form)
}

func (s *ConfigurationReviewState) generateSummary() string {
    cfg := s.flow.config
    
    var summary strings.Builder
    summary.WriteString("=== Configuration Summary ===\n\n")
    
    summary.WriteString("Database Configuration:\n")
    summary.WriteString(fmt.Sprintf("  Engine: %s\n", cfg.DatabaseConfig.Engine))
    summary.WriteString(fmt.Sprintf("  Host: %s:%d\n", cfg.DatabaseConfig.Host, cfg.DatabaseConfig.Port))
    summary.WriteString(fmt.Sprintf("  Database: %s\n", cfg.DatabaseConfig.Database))
    summary.WriteString(fmt.Sprintf("  Username: %s\n", cfg.DatabaseConfig.Username))
    summary.WriteString("\n")
    
    summary.WriteString("Network Configuration:\n")
    summary.WriteString(fmt.Sprintf("  Port: %d\n", cfg.NetworkConfig.Port))
    summary.WriteString(fmt.Sprintf("  HTTPS: %v\n", cfg.NetworkConfig.EnableHTTPS))
    if cfg.NetworkConfig.EnableHTTPS {
        summary.WriteString(fmt.Sprintf("  Certificate: %s\n", cfg.NetworkConfig.CertPath))
        summary.WriteString(fmt.Sprintf("  Private Key: %s\n", cfg.NetworkConfig.KeyPath))
    }
    summary.WriteString("\n")
    
    summary.WriteString("Administrator Account:\n")
    summary.WriteString(fmt.Sprintf("  Username: %s\n", cfg.AdminUser.Username))
    summary.WriteString(fmt.Sprintf("  Email: %s\n", cfg.AdminUser.Email))
    
    return summary.String()
}

func (s *ConfigurationReviewState) Exit() error { return nil }

func (s *ConfigurationReviewState) Handle(event wizard.Event) (wizard.WizardState, error) {
    switch event.Type {
    case wizard.NextStep:
        input := event.Data.(ui.Input)
        
        if confirmed, ok := input["confirm_config"].(bool); !ok || !confirmed {
            s.flow.wizard.GetUI().ShowError(errors.New("please confirm the configuration"))
            return s, nil
        }
        
        // Zur Installation Phase übergehen
        return s.flow.GetNextState("review"), nil
        
    case wizard.PreviousStep:
        return s.flow.states["admin"], nil
        
    default:
        return s, nil
    }
}

func (s *ConfigurationReviewState) GetName() string {
    return "ConfigurationReview"
}

func (s *ConfigurationReviewState) Validate() error {
    return nil
}

// Utility-Funktionen
func isValidHostname(hostname string) bool {
    if len(hostname) == 0 || len(hostname) > 253 {
        return false
    }
    
    // Simplified hostname validation
    for _, char := range hostname {
        if !((char >= 'a' && char <= 'z') || 
             (char >= 'A' && char <= 'Z') || 
             (char >= '0' && char <= '9') || 
             char == '-' || char == '.') {
            return false
        }
    }
    
    return true
}

func isValidDatabaseName(name string) bool {
    if len(name) == 0 {
        return false
    }
    
    // Database name should start with letter and contain only alphanumeric and underscore
    for i, char := range name {
        if i == 0 {
            if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
                return false
            }
        } else {
            if !((char >= 'a' && char <= 'z') || 
                 (char >= 'A' && char <= 'Z') || 
                 (char >= '0' && char <= '9') || 
                 char == '_') {
                return false
            }
        }
    }
    
    return true
}

func isValidUsername(username string) bool {
    if len(username) == 0 {
        return false
    }
    
    // Username: alphanumeric, underscore, hyphen
    for _, char := range username {
        if !((char >= 'a' && char <= 'z') || 
             (char >= 'A' && char <= 'Z') || 
             (char >= '0' && char <= '9') || 
             char == '_' || char == '-') {
            return false
        }
    }
    
    return true
}

// CLI-spezifische Input-Behandlung
type CLIInputHandler struct {
    scanner *bufio.Scanner
}

func NewCLIInputHandler() *CLIInputHandler {
    return &CLIInputHandler{
        scanner: bufio.NewScanner(os.Stdin),
    }
}

func (h *CLIInputHandler) GetInput(prompt string) string {
    fmt.Print(prompt)
    if h.scanner.Scan() {
        return strings.TrimSpace(h.scanner.Text())
    }
    return ""
}

func (h *CLIInputHandler) GetIntInput(prompt string, min, max int) int {
    for {
        input := h.GetInput(prompt)
        if value, err := strconv.Atoi(input); err == nil {
            if value >= min && value <= max {
                return value
            }
        }
        fmt.Printf("Please enter a number between %d and %d.\n", min, max)
    }
}

func (h *CLIInputHandler) GetBoolInput(prompt string) bool {
    for {
        input := strings.ToLower(h.GetInput(prompt))
        switch input {
        case "y", "yes", "true", "1":
            return true
        case "n", "no", "false", "0", "":
            return false
        default:
            fmt.Println("Please enter y/n.")
        }
    }
}

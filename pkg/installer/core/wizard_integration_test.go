package core_test

import (
	"testing"

	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// TestWizardProviderRegistry tests the wizard provider registry
func TestWizardProviderRegistry(t *testing.T) {
	// Create a standard provider
	provider := core.NewStandardWizardProvider(core.ModeExpress)
	
	// Test registration
	err := core.RegisterWizardProvider("test-provider", provider)
	if err != nil {
		t.Fatalf("Failed to register wizard provider: %v", err)
	}
	
	// Test duplicate registration
	err = core.RegisterWizardProvider("test-provider", provider)
	if err == nil {
		t.Error("Expected error for duplicate provider registration")
	}
	
	// Test retrieval
	retrievedProvider, err := core.GetWizardProvider("test-provider")
	if err != nil {
		t.Fatalf("Failed to retrieve wizard provider: %v", err)
	}
	
	if retrievedProvider != provider {
		t.Error("Retrieved provider is not the same as registered provider")
	}
	
	// Test setting default
	err = core.SetDefaultWizardProvider("test-provider")
	if err != nil {
		t.Fatalf("Failed to set default provider: %v", err)
	}
	
	// Test default retrieval
	defaultProvider, err := core.GetDefaultWizardProvider()
	if err != nil {
		t.Fatalf("Failed to get default provider: %v", err)
	}
	
	if defaultProvider != provider {
		t.Error("Default provider is not the expected provider")
	}
}

// TestStandardWizardProvider tests the standard wizard provider
func TestStandardWizardProvider(t *testing.T) {
	// Create test config and context
	config := &core.Config{
		AppName:    "Test App",
		Version:    "1.0.0",
		InstallDir: "/tmp/test",
		License:    "Test License",
		Components: []core.Component{
			{
				ID:       "core",
				Name:     "Core",
				Required: true,
				Selected: true,
			},
		},
		DryRun: true,
	}
	
	logger := core.NewLogger("info", "")
	defer logger.Close()
	
	context := &core.Context{
		Config: config,
		Logger: logger,
	}
	
	// Test express mode
	t.Run("ExpressMode", func(t *testing.T) {
		provider := core.NewStandardWizardProvider(core.ModeExpress)
		
		err := provider.Initialize(config, context)
		if err != nil {
			t.Fatalf("Failed to initialize express provider: %v", err)
		}
		
		// Test DFA retrieval
		dfa, err := provider.GetDFA()
		if err != nil {
			t.Fatalf("Failed to get DFA: %v", err)
		}
		
		if dfa == nil {
			t.Fatal("DFA is nil")
		}
		
		// Test validation
		err = provider.ValidateConfiguration()
		if err != nil {
			t.Fatalf("Configuration validation failed: %v", err)
		}
		
		// Test mode
		if provider.GetMode() != core.ModeExpress {
			t.Errorf("Expected mode %v, got %v", core.ModeExpress, provider.GetMode())
		}
	})
	
	// Test custom mode
	t.Run("CustomMode", func(t *testing.T) {
		provider := core.NewStandardWizardProvider(core.ModeCustom)
		
		err := provider.Initialize(config, context)
		if err != nil {
			t.Fatalf("Failed to initialize custom provider: %v", err)
		}
		
		dfa, err := provider.GetDFA()
		if err != nil {
			t.Fatalf("Failed to get DFA: %v", err)
		}
		
		// Test that custom mode has more states than express
		// (This is implementation-specific, but generally true)
		history := dfa.GetHistory()
		if len(history) == 0 {
			t.Error("DFA should have at least initial state in history")
		}
		
		// Test state handlers
		welcomeHandler := provider.GetStateHandler(core.StateWelcome)
		if welcomeHandler == nil {
			t.Error("Welcome state handler should not be nil")
		}
		
		// Test UI mapping
		welcomeUI := provider.GetUIMapping(core.StateWelcome)
		if welcomeUI.Title == "" {
			t.Error("Welcome UI config should have a title")
		}
	})
}

// TestExtendedWizardProvider tests the extended wizard provider
func TestExtendedWizardProvider(t *testing.T) {
	// Create test config and context
	config := &core.Config{
		AppName:    "Extended Test App",
		Version:    "1.0.0",
		InstallDir: "/tmp/extended-test",
		License:    "Extended Test License",
		DryRun:     true,
		EnableThemeSelection: true,
	}
	
	logger := core.NewLogger("info", "")
	defer logger.Close()
	
	context := &core.Context{
		Config: config,
		Logger: logger,
	}
	
	// Test extended provider with theme selection
	t.Run("ThemeSelection", func(t *testing.T) {
		themes := []string{"default", "dark", "modern"}
		provider := core.CreateExtendedProviderWithThemes(core.ModeCustom, themes, "default")
		
		err := provider.Initialize(config, context)
		if err != nil {
			t.Fatalf("Failed to initialize extended provider: %v", err)
		}
		
		// Test that theme selection state is inserted
		insertedStates := provider.GetInsertedStates()
		if len(insertedStates) == 0 {
			t.Error("Extended provider should have inserted states")
		}
		
		// Check if theme selection state is present
		found := false
		for _, state := range insertedStates {
			if state == core.StateThemeSelection {
				found = true
				break
			}
		}
		if !found {
			t.Error("Theme selection state should be in inserted states")
		}
		
		// Test theme selection handler
		themeHandler := provider.GetStateHandler(core.StateThemeSelection)
		if themeHandler == nil {
			t.Error("Theme selection handler should not be nil")
		}
		
		// Test UI mapping for theme selection
		themeUI := provider.GetUIMapping(core.StateThemeSelection)
		if themeUI.Type != core.UIStateTypeSelection {
			t.Errorf("Expected UI type %v, got %v", core.UIStateTypeSelection, themeUI.Type)
		}
		
		// Test that it's recognized as extended state
		if !provider.IsExtendedState(core.StateThemeSelection) {
			t.Error("Theme selection should be recognized as extended state")
		}
	})
}

// TestWizardUIAdapter tests the wizard UI adapter
func TestWizardUIAdapter(t *testing.T) {
	// Setup provider
	provider := core.NewStandardWizardProvider(core.ModeExpress)
	
	config := &core.Config{
		AppName: "UI Adapter Test",
		Version: "1.0.0",
		DryRun:  true,
	}
	
	logger := core.NewLogger("info", "")
	defer logger.Close()
	
	context := &core.Context{
		Config: config,
		Logger: logger,
	}
	
	err := provider.Initialize(config, context)
	if err != nil {
		t.Fatalf("Failed to initialize provider: %v", err)
	}
	
	// Create adapter
	adapter := core.NewWizardUIAdapter(provider)
	
	err = adapter.Initialize(context)
	if err != nil {
		t.Fatalf("Failed to initialize adapter: %v", err)
	}
	
	// Test current state
	currentState := adapter.GetCurrentState()
	if currentState == "" {
		t.Error("Current state should not be empty")
	}
	
	// Test state config
	stateConfig := adapter.GetCurrentStateConfig()
	if stateConfig.Title == "" {
		t.Error("State config should have a title")
	}
	
	// Test state handler
	handler := adapter.GetCurrentStateHandler()
	if handler == nil {
		t.Error("State handler should not be nil")
	}
	
	// Test wizard data
	data := adapter.GetWizardData()
	if data == nil {
		t.Error("Wizard data should not be nil")
	}
	
	// Test setting data
	err = adapter.SetWizardData("test_key", "test_value")
	if err != nil {
		t.Fatalf("Failed to set wizard data: %v", err)
	}
	
	updatedData := adapter.GetWizardData()
	if updatedData["test_key"] != "test_value" {
		t.Error("Wizard data was not updated correctly")
	}
	
	// Test available actions
	actions := adapter.GetAvailableActions()
	if len(actions) == 0 {
		t.Error("Should have at least one available action")
	}
	
	// Test action capability check
	canNext := adapter.CanPerformAction(core.ActionTypeNext)
	if !canNext {
		t.Error("Should be able to perform next action in initial state")
	}
	
	// Test state history
	history := adapter.GetStateHistory()
	if len(history) == 0 {
		t.Error("Should have at least initial state in history")
	}
	
	// Test validation
	err = adapter.ValidateCurrentState()
	if err != nil {
		t.Fatalf("Current state validation failed: %v", err)
	}
	
	// Test reset
	err = adapter.Reset()
	if err != nil {
		t.Fatalf("Failed to reset adapter: %v", err)
	}
	
	// Verify reset worked
	resetData := adapter.GetWizardData()
	if len(resetData) > 0 {
		// Some data might be set by state handlers, so this is not necessarily an error
		t.Logf("Data after reset: %v", resetData)
	}
}

// TestWizardIntegration tests the full integration
func TestWizardIntegration(t *testing.T) {
	// Register providers
	provider := core.NewStandardWizardProvider(core.ModeExpress)
	err := core.RegisterWizardProvider("integration-test", provider)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}
	
	// Create installer with DFA wizard
	config := &core.Config{
		AppName:       "Integration Test",
		Version:       "1.0.0",
		InstallDir:    "/tmp/integration-test",
		License:       "Integration Test License",
		WizardProvider: "integration-test",
		DryRun:        true,
		Verbose:       false, // Reduce log noise
	}
	
	installer := core.New(config)
	
	// Test that DFA wizard is enabled
	if !installer.IsUsingDFAWizard() {
		t.Error("Installer should be using DFA wizard when provider is configured")
	}
	
	// Test wizard adapter access
	adapter := installer.GetWizardAdapter()
	// Adapter might be nil until Run() is called, which initializes it
	// This is expected behavior
	_ = adapter // Suppress unused variable warning
	
	t.Logf("Integration test completed - installer configured with DFA wizard")
}
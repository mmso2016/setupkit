package wizard

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestNewDFA tests DFA creation
func TestNewDFA(t *testing.T) {
	dfa := New()
	
	if dfa == nil {
		t.Fatal("New() returned nil")
	}
	
	if dfa.states == nil {
		t.Error("states map not initialized")
	}
	
	if dfa.data == nil {
		t.Error("data map not initialized")
	}
	
	if dfa.history == nil {
		t.Error("history slice not initialized")
	}
	
	if dfa.maxHistory != 100 {
		t.Errorf("Expected maxHistory 100, got %d", dfa.maxHistory)
	}
	
	if !dfa.strictMode {
		t.Error("Expected strictMode to be true by default")
	}
}

// TestAddState tests adding states to DFA
func TestAddState(t *testing.T) {
	dfa := New()
	
	// Test adding valid state
	err := dfa.AddState("welcome", &StateConfig{
		Name:      "Welcome",
		CanGoNext: true,
		CanCancel: true,
	})
	
	if err != nil {
		t.Errorf("AddState failed: %v", err)
	}
	
	// Test initial state is set automatically
	if dfa.initial != "welcome" {
		t.Error("Initial state not set automatically")
	}
	
	if dfa.current != "welcome" {
		t.Error("Current state not set to initial")
	}
	
	// Test adding duplicate state
	err = dfa.AddState("welcome", &StateConfig{
		Name: "Welcome Duplicate",
	})
	
	if err == nil {
		t.Error("Expected error for duplicate state")
	}
	
	// Add more states
	err = dfa.AddState("license", &StateConfig{
		Name:      "License",
		CanGoBack: true,
		CanGoNext: true,
		CanCancel: true,
	})
	
	if err != nil {
		t.Errorf("Failed to add license state: %v", err)
	}
	
	// Verify initial state didn't change
	if dfa.initial != "welcome" {
		t.Error("Initial state changed unexpectedly")
	}
}

// TestDryRunMode tests dry-run functionality
func TestDryRunMode(t *testing.T) {
	dfa := New()
	dfa.SetDryRun(true)
	
	// Setup wizard states
	dfa.AddState("welcome", &StateConfig{
		Name:      "Welcome",
		CanGoNext: true,
		Transitions: map[Action]State{
			ActionNext: "license",
		},
	})
	
	dfa.AddState("license", &StateConfig{
		Name:      "License Agreement",
		CanGoBack: true,
		CanGoNext: true,
		CanSkip:   true,
		ValidateFunc: func(data map[string]interface{}) error {
			if data["accepted"] != true {
				return fmt.Errorf("license must be accepted")
			}
			return nil
		},
		Transitions: map[Action]State{
			ActionNext: "install",
			ActionSkip: "install",
		},
	})
	
	dfa.AddState("install", &StateConfig{
		Name:      "Installation",
		CanGoBack: true,
		CanGoNext: true,
		Transitions: map[Action]State{
			ActionNext: "complete",
		},
	})
	
	dfa.AddState("complete", &StateConfig{
		Name: "Complete",
	})
	
	// Mark complete as final
	dfa.AddFinalState("complete")
	
	// Perform dry-run navigation
	dfa.SetData("user", "testuser")
	dfa.Next() // to license
	dfa.SetData("accepted", false)
	dfa.Next() // should fail validation in non-dry-run
	dfa.Back() // back to welcome
	dfa.Next() // to license again
	dfa.SetData("accepted", true)
	dfa.Next() // to install
	dfa.Next() // to complete
	
	// Get dry-run log
	log := dfa.GetDryRunLog()
	
	// Verify log contains expected entries
	expectedPatterns := []string{
		"Initial state set to: welcome",
		"Added state: welcome",
		"Added state: license",
		"Added state: install",
		"Added state: complete",
		"State complete marked as final",
		"SetData: user = testuser",
		"Attempting Next from state: welcome",
		"Transition: welcome -> license",
		"SetData: accepted = false",
		"Attempting Next from state: license",
		"Transition: license -> install",
		"Attempting Back from state: install",
		"Attempting Next from state: license",
		"SetData: accepted = true",
		"Attempting Next from state: license",
		"Transition: license -> install",
		"Attempting Next from state: install",
		"Transition: install -> complete",
	}
	
	for _, pattern := range expectedPatterns {
		found := false
		for _, entry := range log {
			if strings.Contains(entry, pattern) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected log pattern not found: %s", pattern)
			t.Logf("Log entries: %v", log)
		}
	}
	
	// Verify we reached final state
	if !dfa.IsInFinalState() {
		t.Error("Should be in final state")
	}
}

// TestComplexWizardDryRun tests a complex installation wizard scenario
func TestComplexWizardDryRun(t *testing.T) {
	dfa := New()
	dfa.SetDryRun(true)
	
	// Track callback invocations
	var callbackLog []string
	
	// Setup complex wizard with callbacks
	dfa.SetCallbacks(&Callbacks{
		OnEnter: func(state State, data map[string]interface{}) error {
			msg := fmt.Sprintf("OnEnter: %s", state)
			callbackLog = append(callbackLog, msg)
			return nil
		},
		OnLeave: func(state State, data map[string]interface{}) error {
			msg := fmt.Sprintf("OnLeave: %s", state)
			callbackLog = append(callbackLog, msg)
			return nil
		},
		OnTransition: func(from, to State, action Action) error {
			msg := fmt.Sprintf("OnTransition: %s -> %s (%s)", from, to, action)
			callbackLog = append(callbackLog, msg)
			return nil
		},
		OnDataChange: func(state State, key string, oldValue, newValue interface{}) error {
			msg := fmt.Sprintf("OnDataChange: %s[%s]: %v -> %v", state, key, oldValue, newValue)
			callbackLog = append(callbackLog, msg)
			return nil
		},
	})
	
	// Create states
	dfa.AddState("welcome", &StateConfig{
		Name:      "Welcome",
		CanGoNext: true,
		CanCancel: true,
		Transitions: map[Action]State{
			ActionNext: "system_check",
		},
	})
	
	dfa.AddState("system_check", &StateConfig{
		Name:      "System Requirements Check",
		CanGoNext: true,
		CanGoBack: true,
		CanCancel: true,
		NextStateFunc: func(data map[string]interface{}) (State, error) {
			if data["system_ok"] == true {
				return "license", nil
			}
			return "system_error", nil
		},
		OnEnter: func(data map[string]interface{}) error {
			// Simulate system check
			data["cpu_cores"] = 8
			data["ram_gb"] = 16
			data["disk_gb"] = 100
			return nil
		},
	})
	
	dfa.AddState("system_error", &StateConfig{
		Name:      "System Requirements Not Met",
		CanGoBack: true,
		CanCancel: true,
		Transitions: map[Action]State{
			ActionRetry: "system_check",
		},
	})
	
	dfa.AddState("license", &StateConfig{
		Name:      "License Agreement",
		CanGoNext: true,
		CanGoBack: true,
		CanCancel: true,
		CanSkip:   false,
		ValidateFunc: func(data map[string]interface{}) error {
			if data["license_accepted"] != true {
				return fmt.Errorf("license must be accepted")
			}
			return nil
		},
		Transitions: map[Action]State{
			ActionNext: "user_type",
		},
	})
	
	dfa.AddState("user_type", &StateConfig{
		Name:      "Select User Type",
		CanGoNext: true,
		CanGoBack: true,
		CanCancel: true,
		NextStateFunc: func(data map[string]interface{}) (State, error) {
			userType := data["user_type"]
			switch userType {
			case "typical":
				return "install_location", nil
			case "custom":
				return "components", nil
			case "portable":
				return "portable_location", nil
			default:
				return "", fmt.Errorf("user type not selected")
			}
		},
	})
	
	dfa.AddState("components", &StateConfig{
		Name:      "Select Components",
		CanGoNext: true,
		CanGoBack: true,
		CanCancel: true,
		ValidateFunc: func(data map[string]interface{}) error {
			components := data["components"]
			if components == nil {
				return fmt.Errorf("no components selected")
			}
			return nil
		},
		Transitions: map[Action]State{
			ActionNext: "install_location",
		},
	})
	
	dfa.AddState("install_location", &StateConfig{
		Name:      "Choose Install Location",
		CanGoNext: true,
		CanGoBack: true,
		CanCancel: true,
		ValidateFunc: func(data map[string]interface{}) error {
			location := data["install_path"]
			if location == nil || location == "" {
				return fmt.Errorf("installation path required")
			}
			return nil
		},
		Transitions: map[Action]State{
			ActionNext: "confirm",
		},
	})
	
	dfa.AddState("portable_location", &StateConfig{
		Name:      "Choose Portable Location",
		CanGoNext: true,
		CanGoBack: true,
		CanCancel: true,
		Transitions: map[Action]State{
			ActionNext: "confirm",
		},
	})
	
	dfa.AddState("confirm", &StateConfig{
		Name:      "Confirm Installation",
		CanGoNext: true,
		CanGoBack: true,
		CanCancel: true,
		Transitions: map[Action]State{
			ActionNext: "installing",
		},
		OnEnter: func(data map[string]interface{}) error {
			// Generate installation summary
			summary := fmt.Sprintf("Type: %v, Path: %v", 
				data["user_type"], data["install_path"])
			data["summary"] = summary
			return nil
		},
	})
	
	dfa.AddState("installing", &StateConfig{
		Name:      "Installing",
		CanCancel: false, // Cannot cancel during installation
		CanGoNext: true,
		Transitions: map[Action]State{
			ActionNext: "complete",
		},
		OnEnter: func(data map[string]interface{}) error {
			data["install_progress"] = 0
			return nil
		},
	})
	
	dfa.AddState("complete", &StateConfig{
		Name: "Installation Complete",
		OnEnter: func(data map[string]interface{}) error {
			data["install_progress"] = 100
			data["completed_at"] = "2024-01-01 12:00:00"
			return nil
		},
	})
	
	// Mark final state
	dfa.AddFinalState("complete")
	
	// Scenario 1: Typical installation with system check failure and retry
	t.Run("TypicalInstallWithRetry", func(t *testing.T) {
		// Start fresh
		dfa.Reset()
		dfa.SetDryRun(true)
		
		// Navigate through wizard
		dfa.SetData("app_name", "TestApp")
		
		// Go to system check
		err := dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to system_check: %v", err)
		}
		
		// System check fails first time
		dfa.SetData("system_ok", false)
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to system_error: %v", err)
		}
		
		if dfa.CurrentState() != "system_error" {
			t.Errorf("Expected system_error, got %s", dfa.CurrentState())
		}
		
		// Retry system check
		dfa.SetData("system_ok", true)
		err = dfa.Transition(ActionRetry)
		if err != nil {
			t.Errorf("Failed to retry: %v", err)
		}
		
		// Now proceed
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to license: %v", err)
		}
		
		// Accept license
		dfa.SetData("license_accepted", true)
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to user_type: %v", err)
		}
		
		// Select typical installation
		dfa.SetData("user_type", "typical")
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to install_location: %v", err)
		}
		
		// Set install path
		dfa.SetData("install_path", "C:\\Program Files\\TestApp")
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to confirm: %v", err)
		}
		
		// Confirm and install
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to installing: %v", err)
		}
		
		// Complete installation
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to complete: %v", err)
		}
		
		// Verify final state
		if !dfa.IsInFinalState() {
			t.Error("Should be in final state")
		}
		
		// Verify data
		data := dfa.GetAllData()
		if data["completed_at"] == nil {
			t.Error("Installation completion time not set")
		}
	})
	
	// Scenario 2: Custom installation with component selection
	t.Run("CustomInstallWithComponents", func(t *testing.T) {
		dfa.Reset()
		dfa.SetDryRun(true)
		
		// Quick navigation to user type
		dfa.SetData("system_ok", true)
		dfa.Next() // to system_check
		dfa.Next() // to license
		dfa.SetData("license_accepted", true)
		dfa.Next() // to user_type
		
		// Select custom installation
		dfa.SetData("user_type", "custom")
		err := dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to components: %v", err)
		}
		
		if dfa.CurrentState() != "components" {
			t.Errorf("Expected components, got %s", dfa.CurrentState())
		}
		
		// Select components
		dfa.SetData("components", []string{"core", "plugins", "docs"})
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed after component selection: %v", err)
		}
		
		// Continue with installation
		dfa.SetData("install_path", "D:\\CustomApp")
		dfa.Next() // to confirm
		dfa.Next() // to installing
		dfa.Next() // to complete
		
		if !dfa.IsInFinalState() {
			t.Error("Should complete custom installation")
		}
	})
	
	// Scenario 3: Navigation with Back button
	t.Run("NavigationWithBack", func(t *testing.T) {
		dfa.Reset()
		dfa.SetDryRun(true)
		
		// Navigate forward
		dfa.SetData("system_ok", true)
		dfa.Next() // to system_check
		dfa.Next() // to license
		dfa.SetData("license_accepted", true)
		dfa.Next() // to user_type
		dfa.SetData("user_type", "typical")
		dfa.Next() // to install_location
		
		// Go back to change user type
		err := dfa.Back()
		if err != nil {
			t.Errorf("Failed to go back: %v", err)
		}
		
		if dfa.CurrentState() != "user_type" {
			t.Errorf("Expected user_type after back, got %s", dfa.CurrentState())
		}
		
		// Change to portable
		dfa.SetData("user_type", "portable")
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to go to portable_location: %v", err)
		}
		
		if dfa.CurrentState() != "portable_location" {
			t.Errorf("Expected portable_location, got %s", dfa.CurrentState())
		}
		
		// Verify history
		history := dfa.GetHistory()
		if len(history) < 2 {
			t.Error("History should contain multiple states")
		}
	})
	
	// Verify dry-run log has entries
	log := dfa.GetDryRunLog()
	if len(log) == 0 {
		t.Error("Dry-run log should not be empty")
	}
	
	// In dry-run mode, callbacks should not be called
	if len(callbackLog) > 0 {
		t.Errorf("Callbacks should not be invoked in dry-run mode, but got: %v", callbackLog)
	}
}

// TestValidationInDryRun tests that validation is skipped in dry-run mode
func TestValidationInDryRun(t *testing.T) {
	dfa := New()
	
	// Add state with strict validation
	dfa.AddState("form", &StateConfig{
		Name:      "Form",
		CanGoNext: true,
		ValidateFunc: func(data map[string]interface{}) error {
			// This would normally fail
			return fmt.Errorf("validation error")
		},
		ValidateOnEntry: func(data map[string]interface{}) error {
			return fmt.Errorf("entry validation error")
		},
		ValidateOnExit: func(data map[string]interface{}) error {
			return fmt.Errorf("exit validation error")
		},
		Transitions: map[Action]State{
			ActionNext: "next",
		},
	})
	
	dfa.AddState("next", &StateConfig{
		Name: "Next",
	})
	
	// In normal mode, transition should fail
	dfa.SetStrictMode(true)
	err := dfa.Next()
	if err == nil {
		t.Error("Expected validation error in normal mode")
	}
	
	// In dry-run mode, transition should succeed
	dfa.SetDryRun(true)
	err = dfa.Next()
	if err != nil {
		t.Errorf("Validation should be skipped in dry-run: %v", err)
	}
	
	if dfa.CurrentState() != "next" {
		t.Error("Should transition despite validation in dry-run")
	}
}

// TestDryRunConcurrency tests dry-run mode with concurrent operations
func TestDryRunConcurrency(t *testing.T) {
	dfa := New()
	dfa.SetDryRun(true)
	
	// Setup states
	for i := 0; i < 10; i++ {
		state := State(fmt.Sprintf("state%d", i))
		nextState := State(fmt.Sprintf("state%d", i+1))
		
		config := &StateConfig{
			Name:      fmt.Sprintf("State %d", i),
			CanGoNext: true,
			CanGoBack: true,
			Transitions: map[Action]State{
				ActionNext: nextState,
			},
		}
		
		dfa.AddState(state, config)
	}
	
	// Add final state
	dfa.AddState("state10", &StateConfig{
		Name: "Final",
	})
	
	// Run concurrent operations
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	
	// Goroutine 1: Navigate forward
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			if err := dfa.Next(); err != nil {
				errors <- err
			}
		}
	}()
	
	// Goroutine 2: Set data
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("key%d", i)
			if err := dfa.SetData(key, i); err != nil {
				errors <- err
			}
		}
	}()
	
	// Goroutine 3: Check state
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			_ = dfa.CurrentState()
			_ = dfa.GetAvailableActions()
			_ = dfa.GetHistory()
		}
	}()
	
	// Wait for completion
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}
	
	// Verify dry-run log exists
	log := dfa.GetDryRunLog()
	if len(log) == 0 {
		t.Error("Dry-run log should contain entries from concurrent operations")
	}
}

// TestDFAClone tests cloning functionality
func TestDFAClone(t *testing.T) {
	// Create and configure original DFA
	original := New()
	original.SetDryRun(true)
	original.SetMaxHistory(50)
	
	// Add states
	original.AddState("state1", &StateConfig{
		Name:      "State 1",
		CanGoNext: true,
		Transitions: map[Action]State{
			ActionNext: "state2",
		},
	})
	
	original.AddState("state2", &StateConfig{
		Name:      "State 2",
		CanGoBack: true,
		CanGoNext: true,
		Transitions: map[Action]State{
			ActionNext: "state3",
		},
	})
	
	original.AddState("state3", &StateConfig{
		Name: "State 3",
	})
	
	// Add final state
	original.AddFinalState("state3")
	
	// Navigate and set data
	original.SetData("key1", "value1")
	original.Next()
	original.SetData("key2", "value2")
	
	// Clone the DFA
	clone := original.Clone()
	
	// Verify clone has same state
	if clone.CurrentState() != original.CurrentState() {
		t.Errorf("Clone state mismatch: %s vs %s", 
			clone.CurrentState(), original.CurrentState())
	}
	
	// Verify clone has same data
	cloneData := clone.GetAllData()
	originalData := original.GetAllData()
	
	if len(cloneData) != len(originalData) {
		t.Error("Clone data size mismatch")
	}
	
	for k, v := range originalData {
		if cloneData[k] != v {
			t.Errorf("Clone data mismatch for key %s", k)
		}
	}
	
	// Verify clone has same history
	if len(clone.GetHistory()) != len(original.GetHistory()) {
		t.Error("Clone history mismatch")
	}
	
	// Verify modifications to clone don't affect original
	clone.SetData("key3", "value3")
	if _, exists := original.GetData("key3"); exists {
		t.Error("Clone modification affected original")
	}
	
	// Verify clone has same configuration
	if clone.maxHistory != original.maxHistory {
		t.Error("Clone configuration mismatch")
	}
	
	if clone.DryRun != original.DryRun {
		t.Error("Clone dry-run mode mismatch")
	}
}

// TestDFAValidate tests DFA validation
func TestDFAValidate(t *testing.T) {
	// Test empty DFA
	t.Run("EmptyDFA", func(t *testing.T) {
		dfa := New()
		err := dfa.Validate()
		if err == nil {
			t.Error("Expected error for empty DFA")
		}
	})
	
	// Test DFA with invalid transition
	t.Run("InvalidTransition", func(t *testing.T) {
		dfa := New()
		dfa.AddState("state1", &StateConfig{
			Name: "State 1",
			Transitions: map[Action]State{
				ActionNext: "nonexistent",
			},
		})
		
		err := dfa.Validate()
		if err == nil {
			t.Error("Expected error for invalid transition")
		}
	})
	
	// Test valid DFA
	t.Run("ValidDFA", func(t *testing.T) {
		dfa := New()
		dfa.AddState("state1", &StateConfig{
			Name: "State 1",
			Transitions: map[Action]State{
				ActionNext: "state2",
			},
		})
		dfa.AddState("state2", &StateConfig{
			Name: "State 2",
		})
		
		err := dfa.Validate()
		if err != nil {
			t.Errorf("Valid DFA validation failed: %v", err)
		}
	})
	
	// Test DFA with invalid global transition
	t.Run("InvalidGlobalTransition", func(t *testing.T) {
		dfa := New()
		dfa.AddState("state1", &StateConfig{Name: "State 1"})
		
		// This should fail
		err := dfa.AddTransition(TransitionRule{
			From:   "state1",
			To:     "nonexistent",
			Action: ActionNext,
		})
		
		if err == nil {
			t.Error("Expected error when adding transition to non-existent state")
		}
	})
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	// Test transition to same state
	t.Run("SelfTransition", func(t *testing.T) {
		dfa := New()
		dfa.AddState("loop", &StateConfig{
			Name: "Loop",
			Transitions: map[Action]State{
				ActionRetry: "loop",
			},
		})
		
		err := dfa.Transition(ActionRetry)
		if err != nil {
			t.Errorf("Self-transition failed: %v", err)
		}
		
		if dfa.CurrentState() != "loop" {
			t.Error("Should remain in same state")
		}
	})
	
	// Test max history limit
	t.Run("MaxHistory", func(t *testing.T) {
		dfa := New()
		dfa.SetMaxHistory(3)
		
		// Create chain of states
		for i := 0; i < 5; i++ {
			state := State(fmt.Sprintf("state%d", i))
			next := State(fmt.Sprintf("state%d", i+1))
			
			dfa.AddState(state, &StateConfig{
				Name:      fmt.Sprintf("State %d", i),
				CanGoNext: true,
				Transitions: map[Action]State{
					ActionNext: next,
				},
			})
		}
		dfa.AddState("state5", &StateConfig{Name: "Final"})
		
		// Navigate through all
		for i := 0; i < 5; i++ {
			dfa.Next()
		}
		
		// Check history is limited
		history := dfa.GetHistory()
		if len(history) > 3 {
			t.Errorf("History exceeds max: %d", len(history))
		}
	})
	
	// Test empty state config
	t.Run("EmptyStateConfig", func(t *testing.T) {
		dfa := New()
		err := dfa.AddState("empty", nil)
		
		if err != nil {
			t.Errorf("Should accept nil config: %v", err)
		}
		
		config, _ := dfa.GetStateConfig("empty")
		if config.Name != "empty" {
			t.Error("Should auto-generate name from state")
		}
	})
	
	// Test callbacks returning errors
	t.Run("CallbackErrors", func(t *testing.T) {
		dfa := New()
		
		errorCount := 0
		dfa.SetCallbacks(&Callbacks{
			OnTransition: func(from, to State, action Action) error {
				errorCount++
				if errorCount > 2 {
					return fmt.Errorf("transition error")
				}
				return nil
			},
		})
		
		dfa.AddState("s1", &StateConfig{
			CanGoNext: true,
			Transitions: map[Action]State{
				ActionNext: "s2",
			},
		})
		dfa.AddState("s2", &StateConfig{Name: "S2"})
		
		// First transition should succeed
		err := dfa.Next()
		if err != nil {
			t.Errorf("First transition should succeed: %v", err)
		}
		
		// Reset for second transition
		dfa.Reset()
		
		// Second transition should succeed
		err = dfa.Next()
		if err != nil {
			t.Errorf("Second transition should succeed: %v", err)
		}
		
		// Reset for third transition
		dfa.Reset()
		
		// Third transition should fail (errorCount > 2)
		err = dfa.Next()
		if err == nil {
			t.Error("Third transition should fail")
		}
	})
}

// TestAvailableActions tests getting available actions
func TestAvailableActions(t *testing.T) {
	dfa := New()
	
	// Create state with mixed capabilities
	dfa.AddState("flexible", &StateConfig{
		Name:      "Flexible",
		CanGoNext: true,
		CanGoBack: false, // No history yet
		CanSkip:   true,
		CanCancel: true,
		Transitions: map[Action]State{
			ActionSave:  "saved",
			ActionRetry: "flexible",
		},
	})
	
	dfa.AddState("saved", &StateConfig{Name: "Saved"})
	
	// Get available actions
	actions := dfa.GetAvailableActions()
	
	// Check expected actions
	expectedActions := map[Action]bool{
		ActionNext:   true,
		ActionSkip:   true,
		ActionCancel: true,
		ActionSave:   true,
		ActionRetry:  true,
	}
	
	for _, action := range actions {
		if !expectedActions[action] {
			t.Errorf("Unexpected action: %s", action)
		}
		delete(expectedActions, action)
	}
	
	// Remaining actions should not be available
	for action := range expectedActions {
		if action == ActionBack {
			continue // Back not available without history
		}
		t.Errorf("Expected action not found: %s", action)
	}
	
	// Add history and test back
	dfa.history = append(dfa.history, "previous")
	actions = dfa.GetAvailableActions()
	
	backFound := false
	for _, action := range actions {
		if action == ActionBack {
			backFound = true
			break
		}
	}
	
	if backFound {
		t.Error("Back should not be available when CanGoBack is false")
	}
}

// Benchmark tests
func BenchmarkDryRunTransition(b *testing.B) {
	dfa := New()
	dfa.SetDryRun(true)
	
	// Setup circular states
	dfa.AddState("state1", &StateConfig{
		Name:      "State 1",
		CanGoNext: true,
		Transitions: map[Action]State{
			ActionNext: "state2",
		},
	})
	
	dfa.AddState("state2", &StateConfig{
		Name:      "State 2",
		CanGoNext: true,
		Transitions: map[Action]State{
			ActionNext: "state1",
		},
	})
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		dfa.Next()
	}
}

func BenchmarkDryRunDataOps(b *testing.B) {
	dfa := New()
	dfa.SetDryRun(true)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%100)
		dfa.SetData(key, i)
		dfa.GetData(key)
	}
}

func BenchmarkClone(b *testing.B) {
	dfa := New()
	
	// Setup complex DFA
	for i := 0; i < 20; i++ {
		state := State(fmt.Sprintf("state%d", i))
		dfa.AddState(state, &StateConfig{
			Name:      fmt.Sprintf("State %d", i),
			CanGoNext: true,
			CanGoBack: true,
		})
		dfa.SetData(fmt.Sprintf("key%d", i), i)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = dfa.Clone()
	}
}

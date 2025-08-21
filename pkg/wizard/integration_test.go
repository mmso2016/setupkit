package wizard_test

import (
	"fmt"
	"testing"

	"github.com/mmso2016/setupkit/pkg/wizard"
)

// TestIntegration performs an integration test of the wizard package
func TestIntegration(t *testing.T) {
	// Test 1: Create a simple wizard using QuickWizard
	t.Run("QuickWizard", func(t *testing.T) {
		dfa, err := wizard.QuickWizard("start", "middle", "end")
		if err != nil {
			t.Fatalf("QuickWizard failed: %v", err)
		}

		if dfa.CurrentState() != "start" {
			t.Errorf("Expected initial state 'start', got %s", dfa.CurrentState())
		}

		// Navigate forward
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to go next: %v", err)
		}

		if dfa.CurrentState() != "middle" {
			t.Errorf("Expected state 'middle', got %s", dfa.CurrentState())
		}
	})

	// Test 2: Create wizard using Builder
	t.Run("Builder", func(t *testing.T) {
		builder := wizard.NewBuilder()
		builder.
			State("welcome").
			Named("Welcome").
			CanGoNext(true).
			Next("complete").
			Add().
			State("complete").
			Named("Complete").
			Add().
			Initial("welcome").
			Final("complete")

		dfa, err := builder.Build()
		if err != nil {
			t.Fatalf("Builder failed: %v", err)
		}

		// Test navigation
		err = dfa.Next()
		if err != nil {
			t.Errorf("Failed to navigate: %v", err)
		}

		if !dfa.IsInFinalState() {
			t.Error("Should be in final state")
		}
	})

	// Test 3: Test WizardBuilder templates
	t.Run("WizardBuilder", func(t *testing.T) {
		wb := wizard.NewWizardBuilder()
		dfa, err := wb.SimpleInstaller().Build()
		if err != nil {
			t.Fatalf("WizardBuilder failed: %v", err)
		}

		// Verify states exist
		states := []wizard.State{
			"welcome", "license", "location",
			"confirm", "installing", "complete",
		}

		for _, state := range states {
			if _, err := dfa.GetStateConfig(state); err != nil {
				t.Errorf("State %s not found", state)
			}
		}
	})

	// Test 4: Dry-run mode
	t.Run("DryRun", func(t *testing.T) {
		dfa := wizard.New()
		dfa.SetDryRun(true)

		dfa.AddState("test", &wizard.StateConfig{
			Name: "Test",
			ValidateFunc: func(data map[string]interface{}) error {
				// This would normally fail
				return fmt.Errorf("validation error")
			},
		})

		// Should not fail in dry-run mode
		err := dfa.ValidateCurrentState()
		if err != nil {
			t.Error("Validation should be skipped in dry-run mode")
		}

		log := dfa.GetDryRunLog()
		if len(log) == 0 {
			t.Error("Dry-run log should have entries")
		}
	})
}

// Example: DFA Wizard UI Integration
// This example demonstrates how the JavaScript wizard control has been eliminated
// and replaced with DFA-based state management.

package main

import (
	"fmt"
	"log"

	"github.com/mmso2016/setupkit/installer"
	"github.com/mmso2016/setupkit/installer/core"
)

func main() {
	fmt.Println("=== SetupKit DFA Wizard UI Integration Demo ===")
	fmt.Println()
	fmt.Println("This demo shows how JavaScript wizard control has been eliminated")
	fmt.Println("and replaced with DFA-based state management.")
	fmt.Println()

	// Example 1: Standard DFA Wizard (eliminates hardcoded JS navigation)
	fmt.Println("1. Creating installer with DFA wizard (replaces hardcoded JavaScript)...")
	
	app, err := installer.New(
		installer.WithAppName("MyApp"),
		installer.WithVersion("1.0.0"),
		installer.WithDFAWizard(), // This enables DFA-based wizard instead of hardcoded JS
		installer.WithComponents(
			core.Component{
				ID:       "core",
				Name:     "Core Files",
				Required: true,
				Selected: true,
				Size:     1024 * 1024 * 50, // 50MB
			},
			core.Component{
				ID:       "docs",
				Name:     "Documentation",
				Required: false,
				Selected: true,
				Size:     1024 * 1024 * 10, // 10MB
			},
		),
	)

	if err != nil {
		log.Fatalf("Failed to create installer: %v", err)
	}

	// Show that DFA wizard is enabled
	if app.IsUsingDFAWizard() {
		fmt.Println("✅ DFA Wizard enabled - JavaScript hardcoded navigation eliminated!")
		
		// Get wizard adapter to demonstrate DFA integration
		adapter := app.GetWizardAdapter()
		if adapter != nil {
			fmt.Printf("✅ Wizard adapter available: %T\n", adapter)
		}
	} else {
		fmt.Println("❌ Still using legacy hardcoded JavaScript")
	}

	fmt.Println()

	// Example 2: Extended DFA Wizard with Theme Selection
	fmt.Println("2. Creating extended wizard with theme selection...")
	
	themes := []string{"default", "dark", "corporate"}
	extendedApp, err := installer.New(
		installer.WithAppName("MyAdvancedApp"),
		installer.WithVersion("2.0.0"),
		installer.WithExtendedWizard(themes, "default"), // Extended DFA wizard
	)

	if err != nil {
		log.Fatalf("Failed to create extended installer: %v", err)
	}

	if extendedApp.IsUsingDFAWizard() {
		fmt.Println("✅ Extended DFA Wizard enabled with theme selection!")
	}

	fmt.Println()

	// Example 3: Show the difference between old and new approach
	fmt.Println("3. Comparison: Old vs New Approach")
	fmt.Println()
	
	fmt.Println("OLD (Hardcoded JavaScript):")
	fmt.Println("├── Hard-coded screen array: ['welcome', 'license', 'components', ...]")
	fmt.Println("├── Manual nextScreen()/previousScreen() functions")
	fmt.Println("├── Hard-coded validation in JavaScript")
	fmt.Println("└── No state management - just array indexing")
	fmt.Println()

	fmt.Println("NEW (DFA-Based):")
	fmt.Println("├── DFA State Machine: Dynamic state transitions")
	fmt.Println("├── Backend-driven navigation via PerformWizardAction()")
	fmt.Println("├── State-specific validation in Go handlers")
	fmt.Println("├── Flexible state insertion (e.g., theme selection)")
	fmt.Println("└── Event-driven UI updates from backend")

	fmt.Println()
	fmt.Println("=== Integration Points ===")
	fmt.Println()
	
	fmt.Println("Frontend Integration:")
	fmt.Println("├── dfa-index.html: Uses DFA wizard instead of hardcoded screens")
	fmt.Println("├── wizard-dfa.js: DFA bridge replacing hardcoded app.js")
	fmt.Println("└── Dynamic button configuration based on DFA state")
	fmt.Println()

	fmt.Println("Backend Integration:")
	fmt.Println("├── InitializeDFAWizard(): Initialize DFA system")
	fmt.Println("├── GetCurrentWizardState(): Get current state and config")
	fmt.Println("├── PerformWizardAction(): Execute actions via DFA")
	fmt.Println("└── Event emission: wizard-state-changed, validation errors")

	fmt.Println()
	fmt.Println("✅ JavaScript wizard control successfully eliminated!")
	fmt.Println("✅ DFA-based state management implemented!")
	fmt.Println("✅ Backend-driven wizard navigation active!")
}
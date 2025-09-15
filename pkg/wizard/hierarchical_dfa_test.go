package wizard

import (
	"testing"
)

// TestHierarchicalDFABasic tests basic hierarchical DFA functionality
func TestHierarchicalDFABasic(t *testing.T) {
	h := NewHierarchical()
	
	if h == nil {
		t.Fatal("NewHierarchical() returned nil")
	}
	
	// Test initial state
	current := h.GetCurrentState()
	if current.Main != "" {
		t.Error("Expected empty initial main state")
	}
	
	// Add license main state with sub-states
	licenseConfig := &MainStateConfig{
		StateConfig: &StateConfig{
			Name:      "License Agreement",
			CanGoNext: true,
			CanGoBack: true,
			CanCancel: true,
		},
		SubStates:               make(map[SubState]*SubStateConfig),
		InitialSubState:         "reading",
		RequireSubStateCompletion: true,
	}
	
	err := h.AddMainState("license", licenseConfig)
	if err != nil {
		t.Fatalf("Failed to add license main state: %v", err)
	}
	
	// Verify initial state was set automatically
	current = h.GetCurrentState()
	if current.Main != "license" {
		t.Errorf("Expected main state 'license', got '%s'", current.Main)
	}
	if current.Sub != "reading" {
		t.Errorf("Expected sub state 'reading', got '%s'", current.Sub)
	}
}

// TestHierarchicalSubStates tests sub-state functionality
func TestHierarchicalSubStates(t *testing.T) {
	h := NewHierarchical()
	h.SetDryRun(true)
	
	// Add license main state
	licenseConfig := &MainStateConfig{
		StateConfig: &StateConfig{
			Name: "License Agreement",
		},
		SubStates:       make(map[SubState]*SubStateConfig),
		InitialSubState: "reading",
	}
	
	err := h.AddMainState("license", licenseConfig)
	if err != nil {
		t.Fatalf("Failed to add main state: %v", err)
	}
	
	// Add sub-states
	readingConfig := &SubStateConfig{
		Name:        "Reading License",
		Description: "User is reading the license text",
		AllowedActions: map[SubAction]bool{
			SubActionScroll: true,
			SubActionInput:  false,
		},
		AutoTransitionTo: "accepting",
	}
	
	err = h.AddSubState("license", "reading", readingConfig)
	if err != nil {
		t.Fatalf("Failed to add reading sub-state: %v", err)
	}
	
	acceptingConfig := &SubStateConfig{
		Name:           "Accepting License",
		Description:    "User can accept or reject license",
		AllowedActions: map[SubAction]bool{
			SubActionSelect: true,
		},
		CanComplete: func(data map[string]interface{}) bool {
			accepted, ok := data["license_accepted"]
			return ok && accepted.(bool)
		},
	}
	
	err = h.AddSubState("license", "accepting", acceptingConfig)
	if err != nil {
		t.Fatalf("Failed to add accepting sub-state: %v", err)
	}
	
	// Test navigation to sub-state
	err = h.NavigateToSubState("accepting")
	if err != nil {
		t.Fatalf("Failed to navigate to accepting sub-state: %v", err)
	}
	
	current := h.GetCurrentState()
	if current.Sub != "accepting" {
		t.Errorf("Expected sub-state 'accepting', got '%s'", current.Sub)
	}
	
	// Test sub-action handling
	err = h.HandleSubAction(SubActionSelect)
	if err != nil {
		t.Fatalf("Failed to handle sub-action: %v", err)
	}
	
	// Test completion condition (should fail without data)
	if h.CanCompleteCurrentSubState() {
		t.Error("Expected sub-state to be incomplete without license_accepted data")
	}
	
	// Set acceptance data
	h.SetData("license_accepted", true)
	
	// Now should be completable
	if !h.CanCompleteCurrentSubState() {
		t.Error("Expected sub-state to be completable after setting license_accepted")
	}
	
	// Verify dry-run log
	log := h.GetDryRunLog()
	if len(log) == 0 {
		t.Error("Expected dry-run log entries")
	}
}

// TestHierarchicalNavigation tests navigation between main states
func TestHierarchicalNavigation(t *testing.T) {
	h := NewHierarchical()
	
	// Add welcome state
	welcomeConfig := &MainStateConfig{
		StateConfig: &StateConfig{Name: "Welcome"},
		SubStates:   make(map[SubState]*SubStateConfig),
	}
	err := h.AddMainState("welcome", welcomeConfig)
	if err != nil {
		t.Fatalf("Failed to add welcome state: %v", err)
	}
	
	// Add components state with sub-states
	componentsConfig := &MainStateConfig{
		StateConfig: &StateConfig{Name: "Component Selection"},
		SubStates:   make(map[SubState]*SubStateConfig),
		InitialSubState: "browsing",
	}
	err = h.AddMainState("components", componentsConfig)
	if err != nil {
		t.Fatalf("Failed to add components state: %v", err)
	}
	
	// Add browsing sub-state
	browsingConfig := &SubStateConfig{
		Name: "Browsing Components",
		AllowedActions: map[SubAction]bool{
			SubActionSelect:   true,
			SubActionDeselect: true,
		},
	}
	err = h.AddSubState("components", "browsing", browsingConfig)
	if err != nil {
		t.Fatalf("Failed to add browsing sub-state: %v", err)
	}
	
	// Test navigation between main states
	err = h.NavigateToMainState("components")
	if err != nil {
		t.Fatalf("Failed to navigate to components: %v", err)
	}
	
	current := h.GetCurrentState()
	if current.Main != "components" {
		t.Errorf("Expected main state 'components', got '%s'", current.Main)
	}
	if current.Sub != "browsing" {
		t.Errorf("Expected initial sub-state 'browsing', got '%s'", current.Sub)
	}
	
	// Test history
	if !h.CanGoBack() {
		t.Error("Expected to be able to go back")
	}
	
	err = h.GoBack()
	if err != nil {
		t.Fatalf("Failed to go back: %v", err)
	}
	
	current = h.GetCurrentState()
	if current.Main != "welcome" {
		t.Errorf("Expected to be back at welcome, got '%s'", current.Main)
	}
}

// TestCompositeStateString tests CompositeState string representations
func TestCompositeStateString(t *testing.T) {
	// Test main state only
	cs1 := CompositeState{Main: "welcome", Sub: ""}
	if cs1.String() != "welcome" {
		t.Errorf("Expected 'welcome', got '%s'", cs1.String())
	}
	
	// Test composite state
	cs2 := CompositeState{Main: "license", Sub: "reading"}
	if cs2.String() != "license.reading" {
		t.Errorf("Expected 'license.reading', got '%s'", cs2.String())
	}
	
	// Test parsing
	parsed := ParseCompositeState("components.selecting")
	if parsed.Main != "components" || parsed.Sub != "selecting" {
		t.Errorf("Failed to parse composite state correctly: %+v", parsed)
	}
	
	// Test parsing main-only state
	parsed2 := ParseCompositeState("welcome")
	if parsed2.Main != "welcome" || parsed2.Sub != "" {
		t.Errorf("Failed to parse main-only state correctly: %+v", parsed2)
	}
}

// TestHierarchicalCallbacks tests callback functionality
func TestHierarchicalCallbacks(t *testing.T) {
	h := NewHierarchical()
	
	callbackLog := []string{}
	
	callbacks := &Callbacks{
		OnTransition: func(from, to State, action Action) error {
			callbackLog = append(callbackLog, string(from)+" -> "+string(to))
			return nil
		},
		OnDataChange: func(state State, key string, oldValue, newValue interface{}) error {
			callbackLog = append(callbackLog, "Data changed: "+key)
			return nil
		},
	}
	
	h.SetCallbacks(callbacks)
	
	// Add states and navigate
	config := &MainStateConfig{
		StateConfig: &StateConfig{Name: "Test"},
		SubStates:   make(map[SubState]*SubStateConfig),
	}
	
	h.AddMainState("test1", config)
	h.AddMainState("test2", config)
	
	h.NavigateToMainState("test2")
	h.SetData("test_key", "test_value")
	
	if len(callbackLog) != 2 {
		t.Errorf("Expected 2 callback entries, got %d: %v", len(callbackLog), callbackLog)
	}
}

// TestHierarchicalStateExplosionPrevention demonstrates the main benefit
func TestHierarchicalStateExplosionPrevention(t *testing.T) {
	h := NewHierarchical()
	
	// Traditional approach would require:
	// license_reading, license_scrolling, license_accepting, license_rejected
	// components_browsing, components_selecting, components_deselecting, components_calculating
	// = 8+ states for just 2 screens
	
	// Hierarchical approach:
	// 2 main states: license, components
	// 2+4 sub-states total = 6 states, but much better organized
	
	// License screen with multiple interactions
	licenseConfig := &MainStateConfig{
		StateConfig:     &StateConfig{Name: "License"},
		SubStates:       make(map[SubState]*SubStateConfig),
		InitialSubState: "reading",
	}
	h.AddMainState("license", licenseConfig)
	
	h.AddSubState("license", "reading", &SubStateConfig{Name: "Reading"})
	h.AddSubState("license", "accepting", &SubStateConfig{Name: "Accepting"})
	
	// Components screen with multiple interactions
	componentsConfig := &MainStateConfig{
		StateConfig:     &StateConfig{Name: "Components"},
		SubStates:       make(map[SubState]*SubStateConfig),
		InitialSubState: "browsing",
	}
	h.AddMainState("components", componentsConfig)
	
	h.AddSubState("components", "browsing", &SubStateConfig{Name: "Browsing"})
	h.AddSubState("components", "selecting", &SubStateConfig{Name: "Selecting"})
	h.AddSubState("components", "calculating", &SubStateConfig{Name: "Calculating Size"})
	h.AddSubState("components", "validating", &SubStateConfig{Name: "Validating Selection"})
	
	// Demonstrate complex navigation without state explosion
	h.NavigateToMainState("license")
	h.NavigateToSubState("accepting")
	
	h.NavigateToMainState("components")
	h.NavigateToSubState("selecting")
	h.NavigateToSubState("calculating")
	h.NavigateToSubState("validating")
	
	// Still only 2 main states, but rich interaction model
	current := h.GetCurrentState()
	if current.Main != "components" {
		t.Errorf("Expected main state 'components', got '%s'", current.Main)
	}
	if current.Sub != "validating" {
		t.Errorf("Expected sub-state 'validating', got '%s'", current.Sub)
	}
	
	// Complex state represented simply
	stateString := current.String()
	if stateString != "components.validating" {
		t.Errorf("Expected 'components.validating', got '%s'", stateString)
	}
	
	t.Logf("✅ Successfully prevented state explosion: 2 main states with rich sub-state interactions")
	t.Logf("Current complex state: %s", stateString)
}
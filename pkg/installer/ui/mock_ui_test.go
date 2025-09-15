package ui_test

import (
	"fmt"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// MockUI implements core.UI interface for testing
type MockUI struct {
	initialized         bool
	responses          map[string]interface{}
	recordedCalls      []string
	licenseAccepted    bool
	selectedComponents []core.Component
	selectedPath       string
	progressCalls      []*core.Progress
	errors             []error
	elevationGranted   bool
	successSummary     *core.InstallSummary
}

// NewMockUI creates a new mock UI for testing
func NewMockUI() *MockUI {
	return &MockUI{
		responses:     make(map[string]interface{}),
		recordedCalls: make([]string, 0),
		progressCalls: make([]*core.Progress, 0),
		errors:        make([]error, 0),
	}
}

// SetLicenseAccepted configures whether license should be accepted
func (m *MockUI) SetLicenseAccepted(accepted bool) {
	m.licenseAccepted = accepted
}

// SetSelectedComponents configures which components should be selected
func (m *MockUI) SetSelectedComponents(components []core.Component) {
	m.selectedComponents = components
}

// SetSelectedPath configures the installation path
func (m *MockUI) SetSelectedPath(path string) {
	m.selectedPath = path
}

// SetElevationGranted configures whether elevation should be granted
func (m *MockUI) SetElevationGranted(granted bool) {
	m.elevationGranted = granted
}

// GetRecordedCalls returns all method calls made to this mock
func (m *MockUI) GetRecordedCalls() []string {
	return m.recordedCalls
}

// GetProgressCalls returns all progress updates received
func (m *MockUI) GetProgressCalls() []*core.Progress {
	return m.progressCalls
}

// GetErrors returns all errors displayed
func (m *MockUI) GetErrors() []error {
	return m.errors
}

// GetSuccessSummary returns the success summary if ShowSuccess was called
func (m *MockUI) GetSuccessSummary() *core.InstallSummary {
	return m.successSummary
}

// Core.UI interface implementation

func (m *MockUI) Initialize(ctx *core.Context) error {
	m.recordedCalls = append(m.recordedCalls, "Initialize")
	m.initialized = true
	return nil
}

func (m *MockUI) Run() error {
	m.recordedCalls = append(m.recordedCalls, "Run")
	return nil
}

func (m *MockUI) Shutdown() error {
	m.recordedCalls = append(m.recordedCalls, "Shutdown")
	return nil
}

func (m *MockUI) ShowWelcome() error {
	m.recordedCalls = append(m.recordedCalls, "ShowWelcome")
	return nil
}

func (m *MockUI) ShowLicense(license string) (bool, error) {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowLicense: %s", license))
	return m.licenseAccepted, nil
}

func (m *MockUI) SelectComponents(components []core.Component) ([]core.Component, error) {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("SelectComponents: %d components", len(components)))
	if m.selectedComponents != nil {
		return m.selectedComponents, nil
	}
	// Default: return required components
	var selected []core.Component
	for _, comp := range components {
		if comp.Required || comp.Selected {
			selected = append(selected, comp)
		}
	}
	return selected, nil
}

func (m *MockUI) SelectInstallPath(defaultPath string) (string, error) {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("SelectInstallPath: %s", defaultPath))
	if m.selectedPath != "" {
		return m.selectedPath, nil
	}
	return defaultPath, nil
}

func (m *MockUI) ShowProgress(progress *core.Progress) error {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowProgress: %s %.1f%%", progress.ComponentName, progress.OverallProgress*100))
	m.progressCalls = append(m.progressCalls, progress)
	return nil
}

func (m *MockUI) ShowError(err error, canRetry bool) (retry bool, errOut error) {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowError: %v (canRetry: %v)", err, canRetry))
	m.errors = append(m.errors, err)
	return false, nil // Never retry in tests
}

func (m *MockUI) ShowSuccess(summary *core.InstallSummary) error {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowSuccess: success=%v path=%s", summary.Success, summary.InstallPath))
	m.successSummary = summary
	return nil
}

func (m *MockUI) RequestElevation(reason string) (bool, error) {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("RequestElevation: %s", reason))
	return m.elevationGranted, nil
}

// MockInstallerView implements the InstallerView interface for DFA testing
type MockInstallerView struct {
	*MockUI
	stateTransitions []StateTransition
}

type StateTransition struct {
	From wizard.State
	To   wizard.State
}

// NewMockInstallerView creates a new mock installer view
func NewMockInstallerView() *MockInstallerView {
	return &MockInstallerView{
		MockUI:           NewMockUI(),
		stateTransitions: make([]StateTransition, 0),
	}
}

// GetStateTransitions returns all state transitions recorded
func (m *MockInstallerView) GetStateTransitions() []StateTransition {
	return m.stateTransitions
}

// InstallerView interface implementation (extends core.UI)

func (m *MockInstallerView) ShowComponents(components []core.Component) (selected []core.Component, err error) {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowComponents: %d components", len(components)))
	if m.selectedComponents != nil {
		return m.selectedComponents, nil
	}
	// Default: return required and selected components
	var result []core.Component
	for _, comp := range components {
		if comp.Required || comp.Selected {
			result = append(result, comp)
		}
	}
	return result, nil
}

func (m *MockInstallerView) ShowInstallPath(defaultPath string) (path string, err error) {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowInstallPath: %s", defaultPath))
	if m.selectedPath != "" {
		return m.selectedPath, nil
	}
	return defaultPath, nil
}

func (m *MockInstallerView) ShowSummary(config *core.Config, selectedComponents []core.Component, installPath string) (proceed bool, err error) {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowSummary: %s v%s -> %s (%d components)", 
		config.AppName, config.Version, installPath, len(selectedComponents)))
	return true, nil // Always proceed in tests
}

func (m *MockInstallerView) ShowComplete(summary *core.InstallSummary) error {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowComplete: success=%v", summary.Success))
	m.successSummary = summary
	return nil
}

func (m *MockInstallerView) ShowErrorMessage(err error) error {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("ShowErrorMessage: %v", err))
	m.errors = append(m.errors, err)
	return nil
}

func (m *MockInstallerView) OnStateChanged(oldState, newState wizard.State) error {
	m.recordedCalls = append(m.recordedCalls, fmt.Sprintf("OnStateChanged: %s -> %s", oldState, newState))
	m.stateTransitions = append(m.stateTransitions, StateTransition{From: oldState, To: newState})
	return nil
}
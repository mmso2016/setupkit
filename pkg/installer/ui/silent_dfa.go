package ui

import (
	"fmt"

	"github.com/mmso2016/setupkit/pkg/installer/controller"
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// SilentUIDFA implements a DFA-controlled non-interactive UI for unattended installations
type SilentUIDFA struct {
	context    *core.Context
	controller *controller.InstallerController
	config     *core.Config
}

// NewSilentUIDFA creates a new DFA-controlled silent UI instance
func NewSilentUIDFA() *SilentUIDFA {
	return &SilentUIDFA{}
}

// Initialize sets up the Silent UI with context and controller
func (s *SilentUIDFA) Initialize(ctx *core.Context) error {
	s.context = ctx
	s.config = ctx.Config

	// Get installer from context
	installer, ok := ctx.Metadata["installer"].(*core.Installer)
	if !ok {
		return fmt.Errorf("no installer found in context")
	}

	// Create DFA controller
	s.controller = controller.NewInstallerController(ctx.Config, installer)
	s.controller.SetView(s)

	s.context.Logger.Info("Starting DFA-controlled silent installation")
	return nil
}

// Run starts the DFA-controlled silent installation flow
func (s *SilentUIDFA) Run() error {
	s.context.Logger.Info("Starting DFA-controlled silent installation flow")
	return s.controller.Start()
}

func (s *SilentUIDFA) Shutdown() error {
	return nil
}

// ============================================================================
// InstallerView Interface Implementation  
// ============================================================================

// ShowWelcome displays welcome information (silent)
func (s *SilentUIDFA) ShowWelcome() error {
	s.context.Logger.Info("Silent installation started",
		"app", s.config.AppName,
		"version", s.config.Version)
	return nil
}

// ShowLicense handles license acceptance (silent)
func (s *SilentUIDFA) ShowLicense(license string) (accepted bool, err error) {
	// In silent mode, check if license was pre-accepted
	if !s.config.AcceptLicense {
		s.context.Logger.Error("License not accepted in silent mode")
		return false, fmt.Errorf("license not accepted in silent mode")
	}
	s.context.Logger.Info("License accepted (silent mode)")
	return true, nil
}

// ShowComponents handles component selection (silent)
func (s *SilentUIDFA) ShowComponents(components []core.Component) (selected []core.Component, err error) {
	// In silent mode, use pre-selected components
	var result []core.Component
	for _, c := range components {
		if c.Selected || c.Required {
			result = append(result, c)
			s.context.Logger.Info("Component selected", "name", c.Name)
		}
	}
	return result, nil
}

// ShowInstallPath handles installation path selection (silent)
func (s *SilentUIDFA) ShowInstallPath(defaultPath string) (path string, err error) {
	// Use configured path or default
	path = s.config.InstallDir
	if path == "" {
		path = defaultPath
	}
	s.context.Logger.Info("Install path", "path", path)
	return path, nil
}

// ShowSummary handles installation summary (silent - auto-proceed)
func (s *SilentUIDFA) ShowSummary(config *core.Config, selectedComponents []core.Component, installPath string) (proceed bool, err error) {
	s.context.Logger.Info("Installation summary",
		"app", config.AppName,
		"version", config.Version,
		"path", installPath,
		"components", len(selectedComponents))
	
	// Log selected components
	for _, comp := range selectedComponents {
		s.context.Logger.Info("Will install component", "name", comp.Name)
	}
	
	// In silent mode, always proceed
	return true, nil
}

// ShowProgress displays installation progress (silent)
func (s *SilentUIDFA) ShowProgress(progress *core.Progress) error {
	s.context.Logger.Info("Progress",
		"component", progress.ComponentName,
		"overall", progress.OverallProgress*100,
		"message", progress.Message)
	return nil
}

// ShowComplete displays installation completion (silent)
func (s *SilentUIDFA) ShowComplete(summary *core.InstallSummary) error {
	s.context.Logger.Info("Installation completed successfully",
		"duration", summary.Duration,
		"path", summary.InstallPath,
		"success", summary.Success)
	
	// Log installed components
	for _, comp := range summary.ComponentsInstalled {
		s.context.Logger.Info("Installed component", "name", comp)
	}
	
	return nil
}

// ShowErrorMessage displays an error message (silent)
func (s *SilentUIDFA) ShowErrorMessage(err error) error {
	s.context.Logger.Error("Installation error", "error", err)
	return nil
}

// OnStateChanged handles state change notifications (silent)
func (s *SilentUIDFA) OnStateChanged(oldState, newState wizard.State) error {
	s.context.Logger.Info("State transition", "from", oldState, "to", newState)
	return nil
}

// ============================================================================
// core.UI Interface Compatibility Methods
// ============================================================================

// SelectComponents - adapts InstallerView.ShowComponents to core.UI interface
func (s *SilentUIDFA) SelectComponents(components []core.Component) ([]core.Component, error) {
	return s.ShowComponents(components)
}

// SelectInstallPath - adapts InstallerView.ShowInstallPath to core.UI interface  
func (s *SilentUIDFA) SelectInstallPath(defaultPath string) (string, error) {
	return s.ShowInstallPath(defaultPath)
}


// ShowSuccess - adapts InstallerView.ShowComplete to core.UI interface
func (s *SilentUIDFA) ShowSuccess(summary *core.InstallSummary) error {
	return s.ShowComplete(summary)
}

// ShowError - core.UI interface method with retry logic
func (s *SilentUIDFA) ShowError(err error, canRetry bool) (retry bool, errOut error) {
	s.context.Logger.Error("Installation error", "error", err, "canRetry", canRetry)
	// In silent mode, never retry
	return false, nil
}

// RequestElevation requests elevated privileges (silent - auto-approve)
func (s *SilentUIDFA) RequestElevation(reason string) (bool, error) {
	// In silent mode, elevation should be handled before starting
	s.context.Logger.Info("Elevation required", "reason", reason)
	return true, nil // Assume elevation is available in silent mode
}

// ============================================================================
// ExtendedInstallerView Interface Implementation
// ============================================================================

// ShowCustomState handles custom states in silent mode
func (s *SilentUIDFA) ShowCustomState(stateID wizard.State, data controller.CustomStateData) (controller.CustomStateData, error) {
	s.context.Logger.Info("Processing custom state in silent mode", "state", stateID)

	// For silent mode, we need to provide default responses based on the state type
	switch stateID {
	case controller.StateDBConfig:
		// Database configuration - use defaults or pre-configured values
		if config, exists := data["config"]; exists {
			s.context.Logger.Info("Using provided database configuration", "config", config)
			return controller.CustomStateData{"db_config": config}, nil
		}

		// Use default database configuration
		defaultDB := controller.DefaultDatabaseConfig()
		s.context.Logger.Info("Using default database configuration",
			"type", defaultDB.Type,
			"host", defaultDB.Host,
			"port", defaultDB.Port)

		return controller.CustomStateData{"db_config": defaultDB}, nil

	default:
		s.context.Logger.Warn("Unknown custom state in silent mode", "state", stateID)
		// Return the data as-is for unknown states
		return data, nil
	}
}
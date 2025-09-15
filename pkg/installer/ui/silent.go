package ui

import (
	"github.com/mmso2016/setupkit/pkg/installer/core"
)

// SilentUI implements a non-interactive UI for unattended installations
type SilentUI struct {
	context *core.Context
	config  *core.Config
}

// NewSilentUI creates a new silent UI instance
func NewSilentUI() *SilentUI {
	return &SilentUI{}
}

func (s *SilentUI) Initialize(ctx *core.Context) error {
	s.context = ctx
	s.config = ctx.Config
	s.context.Logger.Info("Starting silent installation")
	return nil
}

func (s *SilentUI) Run() error {
	// In silent mode, we directly execute the installation
	// Get installer from context and set UI reference
	installer := s.context.Metadata["installer"].(*core.Installer)
	installer.SetUI(s) // Set the UI reference on the installer
	return installer.ExecuteInstallation()
}

func (s *SilentUI) Shutdown() error {
	return nil
}

func (s *SilentUI) ShowWelcome() error {
	s.context.Logger.Info("Silent installation started",
		"app", s.config.AppName,
		"version", s.config.Version)
	return nil
}

func (s *SilentUI) ShowLicense(license string) (bool, error) {
	// In silent mode, check if license was pre-accepted
	if !s.config.AcceptLicense {
		s.context.Logger.Error("License not accepted in silent mode")
		return false, nil
	}
	s.context.Logger.Info("License accepted (silent mode)")
	return true, nil
}

func (s *SilentUI) SelectComponents(components []core.Component) ([]core.Component, error) {
	// In silent mode, use pre-selected components
	var selected []core.Component
	for _, c := range components {
		if c.Selected || c.Required {
			selected = append(selected, c)
			s.context.Logger.Info("Component selected", "name", c.Name)
		}
	}
	return selected, nil
}

func (s *SilentUI) SelectInstallPath(defaultPath string) (string, error) {
	// Use configured path or default
	path := s.config.InstallDir
	if path == "" {
		path = defaultPath
	}
	s.context.Logger.Info("Install path", "path", path)
	return path, nil
}

func (s *SilentUI) ShowProgress(progress *core.Progress) error {
	s.context.Logger.Info("Progress",
		"component", progress.ComponentName,
		"overall", progress.OverallProgress)
	return nil
}

func (s *SilentUI) ShowError(err error, canRetry bool) (bool, error) {
	s.context.Logger.Error("Installation error", "error", err)
	// In silent mode, don't retry
	return false, nil
}

func (s *SilentUI) ShowSuccess(summary *core.InstallSummary) error {
	s.context.Logger.Info("Installation completed successfully",
		"duration", summary.Duration,
		"path", summary.InstallPath)
	return nil
}

func (s *SilentUI) RequestElevation(reason string) (bool, error) {
	// In silent mode, elevation should be handled before starting
	s.context.Logger.Info("Elevation required", "reason", reason)
	return true, nil // Assume elevation is available
}

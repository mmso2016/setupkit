// Package core contains the shared logic and types for the installer framework
package core

import (
	"context"
	"io/fs"
	"time"

	"github.com/mmso2016/setupkit/installer/config"
	"github.com/mmso2016/setupkit/installer/themes"
)

// Mode represents the installation mode
type Mode int

const (
	// ModeAuto automatically detects the best mode (GUI if available, otherwise CLI)
	ModeAuto Mode = iota
	// ModeGUI forces GUI mode
	ModeGUI
	// ModeCLI forces CLI mode
	ModeCLI
	// ModeSilent runs without any user interaction
	ModeSilent
)

// RollbackStrategy defines how to handle failures
type RollbackStrategy int

const (
	// RollbackNone - no automatic rollback
	RollbackNone RollbackStrategy = iota
	// RollbackPartial - rollback only failed component
	RollbackPartial
	// RollbackFull - rollback entire installation
	RollbackFull
)

// ElevationStrategy defines when to request elevated privileges
type ElevationStrategy int

const (
	// ElevationNever - never request elevation
	ElevationNever ElevationStrategy = iota
	// ElevationAuto - request when needed (default)
	ElevationAuto
	// ElevationAlways - always request elevation at start
	ElevationAlways
)

// Component represents an installable component
type Component struct {
	ID          string
	Name        string
	Description string
	Required    bool
	Size        int64
	Selected    bool
	Validator   func() error
	Installer   func(ctx context.Context) error
	Uninstaller func(ctx context.Context) error
}

// PathConfiguration holds PATH-related configuration
type PathConfiguration struct {
	Enabled bool
	System  bool
	Dirs    []string
}

// Config holds the installer configuration
type Config struct {
	// Basic info
	AppName     string
	Version     string
	Publisher   string
	Website     string
	
	// Installation
	Mode             Mode
	InstallDir       string
	Components       []Component
	RequiredSpace    int64 // Required disk space in bytes
	
	// Resources
	Assets       fs.FS
	License      string
	Icon         []byte
	
	// UI Configuration
	UIConfig     *config.UIConfig
	Theme        themes.Theme
	ConfigFile   string
	
	// DFA Wizard Configuration
	WizardProvider   string            // Name of the wizard provider to use
	WizardOptions    map[string]interface{} // Options for the wizard provider
	EnableThemeSelection bool          // Enable theme selection in wizard
	
	// Behavior
	Rollback     RollbackStrategy
	DryRun       bool
	Force        bool
	
	// Unattended
	Unattended   bool
	AcceptLicense bool
	ResponseFile  string
	
	// Logging
	LogFile      string
	LogLevel     string
	Verbose      bool
	
	// PATH management
	PathConfig       *PathConfiguration
	
	// Elevation
	ElevationStrategy ElevationStrategy
	
	// Platform specific
	Platform     PlatformConfig
}

// PlatformConfig holds platform-specific configuration
type PlatformConfig interface {
	Validate() error
	GetDefaults() map[string]interface{}
}

// Context provides context for installation operations
type Context struct {
	Config       *Config
	Logger       Logger
	Progress     ProgressReporter
	StartTime    time.Time
	Checkpoints  []Checkpoint
	Metadata     map[string]interface{}
	UI           UI
}

// Checkpoint represents a rollback point
type Checkpoint struct {
	ID        string
	Timestamp time.Time
	State     map[string]interface{}
	Rollback  func() error
}

// ProgressReporter interface for installation progress
type ProgressReporter interface {
	SetTotal(total int64)
	SetCurrent(current int64)
	SetMessage(message string)
	Done()
	Error(err error)
}

// Logger is defined in the parent installer package to avoid circular dependencies

// UI interface defines the user interface contract
type UI interface {
	// Lifecycle
	Initialize(ctx *Context) error
	Run() error
	Shutdown() error
	
	// User Interaction
	ShowWelcome() error
	ShowLicense(license string) (accepted bool, err error)
	SelectComponents(components []Component) ([]Component, error)
	SelectInstallPath(defaultPath string) (string, error)
	
	// Progress & Status
	ShowProgress(progress *Progress) error
	ShowError(err error, canRetry bool) (retry bool, errOut error)
	ShowSuccess(summary *InstallSummary) error
	
	// Elevation
	RequestElevation(reason string) (bool, error)
}

// Progress represents the installation progress
type Progress struct {
	TotalComponents   int
	CurrentComponent  int
	ComponentName     string
	ComponentProgress float64
	OverallProgress   float64
	Message          string
	IsError          bool
}

// InstallSummary contains the installation summary
type InstallSummary struct {
	Success          bool
	Duration         time.Duration
	ComponentsInstalled []string
	InstallPath      string
	Warnings         []string
	NextSteps        []string
}

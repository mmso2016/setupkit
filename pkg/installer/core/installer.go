package core

import (
	"context"
	"fmt"
	"os"
	"time"
	// Import exitcodes from parent package
	// Note: Adjust import path based on your module name
	// "github.com/mmso2016/setupkit/pkg/installer"
)

// InstallHandler is a function type for custom installation logic
type InstallHandler func(installPath string, components []Component) error

// Installer is the main installer implementation
type Installer struct {
	config   *Config
	context  *Context
	ui       UI
	platform PlatformInstaller
	rollback *RollbackManager
	
	// Custom installation handler
	installHandler InstallHandler
	
	// DFA-based wizard support (optional)
	wizardProvider WizardProvider
	wizardAdapter  *WizardUIAdapter
	useDFAWizard   bool
}

// PlatformInstaller interface for platform-specific operations
type PlatformInstaller interface {
	Initialize() error
	CheckRequirements() error
	IsElevated() bool
	RequiresElevation() bool
	RequestElevation() error
	RegisterWithOS() error
	CreateShortcuts() error
	RegisterUninstaller() error
	UpdatePath(dirs []string, system bool) error

	// PATH management
	AddToPath(dir string, system bool) error
	RemoveFromPath(dir string, system bool) error
	IsInPath(dir string, system bool) bool
}

// UIFactory is a function type for creating UI instances
type UIFactory func(Mode) (UI, error)

// uiFactory holds the registered UI factory function
var uiFactory UIFactory

// RegisterUIFactory registers a UI factory function
func RegisterUIFactory(factory UIFactory) {
	uiFactory = factory
}

// New creates a new installer with the given configuration
func New(config *Config) *Installer {
	installer := &Installer{
		config:       config,
		rollback:     NewRollbackManager(config.Rollback),
		useDFAWizard: false, // Default to legacy mode
	}
	
	// Auto-enable DFA wizard if a wizard provider is set in config
	if config.WizardProvider != "" {
		if err := installer.EnableDFAWizard(config.WizardProvider); err != nil {
			// Log error but don't fail - fall back to legacy mode
			if config.Verbose {
				fmt.Printf("Warning: Failed to enable DFA wizard: %v\n", err)
			}
		}
	}
	
	return installer
}

// EnableDFAWizard enables the DFA-based wizard system
func (i *Installer) EnableDFAWizard(providerName string) error {
	// Get the wizard provider
	provider, err := GetWizardProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get wizard provider %s: %w", providerName, err)
	}
	
	i.wizardProvider = provider
	i.useDFAWizard = true
	
	return nil
}

// EnableExtendedWizardWithThemes enables an extended wizard with theme selection
func (i *Installer) EnableExtendedWizardWithThemes(themes []string, defaultTheme string) error {
	// Create extended provider with theme selection
	mode := ModeCustom // Default to custom mode for extended wizard
	if i.config.Mode == ModeCLI || i.config.Mode == ModeSilent {
		mode = ModeExpress // Fallback to express for non-interactive modes
	}
	
	provider := CreateExtendedProviderWithThemes(InstallMode(mode), themes, defaultTheme)
	
	i.wizardProvider = provider
	i.useDFAWizard = true
	i.config.EnableThemeSelection = true
	
	return nil
}

// GetWizardAdapter returns the wizard UI adapter (if DFA wizard is enabled)
func (i *Installer) GetWizardAdapter() *WizardUIAdapter {
	return i.wizardAdapter
}

// IsUsingDFAWizard returns true if the installer is using the DFA-based wizard
func (i *Installer) IsUsingDFAWizard() bool {
	return i.useDFAWizard
}

// Run executes the installer
func (i *Installer) Run(ctx context.Context) error {
	// Initialize context
	if err := i.initializeContext(ctx); err != nil {
		return fmt.Errorf("failed to initialize context: %w", err)
	}

	// Store installer reference in context for UI to use
	i.context.Metadata["installer"] = i

	// Initialize DFA wizard if enabled
	if i.useDFAWizard && i.wizardProvider != nil {
		adapter := NewWizardUIAdapter(i.wizardProvider)
		if err := adapter.Initialize(i.context); err != nil {
			return fmt.Errorf("failed to initialize wizard adapter: %w", err)
		}
		i.wizardAdapter = adapter
		
		// Store wizard adapter in context for UI to use
		i.context.Metadata["wizard_adapter"] = adapter
		
		i.context.Logger.Info("DFA-based wizard enabled", 
			"provider", fmt.Sprintf("%T", i.wizardProvider))
	}

	// Create and initialize UI based on mode
	if uiFactory == nil {
		return fmt.Errorf("no UI factory registered - ensure UI package is imported")
	}

	ui, err := uiFactory(i.config.Mode)
	if err != nil {
		return fmt.Errorf("failed to create UI: %w", err)
	}
	i.ui = ui
	i.context.UI = ui

	if err := i.ui.Initialize(i.context); err != nil {
		return fmt.Errorf("failed to initialize UI: %w", err)
	}
	defer i.ui.Shutdown()

	// Run the UI (which drives the installation flow)
	return i.ui.Run()
}

// ExecuteInstallation performs the actual installation (called by UI)
func (i *Installer) ExecuteInstallation() error {
	// Pre-checks
	if err := i.preCheck(); err != nil {
		return fmt.Errorf("pre-check failed: %w", err)
	}

	// Check elevation if needed
	if err := i.checkElevation(); err != nil {
		return err
	}

	// Perform installation
	if err := i.performInstallation(); err != nil {
		// Attempt rollback if configured
		if i.config.Rollback != RollbackNone {
			if rollbackErr := i.rollback.Execute(i.context); rollbackErr != nil {
				i.context.Logger.Error("Rollback failed", "error", rollbackErr)
			}
		}
		return err
	}

	// Post-installation tasks
	if err := i.postInstall(); err != nil {
		i.context.Logger.Warn("Post-installation tasks failed", "error", err)
		// Non-fatal, continue
	}

	// Verification
	if err := i.verify(); err != nil {
		i.context.Logger.Warn("Verification failed", "error", err)
		// Non-fatal, continue
	}

	return nil
}

// GetConfig returns the installer configuration
func (i *Installer) GetConfig() *Config {
	return i.config
}

// GetComponents returns the available components
func (i *Installer) GetComponents() []Component {
	return i.config.Components
}

// SetSelectedComponents sets the components to install
func (i *Installer) SetSelectedComponents(components []Component) {
	// Update selected state
	selectedMap := make(map[string]bool)
	for _, c := range components {
		selectedMap[c.ID] = true
	}

	for idx := range i.config.Components {
		i.config.Components[idx].Selected = selectedMap[i.config.Components[idx].ID]
	}
}

// SetInstallPath sets the installation path
func (i *Installer) SetInstallPath(path string) {
	i.config.InstallDir = path
}

// SetInstallHandler sets a custom installation handler
func (i *Installer) SetInstallHandler(handler InstallHandler) {
	i.installHandler = handler
}

// SetUI sets the UI for the installer
func (i *Installer) SetUI(ui UI) {
	i.ui = ui
}

// SetContext sets the context for the installer
func (i *Installer) SetContext(ctx *Context) {
	i.context = ctx
}



// Private methods

func (i *Installer) initializeContext(_ context.Context) error {
	// Set up logging
	logger := NewLogger(i.config.LogLevel, i.config.LogFile)
	if i.config.Verbose {
		logger.SetVerbose(true)
	}

	// Create context
	i.context = &Context{
		Config:      i.config,
		Logger:      logger,
		StartTime:   time.Now(),
		Checkpoints: []Checkpoint{},
		Metadata:    make(map[string]interface{}),
	}

	// Initialize platform
	i.platform = CreatePlatformInstaller(i.config)
	if i.platform != nil {
		if err := i.platform.Initialize(); err != nil {
			return fmt.Errorf("platform initialization failed: %w", err)
		}
	}

	return nil
}

func (i *Installer) preCheck() error {
	// Execute pre-installation callback
	if i.config.BeforeInstall != nil {
		if err := i.config.BeforeInstall(); err != nil {
			return fmt.Errorf("pre-installation callback failed: %w", err)
		}
	}

	// Platform-specific requirements
	if i.platform != nil {
		if err := i.platform.CheckRequirements(); err != nil {
			return err
		}
	}

	// Check disk space
	requiredSpace := i.calculateRequiredSpace()
	if err := CheckDiskSpace(i.config.InstallDir, requiredSpace); err != nil {
		return err
	}

	return nil
}

func (i *Installer) checkElevation() error {
	if i.platform == nil || i.config.DryRun {
		return nil
	}

	shouldElevate := false

	switch i.config.ElevationStrategy {
	case ElevationAlways:
		shouldElevate = !i.platform.IsElevated()
	case ElevationAuto:
		shouldElevate = i.platform.RequiresElevation()
	case ElevationNever:
		shouldElevate = false
	}

	if shouldElevate {
		// Request elevation through UI
		granted, err := i.ui.RequestElevation("Administrative privileges required for installation")
		if err != nil {
			return fmt.Errorf("elevation request failed: %w", err)
		}
		if !granted {
			return fmt.Errorf("elevation denied by user")
		}

		// Actually request elevation from platform
		if err := i.platform.RequestElevation(); err != nil {
			return fmt.Errorf("platform elevation failed: %w", err)
		}
	}

	return nil
}

func (i *Installer) performInstallation() error {
	// Create installation directory
	if err := os.MkdirAll(i.config.InstallDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Calculate total components to install
	componentsToInstall := i.getComponentsToInstall()

	// Create progress tracker
	progress := &Progress{
		TotalComponents: len(componentsToInstall),
	}

	// Install components
	for idx, component := range componentsToInstall {
		progress.CurrentComponent = idx + 1
		progress.ComponentName = component.Name
		progress.ComponentProgress = 0
		progress.OverallProgress = float64(idx) / float64(len(componentsToInstall))
		progress.Message = fmt.Sprintf("Installing %s...", component.Name)

		// Update UI
		if err := i.ui.ShowProgress(progress); err != nil {
			i.context.Logger.Warn("Failed to update progress", "error", err)
		}

		// Validate component
		if component.Validator != nil {
			if err := component.Validator(); err != nil {
				return fmt.Errorf("component validation failed for %s: %w", component.ID, err)
			}
		}

		// Add rollback checkpoint
		i.rollback.AddCheckpoint(component.ID, component.Uninstaller)

		// Install component
		// Create a context with all necessary values for the component
		compCtx := context.WithValue(context.Background(), contextKey("installer_context"), i.context)
		compCtx = context.WithValue(compCtx, contextKey("logger"), i.context.Logger)
		compCtx = context.WithValue(compCtx, contextKey("config"), i.config)
		compCtx = context.WithValue(compCtx, contextKey("platform"), i.platform)
		compCtx = context.WithValue(compCtx, contextKey("assets"), i.config.Assets)

		// Install component using either component-specific installer or custom handler
		var installErr error
		if component.Installer != nil {
			installErr = component.Installer(compCtx)
		} else if i.installHandler != nil {
			// Use custom install handler for single component
			installErr = i.installHandler(i.config.InstallDir, []Component{component})
		}
		
		if installErr != nil {
			progress.IsError = true
			progress.Message = fmt.Sprintf("Failed to install %s", component.Name)
			i.ui.ShowProgress(progress)

			// Ask user if they want to retry
			retry, _ := i.ui.ShowError(installErr, true)
			if !retry {
				return fmt.Errorf("component installation failed for %s: %w", component.ID, installErr)
			}
			// TODO: Implement retry logic
		}

		progress.ComponentProgress = 1.0
		progress.OverallProgress = float64(idx+1) / float64(len(componentsToInstall))
		i.ui.ShowProgress(progress)
	}

	progress.OverallProgress = 1.0
	progress.Message = "Installation complete"
	i.ui.ShowProgress(progress)

	return nil
}

func (i *Installer) postInstall() error {
	if i.platform == nil {
		return nil
	}

	// Register with OS
	if err := i.platform.RegisterWithOS(); err != nil {
		return fmt.Errorf("OS registration failed: %w", err)
	}

	// Update PATH if configured
	if i.config.PathConfig != nil && i.config.PathConfig.Enabled {
		if err := i.platform.UpdatePath(i.config.PathConfig.Dirs, i.config.PathConfig.System); err != nil {
			i.context.Logger.Warn("Failed to update PATH", "error", err)
		}
	}

	// Create shortcuts
	if err := i.platform.CreateShortcuts(); err != nil {
		i.context.Logger.Warn("Failed to create shortcuts", "error", err)
	}

	// Register uninstaller
	if err := i.platform.RegisterUninstaller(); err != nil {
		i.context.Logger.Warn("Failed to register uninstaller", "error", err)
	}

	// Execute post-installation callback
	if i.config.AfterInstall != nil {
		if err := i.config.AfterInstall(); err != nil {
			return fmt.Errorf("post-installation callback failed: %w", err)
		}
	}

	return nil
}

func (i *Installer) verify() error {
	// Basic verification - check if main files exist
	// TODO: Implement verification logic
	return nil
}

func (i *Installer) calculateRequiredSpace() int64 {
	var total int64
	for _, component := range i.config.Components {
		if component.Selected || component.Required {
			total += component.Size
		}
	}
	// Add 20% buffer
	return int64(float64(total) * 1.2)
}

func (i *Installer) getComponentsToInstall() []Component {
	var components []Component
	for _, c := range i.config.Components {
		if c.Selected || c.Required {
			components = append(components, c)
		}
	}
	return components
}

// CreateSummary creates an installation summary
func (i *Installer) CreateSummary() *InstallSummary {
	duration := time.Since(i.context.StartTime)

	var installed []string
	for _, c := range i.getComponentsToInstall() {
		installed = append(installed, c.Name)
	}

	return &InstallSummary{
		Success:             true,
		Duration:            duration,
		ComponentsInstalled: installed,
		InstallPath:         i.config.InstallDir,
		NextSteps: []string{
			fmt.Sprintf("Application installed to: %s", i.config.InstallDir),
			"You can now start using the application",
		},
	}
}

// Helper functions are defined in other files

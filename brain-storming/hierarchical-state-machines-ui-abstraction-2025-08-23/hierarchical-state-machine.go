// pkg/wizard/state_machine.go
package wizard

import (
    "errors"
    "fmt"
)

// Event-Typen für State-Übergänge
type EventType int

const (
    NextStep EventType = iota
    PreviousStep
    Cancel
    Retry
    ValidationError
    Finish
)

// Event mit Daten für State-Übergänge
type Event struct {
    Type EventType
    Data interface{}
}

// WizardState Interface für alle Zustände
type WizardState interface {
    Enter() error
    Exit() error
    Handle(event Event) (WizardState, error)
    GetName() string
    Validate() error
}

// UserInterface Abstraktion für verschiedene UIs
type UserInterface interface {
    ShowStep(step Step) error
    GetUserInput() (Input, error)
    ShowProgress(progress Progress) error
    ShowError(err error) error
}

// Step repräsentiert einen einzelnen Wizard-Schritt
type Step struct {
    Title       string
    Description string
    Fields      []Field
}

type Field struct {
    Name        string
    Type        string
    Label       string
    Required    bool
    Default     interface{}
    Options     []Option
}

type Option struct {
    Value string
    Label string
}

type Input map[string]interface{}
type Progress struct {
    Current int
    Total   int
    Message string
}

// Hauptwizard mit hierarchischen Sub-Wizards
type SetupWizard struct {
    ui           UserInterface
    currentState WizardState
    history      []WizardState
    
    // Sub-State-Machines
    welcomeFlow    *WelcomeFlow
    configFlow     *ConfigurationFlow
    installFlow    *InstallationFlow
    completionFlow *CompletionFlow
}

// NewSetupWizard erstellt einen neuen Setup-Wizard
func NewSetupWizard(ui UserInterface) *SetupWizard {
    wizard := &SetupWizard{
        ui:      ui,
        history: make([]WizardState, 0),
    }
    
    // Sub-Wizards initialisieren
    wizard.welcomeFlow = NewWelcomeFlow(wizard)
    wizard.configFlow = NewConfigurationFlow(wizard)
    wizard.installFlow = NewInstallationFlow(wizard)
    wizard.completionFlow = NewCompletionFlow(wizard)
    
    // Startzustand setzen
    wizard.currentState = wizard.welcomeFlow.GetInitialState()
    
    return wizard
}

// Start startet den Wizard
func (w *SetupWizard) Start() error {
    return w.currentState.Enter()
}

// ProcessEvent verarbeitet Events
func (w *SetupWizard) ProcessEvent(event Event) error {
    nextState, err := w.currentState.Handle(event)
    if err != nil {
        return err
    }
    
    if nextState != w.currentState {
        w.history = append(w.history, w.currentState)
        w.currentState.Exit()
        w.currentState = nextState
        return w.currentState.Enter()
    }
    
    return nil
}

// GoBack geht zum vorherigen State zurück
func (w *SetupWizard) GoBack() error {
    if len(w.history) == 0 {
        return errors.New("no previous state")
    }
    
    w.currentState.Exit()
    
    // Letzten State aus History holen
    prevState := w.history[len(w.history)-1]
    w.history = w.history[:len(w.history)-1]
    
    w.currentState = prevState
    return w.currentState.Enter()
}

// Sub-State-Machine für Welcome-Flow
type WelcomeFlow struct {
    wizard  *SetupWizard
    states  map[string]WizardState
    current string
}

func NewWelcomeFlow(wizard *SetupWizard) *WelcomeFlow {
    wf := &WelcomeFlow{
        wizard: wizard,
        states: make(map[string]WizardState),
    }
    
    // States initialisieren
    wf.states["language"] = &LanguageSelectionState{flow: wf}
    wf.states["license"] = &LicenseAgreementState{flow: wf}
    wf.states["systemcheck"] = &SystemCheckState{flow: wf}
    
    wf.current = "language"
    return wf
}

func (wf *WelcomeFlow) GetInitialState() WizardState {
    return wf.states[wf.current]
}

func (wf *WelcomeFlow) GetNextState(current string) WizardState {
    switch current {
    case "language":
        return wf.states["license"]
    case "license":
        return wf.states["systemcheck"] 
    case "systemcheck":
        // Übergang zur nächsten Sub-State-Machine
        return wf.wizard.configFlow.GetInitialState()
    }
    return nil
}

// Beispiel-State: LanguageSelectionState
type LanguageSelectionState struct {
    flow *WelcomeFlow
}

func (s *LanguageSelectionState) Enter() error {
    step := Step{
        Title:       "Language Selection",
        Description: "Please select your preferred language",
        Fields: []Field{
            {
                Name:     "language",
                Type:     "select",
                Label:    "Language",
                Required: true,
                Options: []Option{
                    {Value: "de", Label: "Deutsch"},
                    {Value: "en", Label: "English"},
                    {Value: "fr", Label: "Français"},
                },
                Default: "en",
            },
        },
    }
    
    return s.flow.wizard.ui.ShowStep(step)
}

func (s *LanguageSelectionState) Exit() error {
    return nil
}

func (s *LanguageSelectionState) Handle(event Event) (WizardState, error) {
    switch event.Type {
    case NextStep:
        if err := s.Validate(); err != nil {
            s.flow.wizard.ui.ShowError(err)
            return s, nil
        }
        return s.flow.GetNextState("language"), nil
        
    case Cancel:
        return nil, errors.New("wizard cancelled")
        
    default:
        return s, nil
    }
}

func (s *LanguageSelectionState) GetName() string {
    return "LanguageSelection"
}

func (s *LanguageSelectionState) Validate() error {
    // Validierung hier implementieren
    return nil
}

// Beispiel-State: LicenseAgreementState  
type LicenseAgreementState struct {
    flow *WelcomeFlow
}

func (s *LicenseAgreementState) Enter() error {
    step := Step{
        Title:       "License Agreement",
        Description: "Please read and accept the license agreement",
        Fields: []Field{
            {
                Name:     "accept_license",
                Type:     "checkbox",
                Label:    "I accept the license agreement",
                Required: true,
            },
        },
    }
    
    return s.flow.wizard.ui.ShowStep(step)
}

func (s *LicenseAgreementState) Exit() error {
    return nil
}

func (s *LicenseAgreementState) Handle(event Event) (WizardState, error) {
    switch event.Type {
    case NextStep:
        if err := s.Validate(); err != nil {
            s.flow.wizard.ui.ShowError(err)
            return s, nil
        }
        return s.flow.GetNextState("license"), nil
        
    case PreviousStep:
        return s.flow.states["language"], nil
        
    case Cancel:
        return nil, errors.New("wizard cancelled")
        
    default:
        return s, nil
    }
}

func (s *LicenseAgreementState) GetName() string {
    return "LicenseAgreement"
}

func (s *LicenseAgreementState) Validate() error {
    // Prüfe ob Lizenz akzeptiert wurde
    return nil
}

// SystemCheckState
type SystemCheckState struct {
    flow *WelcomeFlow
}

func (s *SystemCheckState) Enter() error {
    step := Step{
        Title:       "System Check",
        Description: "Checking system requirements...",
        Fields:      []Field{},
    }
    
    err := s.flow.wizard.ui.ShowStep(step)
    if err != nil {
        return err
    }
    
    // System-Check durchführen
    return s.performSystemCheck()
}

func (s *SystemCheckState) performSystemCheck() error {
    progress := Progress{Current: 0, Total: 3, Message: "Checking operating system..."}
    s.flow.wizard.ui.ShowProgress(progress)
    
    progress = Progress{Current: 1, Total: 3, Message: "Checking available disk space..."}  
    s.flow.wizard.ui.ShowProgress(progress)
    
    progress = Progress{Current: 2, Total: 3, Message: "Checking permissions..."}
    s.flow.wizard.ui.ShowProgress(progress)
    
    progress = Progress{Current: 3, Total: 3, Message: "System check completed"}
    s.flow.wizard.ui.ShowProgress(progress)
    
    return nil
}

func (s *SystemCheckState) Exit() error {
    return nil
}

func (s *SystemCheckState) Handle(event Event) (WizardState, error) {
    switch event.Type {
    case NextStep:
        // Übergang zur Configuration-Phase
        return s.flow.GetNextState("systemcheck"), nil
        
    case PreviousStep:
        return s.flow.states["license"], nil
        
    case Cancel:
        return nil, errors.New("wizard cancelled")
        
    default:
        return s, nil
    }
}

func (s *SystemCheckState) GetName() string {
    return "SystemCheck"
}

func (s *SystemCheckState) Validate() error {
    return nil
}

// Stub für weitere Sub-State-Machines
type ConfigurationFlow struct {
    wizard *SetupWizard
    // Implementierung analog zu WelcomeFlow
}

func NewConfigurationFlow(wizard *SetupWizard) *ConfigurationFlow {
    return &ConfigurationFlow{wizard: wizard}
}

func (cf *ConfigurationFlow) GetInitialState() WizardState {
    // Placeholder - echte Implementierung folgt
    return &DatabaseConfigState{flow: cf}
}

type InstallationFlow struct {
    wizard *SetupWizard
}

func NewInstallationFlow(wizard *SetupWizard) *InstallationFlow {
    return &InstallationFlow{wizard: wizard}
}

func (if_ *InstallationFlow) GetInitialState() WizardState {
    return nil // Placeholder
}

type CompletionFlow struct {
    wizard *SetupWizard
}

func NewCompletionFlow(wizard *SetupWizard) *CompletionFlow {
    return &CompletionFlow{wizard: wizard}
}

func (cf *CompletionFlow) GetInitialState() WizardState {
    return nil // Placeholder
}

// Placeholder für DatabaseConfigState
type DatabaseConfigState struct {
    flow *ConfigurationFlow
}

func (s *DatabaseConfigState) Enter() error { return nil }
func (s *DatabaseConfigState) Exit() error  { return nil }
func (s *DatabaseConfigState) Handle(event Event) (WizardState, error) { return s, nil }
func (s *DatabaseConfigState) GetName() string { return "DatabaseConfig" }
func (s *DatabaseConfigState) Validate() error { return nil }

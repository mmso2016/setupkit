# 🚀 Quick Reference Guide - DFA Wizard Integration

## Schnellstart

### 1. Standard-Wizard verwenden
```go
config := setupkit.Config{
    AppName: "MyApp",
    Version: "1.0.0",
    WizardMode: setupkit.WizardModeCustom,
    Components: []setupkit.Component{...},
}
setupkit.Install(config)
```

### 2. Theme-Selection hinzufügen (Extended Provider)
```go
// In BeforeInstall Hook:
extended := wizard.NewExtendedStandardProvider(wizard.ModeCustom, config)

extended.InsertState(wizard.StateInsertion{
    NewState:   "theme_selection",
    AfterState: wizard.StateLicense,
    Handler:    handlers.NewThemeSelectionHandler(config),
    UIConfig: wizard.UIStateConfig{
        Title:    "Select Theme",
        Template: "theme",
        Type:     wizard.UIStateTypeSelection,
    },
})

wizard.Register("extended", extended)
wizard.SetDefault("extended")
```

### 3. Custom DFA erstellen (für Profis)
```go
builder := wizard.NewFluentDFABuilder().
    WithMode(wizard.ModeUserDefined).
    AddState("start", "Start").
        WithHandler(customHandler).
        TransitionTo("configure", wizard.ActionNext).
        Done().
    AddState("configure", "Configure").
        WithValidator(validateConfig).
        TransitionTo("install", wizard.ActionNext).
        Done().
    SetInitial("start").
    SetFinal("complete")

setupkit.InstallWithDFA(config, builder)
```

### 4. Custom Template registrieren
```go
templateSystem.RegisterCustomTemplate("my_template", `
    <div class="custom-step">
        <h2>{{.Step.Title}}</h2>
        {{range .Step.Fields}}
            <!-- Custom field rendering -->
        {{end}}
    </div>
`)
```

### 5. State Handler mit Template
```go
type MyHandler struct {
    templateName string
}

func (h *MyHandler) OnEnter(ctx context.Context, data map[string]interface{}) error {
    data["_template"] = h.templateName
    data["_template_data"] = map[string]interface{}{
        "custom_field": "value",
    }
    return nil
}

func (h *MyHandler) GetUIConfig() wizard.UIStateConfig {
    return wizard.UIStateConfig{
        Template: "my_template",
        Type:     wizard.UIStateTypeCustom,
    }
}
```

## Wichtige Interfaces

### Provider Interface
```go
type Provider interface {
    GetDFA() (*wizard.DFA, error)
    GetStateHandler(state wizard.State) StateHandler
    GetUIMapping(state wizard.State) UIStateConfig
    ValidateConfiguration() error
    GetMode() InstallMode
}
```

### StateHandler Interface
```go
type StateHandler interface {
    OnEnter(ctx context.Context, data map[string]interface{}) error
    OnExit(ctx context.Context, data map[string]interface{}) error
    Execute(ctx context.Context, data map[string]interface{}) error
    Validate(data map[string]interface{}) error
    GetActions() []StateAction
}
```

## Template-Variablen

Im Template verfügbar:
```go
{{.Wizard.Title}}        // Wizard-Titel
{{.Wizard.CurrentStep}}  // Aktueller Step
{{.Wizard.Steps}}        // Alle Steps für Stepper

{{.Step.Title}}          // Step-Titel
{{.Step.Description}}    // Step-Beschreibung
{{.Step.Fields}}         // Formularfelder
{{.Step.Content}}        // Vorgerenderter Content

{{.Navigation.CanGoBack}}   // Navigation-Flags
{{.Navigation.BackLabel}}   // Button-Labels
{{.Navigation.Actions}}     // Custom Actions

{{.Data}}                // Custom Data vom Handler
```

## Verfügbare Templates

| Template | Verwendung |
|----------|------------|
| `welcome` | Begrüßungsbildschirm |
| `license` | Lizenzvereinbarung |
| `input` | Eingabeformulare |
| `selection` | Auswahl-Listen |
| `theme` | Theme-Auswahl |
| `progress` | Fortschrittsanzeige |
| `summary` | Zusammenfassung |
| `error` | Fehleranzeige |

## Migration Checkliste

- [ ] Provider erstellen/erweitern
- [ ] State Handler implementieren
- [ ] Templates definieren
- [ ] UI-Mappings konfigurieren
- [ ] Feature-Flag aktivieren
- [ ] Tests durchführen
- [ ] Alte Logik entfernen

## Debugging

```go
// Dry-Run Mode aktivieren
dfa.SetDryRun(true)

// State-Historie abrufen
history := dfa.GetHistory()

// Verfügbare Actions prüfen
actions := dfa.GetAvailableActions()

// State-Daten inspizieren
data := dfa.GetAllData()
```

## Performance-Tipps

1. Templates pre-compilen
2. State-Daten cachen
3. Lazy-Loading für große Components
4. Progress-Updates throttlen

---
*Dieses Quick Reference Guide ist Teil der DFA-Wizard Integration Dokumentation*

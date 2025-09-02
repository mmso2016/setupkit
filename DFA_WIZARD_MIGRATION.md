# DFA-Wizard Migration - Entwicklungsstand

## Überblick

Die Migration der SetupKit-Wizard Steuerung von hardcodierter Logik zu einem DFA-basierten System ist **in Entwicklung**. Das bestehende `pkg/wizard` DFA-System bleibt unverändert und wird als Basis verwendet.

## ✅ Implementierte Komponenten

### 1. Provider-Infrastruktur (`installer/core/wizard_provider.go`)
- **WizardProvider Interface** - Definiert die Bridge zwischen DFA und Installer
- **WizardStateHandler Interface** - Definiert State-spezifische Logik
- **UIStateConfig Strukturen** - Konfiguration für UI-Rendering
- **Provider Registry** - Verwaltung verschiedener Wizard-Provider

### 2. Standard-Provider (`installer/core/wizard_standard_provider.go`)
- **StandardWizardProvider** - Implementiert Express/Custom/Advanced Modi
- **Vordefinierte Flows:**
  - **Express**: Welcome → License → Installing → Complete
  - **Custom**: Welcome → License → Components → Location → Ready → Installing → Complete  
  - **Advanced**: Welcome → Mode Select → License → Components → Location → Ready → Installing → Complete

### 3. State Handler (`installer/core/wizard_state_handlers.go`)
- **Vollständige Handler-Implementierungen** für alle Standard-States:
  - WelcomeStateHandler, LicenseStateHandler, ComponentsStateHandler, etc.
- **Validierung und Business Logic** für jeden State
- **UI-Konfiguration** und verfügbare Actions pro State

### 4. Extended Provider (`installer/core/wizard_extended_provider.go`)
- **ExtendedWizardProvider** - Erweitert Standard-Provider
- **Theme-Selection State** - Einfügung zwischen License und Components
- **StateInsertion System** - Flexible Erweiterung des Standard-Flows
- **Theme-Integration** - Automatische Anwendung ausgewählter Themes

### 5. UI-Adapter (`installer/core/wizard_ui_adapter.go`)
- **WizardUIAdapter** - Bridge zwischen DFA und bestehender UI
- **Vollständige DFA-Integration** - State Management, Transitions, Validation
- **Callback System** - OnEnter, OnExit, OnTransition, OnDataChange
- **Action Management** - Mapping von UI-Actions zu DFA-Actions

### 6. Core Integration (`installer/core/installer.go`, `installer/core/config.go`)
- **Installer-Erweiterung** um DFA-Wizard Support
- **Config-Erweiterung** um Wizard-Provider Konfiguration
- **Automatische Provider-Aktivierung** basierend auf Config
- **Backward-Compatibility** - Legacy-Mode bleibt verfügbar

### 7. Public API (`installer/installer.go`)
- **Neue Option-Functions:**
  - `WithDFAWizard()` - Standard Express Wizard
  - `WithCustomDFAWizard()` - Standard Custom Wizard
  - `WithAdvancedDFAWizard()` - Standard Advanced Wizard
  - `WithExtendedWizard(themes, defaultTheme)` - Extended mit Theme-Selection
  - `WithWizardProvider(name)` - Custom Provider
- **Public Methods:**
  - `IsUsingDFAWizard()`, `GetWizardAdapter()`, `EnableDFAWizard()`

### 8. Beispiele und Tests
- **Vollständiges Demo** (`examples/dfa-wizard/main.go`)
- **Comprehensive Tests** (`installer/core/wizard_integration_test.go`)
- **Alle Tests bestehen** ✅

## 🚧 Was noch zu tun ist

### Phase 1: UI-Integration (Kritisch)
- [ ] **UI-Pakete anpassen** (`installer/ui/`) um WizardUIAdapter zu verwenden
- [ ] **Template-System implementieren** für flexible UI-Rendering  
- [ ] **Theme-Selection UI** implementieren (Custom Field Type)
- [ ] **State-spezifische Templates** erstellen

### Phase 2: Provider Registration (Wichtig)
- [ ] **Auto-Registration** der Built-in Provider beim Import
- [ ] **Default Provider Setup** im init() 
- [ ] **Provider Discovery** System
- [ ] **Configuration Validation** erweitern

### Phase 3: Advanced Features (Nice-to-have)
- [ ] **Custom State Insertion** API vereinfachen
- [ ] **Conditional States** (Skip-Logic basierend auf Data)
- [ ] **Multi-Path Flows** (Verzweigungen)
- [ ] **State Persistence** für Wizard-Resume
- [ ] **Rollback zu Previous States** 

### Phase 4: Migration Tools (später)
- [ ] **Legacy-to-DFA Migration Helper**
- [ ] **Provider Generator** für Custom Wizards
- [ ] **Configuration Converter** 
- [ ] **Testing Utilities** für Custom Provider

## 📋 Verwendung (Aktueller Stand)

### Standard DFA Wizard aktivieren:
```go
app, err := installer.New(
    installer.WithAppName("MyApp"),
    installer.WithDFAWizard(), // Express Mode
    // oder
    installer.WithCustomDFAWizard(), // Custom Mode 
    // oder
    installer.WithAdvancedDFAWizard(), // Advanced Mode
)
```

### Extended Wizard mit Theme Selection:
```go
themes := []string{"default", "dark", "corporate"}
app, err := installer.New(
    installer.WithAppName("MyApp"),
    installer.WithExtendedWizard(themes, "default"),
)
```

### Custom Provider registrieren:
```go
// Provider registrieren
core.RegisterWizardProvider("my-provider", myProvider)

// Verwenden
app, err := installer.New(
    installer.WithWizardProvider("my-provider"),
)
```

## ⚠️ Wichtige Hinweise

1. **Nicht Produktiv Ready** - System befindet sich in aktiver Entwicklung
2. **UI Integration fehlt** - Kern-DFA Logik ist implementiert, aber UI-Rendering fehlt noch
3. **Backward Compatibility** - Legacy-System bleibt parallel verfügbar
4. **pkg/wizard unverändert** - Das bestehende DFA-System wird nicht modifiziert
5. **Tests bestehen** - Alle implementierten Komponenten sind getestet

## 🎯 Nächste Prioritäten

1. **UI-Adapter Integration** in `installer/ui/` Pakete
2. **Template-System** für flexible State-Rendering
3. **Provider Auto-Registration** für nahtlose Nutzung
4. **Erweiterte Documentation** und Beispiele

## 🔧 Entwickler-Notizen

- **Alle DFA-Features verfügbar**: Dry-Run, Validation, History, Callbacks
- **Provider-Pattern ermöglicht**: Standard, Extended, Custom Flows
- **State-Insertion funktioniert**: Theme-Selection erfolgreich zwischen License und Components eingefügt
- **Vollständige Tests**: Registry, Provider, Adapter, Integration
- **Clean Architecture**: Klare Trennung zwischen DFA-Logic, Business Logic und UI

**Stand**: Kern-Implementation komplett, UI-Integration und Produktivierung ausstehend.
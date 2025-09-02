# Architektur-Entscheidung: Hierarchische vs. Flache State Machines

## Zusammenfassung

**Entscheidung:** Verwendung hierarchischer State Machines für das setupkit-Projekt

**Status:** Empfohlen ✅  
**Datum:** 23. August 2025  
**Entscheidungsträger:** Entwicklungsteam

## Problemstellung

Bei der Entwicklung des setupkit Setup-Wizards stellt sich die Frage der optimalen State Machine Architektur:

1. **Flacher DFA**: Alle Zustände auf einer Ebene
2. **Hierarchischer DFA**: Eingebettete Sub-State-Machines

Die Entscheidung ist besonders kritisch, da das System sowohl CLI- als auch Web-GUI-Unterstützung benötigt.

## Entscheidungsfaktoren

### Komplexität der GUI-Steuerung
- Setup-Wizards haben natürlich hierarchische Workflows
- Verschiedene Setup-Phasen (Welcome, Config, Installation, Completion)
- Jede Phase hat eigene Sub-Schritte

### Multi-Interface-Anforderungen
- CLI für Automatisierung/Scripting
- Web-GUI für interaktive Benutzer
- Gleiche Geschäftslogik für beide Interfaces

### Wartbarkeit und Skalierbarkeit
- Code-Modularität
- Testbarkeit einzelner Komponenten
- Erweiterbarkeit ohne exponentielles Wachstum

## Optionen

### Option 1: Flacher DFA
```go
WizardStates = {
    "start", "language_select", "license_agree", "system_check",
    "db_host", "db_port", "db_user", "db_pass", "db_test",
    "network_config", "ssl_setup", "user_admin", "user_regular",
    "download_start", "download_progress", "install_files",
    "configure_services", "test_connections", "summary", "finish"
}
```

**Vorteile:**
- ✅ Einfache Implementierung
- ✅ Direkte State-Übergänge
- ✅ Weniger Abstraktionsebenen

**Nachteile:**
- ❌ Zustandsexplosion (20+ States nur für Basic-Setup)
- ❌ Code-Duplikation zwischen ähnlichen States
- ❌ Schwer zu testen und zu erweitern
- ❌ Unnatürliche Struktur für UI-Workflows

### Option 2: Hierarchische State Machines (EMPFOHLEN)
```go
SetupWizard {
    WelcomeFlow {
        LanguageSelection → LicenseAgreement → SystemCheck
    }
    ConfigurationFlow {
        DatabaseSetup → NetworkConfig → UserAccounts
    }
    InstallationFlow {
        DownloadPackages → InstallComponents → ConfigureServices
    }
    CompletionFlow {
        Summary → TestConnections → Finish
    }
}
```

**Vorteile:**
- ✅ Natürliche Workflow-Struktur
- ✅ Modularität: Jeder Sub-Flow testbar
- ✅ Wiederverwendbarkeit von Sub-Flows
- ✅ UI-Interface-agnostische Logik
- ✅ Skalierbar ohne Komplexitätsexplosion

**Nachteile:**
- ❌ Komplexere Implementierung
- ❌ Zusätzliche Abstraktionsebene

## Entscheidung

**Gewählt: Option 2 - Hierarchische State Machines**

### Begründung

1. **Natürliche Struktur**: Setup-Prozesse haben inhärent hierarchische Phasen
2. **UI-Abstraktion**: Ermöglicht elegante Trennung zwischen CLI und Web-Interface
3. **Modularität**: Jeder Sub-Flow kann unabhängig entwickelt und getestet werden
4. **Wiederverwendbarkeit**: Sub-Flows können in verschiedenen Setup-Modi wiederverwendet werden
5. **Skalierbarkeit**: Neue Features durch neue Sub-Flows, nicht exponentielles Wachstum

### Implementierungsansatz

```go
// Interface-Abstraktion
type UserInterface interface {
    ShowStep(step Step) error
    GetUserInput() (Input, error)
    ShowProgress(progress Progress) error
    ShowError(err error) error
}

// Hierarchische State Machine
type SetupWizard struct {
    ui UserInterface
    currentState WizardState
    welcomeFlow *WelcomeFlow      // Sub-DFA
    configFlow  *ConfigurationFlow // Sub-DFA
    installFlow *InstallationFlow  // Sub-DFA
    completionFlow *CompletionFlow // Sub-DFA
}

// CLI Implementation
type CLIInterface struct { ... }

// Web Implementation  
type HTMLInterface struct { ... }
```

## Konsequenzen

### Positive Auswirkungen
- **Wartbarkeit**: Klare Trennung der Concerns
- **Testbarkeit**: Jeder Sub-Flow isoliert testbar
- **Multi-Interface**: Ein Core, mehrere UIs
- **Erweiterbarkeit**: Neue Setup-Modi durch Kombination von Sub-Flows

### Negative Auswirkungen
- **Komplexität**: Mehr Abstraktion bedeutet längere Einarbeitungszeit
- **Performance**: Minimaler Overhead durch zusätzliche Ebenen

### Risiken und Mitigierung
- **Risiko**: Über-Engineering
- **Mitigation**: Start mit einfachen Sub-Flows, iterative Erweiterung

## Referenzen
- State Machine Pattern (GoF Design Patterns)
- Hierarchical State Machines (UML Statecharts)
- Command Pattern für UI-Abstraktion

## Nächste Schritte
1. Implementierung der Basis-State-Machine-Interfaces
2. Entwicklung des ersten Sub-Flows (WelcomeFlow)
3. CLI- und HTML-Interface-Implementierung
4. Testing-Framework für verschiedene UI-Modi
5. Dokumentation der State-Übergänge

## Changelog
- **2025-08-23**: Initiale Entscheidung für hierarchische State Machines
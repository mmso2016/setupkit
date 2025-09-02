# SetupKit Framework Vision - Fortsetzung
## Chat-Archiv vom 23. August 2025 (Teil 2)

### Zusammenfassung der Fortsetzungs-Diskussion

Diese Diskussion erweiterte die anfängliche Architektur-Entscheidung um die **strategische Vision** für SetupKit als Enterprise-Framework und konkrete nächste Schritte für das Kobalt-GK-Projekt.

## 🎯 SetupKit Framework Vision

### Framework-Charakteristik (Erkannt)

**SetupKit = Framework für Workflows/Wizards mit Multi-Interface-Support**

- ✅ Nicht nur Setup-Tool, sondern allgemeines Framework für geführte Workflows
- ✅ GUI und/oder CLI Oberfläche
- ✅ Komplexe Installationsprogramme mit Embedded Objects
- ✅ Wiederkehrende Abläufe als Programme zur wiederholten Nutzung
- ✅ Platform-agnostisch: Kann beliebige Programme installieren (nicht nur Go)

## 💾 Hybrid-Artefakt-Architektur

### Statische vs. Dynamische Artefakte

**Statische Artefakte (Installation):**
- Embedded Objects in .exe (go:embed)
- Zur Compile-Zeit festgelegt
- Binaries, Configs, Scripts

**Dynamische Workflow-Daten (Runtime):**
- SQLite-Template embedded in .exe
- Bei Start: Kopie ins Dateisystem als Runtime-DB
- Jede Workflow-Instanz = eigene SQLite-Datei
- User-Aktionen und State werden persistent

### Vorteile der SQLite-Runtime-Lösung

- 🔄 **Workflow-Isolation**: Jede Instanz hat eigenen State
- ⏸️ **Pausieren/Fortsetzen**: Unterbrechung und Wiederaufnahme möglich
- 📊 **Audit-Trail**: Komplette Geschichte aller User-Aktionen
- 🔙 **Rollback**: Rückkehr zu vorherigen Zuständen
- 👥 **Multi-User**: Parallele Workflows verschiedener User
- 💡 **Debugging**: Runtime-State inspizierbar
- 🔄 **DB-Migration**: SQLite → PostgreSQL möglich

## 🏗️ Code-Generator-Vision

**Enterprise-Szenario:**
```
Requirement Definition → SetupKit Generator → Custom Installer.exe
```

**Distribution Pipeline:**
- Developer definiert Requirements
- CI/CD generiert Setup-Tool
- Automated deployment in firmenweite Tool-Repositories
- Template-Bibliothek mit Workflow-Patterns

## 🏢 Enterprise Business Case

### 40 Jahre EDV-Erfahrung: Das Kernproblem

**Wiederkehrende kritische Aufgaben** in ERP/Logistik/Versicherungs-IT:
- ❌ Menschlicher Faktor: Fehler bei komplexen Prozessen
- ❌ ERP-Komplexität: Hunderte Dialoge, Abhängigkeiten
- ❌ Kritische Folgen: Falsche Daten beschädigen Bestände/Buchungen
- ❌ Repetition ohne Systematik: Jedes Mal "neu erfinden"

### SetupKit als "Guided Process Execution" Framework

**Lösungsansatz:**
- 🛡️ **Fehlerprävention**: Wizard führt durch jeden Schritt
- 📋 **Compliance**: Prozesse laufen immer identisch ab
- 🎓 **Skill-Leveling**: Auch weniger erfahrene MA können komplexe Prozesse ausführen
- 📊 **Nachverfolgbarkeit**: Jeder Schritt dokumentiert

**Anwendungsszenarien:**
- Monatsabschluss-Routinen mit 20+ Schritten
- Stammdaten-Pflege mit Validierungsregeln
- Inventory-Updates mit kritischen Bestandsbewegungen
- Compliance-Reporting mit regulatorischen Anforderungen

## 📈 Organisches Entwicklungs-Pattern

### Der Erkenntnispfad
```
Konkretes Problem (Kobalt-GK Setup) 
    ↓
Abstraktion (wiederverwendbares Framework)
    ↓  
Mustererkennung (40 Jahre Erfahrung!)
    ↓
Große Vision (Business Process Framework)
```

### Dual-Track Entwicklung

**Track 1: Immediate Value**
- Kobalt-GK Setup lösen
- Framework-Basis schaffen
- Proof of Concept

**Track 2: Strategic Vision**
- Mind-Mapping für ERP/Logistik-Szenarien
- Architektur für Enterprise-Skalierung
- Business Case entwickeln

## 🔧 Technische Entscheidung: Wails → Webview

### Erkannte Problematik: Framework-Overkill

**Wails - Überdimensioniert für Setup:**
- ❌ Komplexe Build-Pipeline
- ❌ React/Vue/Svelte Integration (nicht nötig für Setup)
- ❌ Hot-Reload, Dev-Server (Setup läuft einmal)
- ❌ Advanced Binding-Features
- ❌ Größere Binary
- ❌ Overkill für einfache HTML-Seiten

**Webview - Perfekt für Setup:**
- ✅ Minimaler Footprint
- ✅ Einfache HTML/CSS/JS Rendering
- ✅ Bidirektionale Go ↔ JS Kommunikation
- ✅ Natives Window-Management
- ✅ Schneller Build
- ✅ Thin Layer Architektur

### Passt zur Architektur-Vision

```go
// Thin Layer: Webview zeigt nur HTML an
webview.NewWindow("Setup", htmlContent)

// State Machine bleibt im Go-Backend
wizard.ProcessEvent(event)

// Minimale JS-Bridge für Events
webview.Bind("nextStep", wizard.NextStep)
```

**Framework-Eigenschaften:**
- 🪶 **Leichtgewichtig**: Setup-Tools sollen klein und schnell sein
- 🔧 **Einfach**: Weniger moving parts = weniger Fehlerquellen
- 📦 **Portabel**: Einzelne .exe ohne Dependencies

## 🚀 Nächste Schritte

### Sofortig (Kobalt-GK)
1. Wails-Code zu Webview migrieren
2. HTML/CSS/JS Interface verfeinern
3. Go-Backend State Machine testen

### Mittelfristig (Framework-Basis)
1. Abstrahierte Interfaces definieren
2. SQLite-Runtime-System implementieren
3. Multi-Interface-Support (CLI/Web) ausbauen

### Langfristig (Enterprise Vision)
1. Code-Generator-Prototyp
2. Template-System entwickeln
3. Business Process Patterns sammeln

## 💭 Mind-Map Phase Erkenntnisse

- **Database-Abstraktion**: Interface-Design für spätere Enterprise-DB-Integration
- **Requirement-Definition**: DSL oder GUI-Builder für Workflow-Erstellung
- **Template-Bibliothek**: Sammlung von Workflow-Patterns
- **Zielgruppen**: IT-Admins, Developers, Business-Users
- **Customization-Level**: Balance zwischen Flexibilität und Einfachheit

## 📝 Schlüsselerkenntnis

**"Echter Bedarf + Domain-Expertise + Organisches Wachstum"**

SetupKit entwickelt sich mit echten Anforderungen, basiert auf 40 Jahren praktischer Erfahrung und behält durch Kobalt-GK den Reality Check. Viele der besten Enterprise-Tools sind genau so entstanden: Konkrete Probleme führen zur Erkennung universeller Muster.

---

**Fortsetzung der Chat-Session vom:** 23. August 2025  
**Fokus:** Framework-Vision, Enterprise-Anwendung, technische Entscheidungen  
**Status:** Mind-Map Phase, organische Entwicklung  
**Nächster Schritt:** Wails → Webview Migration für Kobalt-GK

# SetupKit - Modernes Installer-Framework

SetupKit ist ein leistungsstarkes, plattformübergreifendes Installer-Framework, das in Go geschrieben wurde und es Entwicklern ermöglicht, professionelle Installer mit minimalem Code zu erstellen. Es bietet mehrere UI-Modi (GUI, CLI, Silent) und verwendet einen DFA-basierten Controller für konsistente Installationsabläufe.

## 🚀 Features

### Einzeldatei-Architektur
- **Alles eingebettet**: Konfiguration, Assets und Installationsdateien sind in der Ausführungsdatei eingebettet
- **Keine Abhängigkeiten**: Einzelne .exe-Datei enthält vollständigen Installer - keine externen Dateien erforderlich
- **Enterprise-bereit**: Perfekt für Unternehmensumgebungen und Massen-Rollouts
- **Optionale Überschreibung**: Externe YAML-Dateien können Verhalten bei Bedarf anpassen

### Multi-Modal UI-Unterstützung
- **GUI-Modus**: Browser-basierte Benutzeroberfläche mit HTML/CSS/JavaScript
- **CLI-Modus**: Interaktive Kommandozeilenschnittstelle
- **Silent-Modus**: Unbeaufsichtigte Installation für Automatisierung
- **Auto-Modus**: Wählt automatisch die beste UI für die Umgebung

### Enterprise-Konfigurationssystem
- **Standardmäßig eingebettet**: Konfiguration und Assets sind in der Ausführungsdatei kompiliert
- **Externe Datei-Überschreibung**: Verwende `-config=datei.yml` für spezifische Rollouts
- **Massen-Rollout-Unterstützung**: Einzelner Installer kann für verschiedene Umgebungen konfiguriert werden
- Komponentendefinitionen mit Dateilisten
- Installationsprofile (minimal, vollständig, entwickler)
- Lizenzvereinbarungsunterstützung
- Erweiterte Einstellungen (Verknüpfungen, PATH, Verifizierung)

### DFA-kontrollierter Ablauf
- Deterministische Endliche Automaten gewährleisten konsistenten Installationsablauf
- Gleiche Ablauflogik für alle UI-Modi
- Zustände: Welcome → License → Components → Install Path → Summary → Progress → Complete

### HTML-Builder-System
- Programmatische HTML-Generierung für Installer-Seiten
- Server-side Rendering (SSR) für dynamische Inhalte
- Responsives Design mit integrierten CSS-Frameworks

## 📁 Projektstruktur

```
setupkit/
├── cmd/                    # Kommandozeilen-Tools
├── pkg/
│   ├── installer/
│   │   ├── core/          # Kern-Installationslogik
│   │   ├── controller/    # DFA-basierter Ablaufcontroller
│   │   └── ui/           # UI-Implementierungen (CLI, GUI, Silent)
│   ├── html/             # HTML-Builder und SSR-System
│   └── wizard/           # DFA-Zustandsmaschinen-Implementierung
├── examples/
│   └── installer-demo/   # Vollständiges Beispiel-Installer
└── bin/                  # Erstellte Binärdateien
```

## 🛠️ Schnellstart

### 1. Beispiel erstellen

```bash
# Mit Make
make build

# Mit Mage
mage build

# Mit Go direkt
go build -o bin/setupkit-installer-demo.exe ./examples/installer-demo
```

### 2. Verschiedene Modi ausführen

```bash
# GUI-Modus (öffnet Browser) - verwendet eingebettete Konfiguration
./bin/setupkit-installer-demo.exe -mode=gui

# CLI-Modus (interaktives Terminal) - verwendet eingebettete Konfiguration
./bin/setupkit-installer-demo.exe -mode=cli

# Silent-Modus mit Profil - verwendet eingebettete Konfiguration
./bin/setupkit-installer-demo.exe -profile=minimal -unattended -dir="./install"

# Externe Konfigurationsdatei verwenden um eingebettete zu überschreiben
./bin/setupkit-installer-demo.exe -config=benutzerdefiniert-installer.yml -mode=gui

# Verfügbare Profile auflisten (aus eingebetteter Konfiguration)
./bin/setupkit-installer-demo.exe -list-profiles

# Profile aus externer Konfiguration auflisten
./bin/setupkit-installer-demo.exe -config=installer.yml -list-profiles
```

## 📝 Konfiguration

Das Installationsverhalten wird in `installer.yml` definiert. Die Konfiguration ist **standardmäßig eingebettet** und kann mit einer externen Datei überschrieben werden:

```yaml
app_name: "DemoApp"
version: "1.0.0"
publisher: "SetupKit Framework"
mode: "auto"
unattended: false

# Zu installierende Komponenten
components:
  - id: "core"
    name: "Kernanwendung"
    required: true
    files: ["README.txt", "LICENSE.txt", "config.json"]

# Installationsprofile
profiles:
  minimal:
    description: "Minimale Installation"
    components: ["core"]
  full:
    description: "Vollständige Installation"
    components: ["core", "docs", "examples"]
```

## 🏗️ Architektur

### MVC-Pattern mit DFA-Controller

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI View      │    │   GUI View      │    │  Silent View    │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────────────────────────────┐
         │          InstallerController                  │
         │          (DFA-basierte Ablaufsteuerung)       │
         └───────────────────────────────────────────────┘
                                 │
         ┌───────────────────────────────────────────────┐
         │             Core Installer                    │
         │         (Geschäftslogik & Dateioperationen)   │
         └───────────────────────────────────────────────┘
```

### Implementierungsstatus der Komponentenkette

Alle Installer-Modi implementieren die vollständige Komponentenkette, wie sie in den Design-Regeln definiert ist:

**✅ Vollständige Kette**: `Welcome → License → Components → Install Path → Summary → Progress → Complete`

| UI-Modus | Welcome | License | Components | Install Path | Summary | Progress | Complete |
|----------|---------|---------|------------|-------------|---------|----------|----------|
| **Silent** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **CLI** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **GUI** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**Alle UI-Modi implementieren jetzt vollständig die komplette Komponentenkette mit ordnungsgemäßem HTML-Rendering und HTTP-API-Interaktion.**

### Zustandsablauf

```
Welcome → License → Components → Install Path → Summary → Progress → Complete
   ↑         ↑          ↑            ↑            ↑          ↑         ↑
   └─────────┴──────────┴────────────┴────────────┴──────────┴─────────┘
                        (Zurück-Navigation unterstützt)
```

## 🎯 Kernkonzepte

### DFA-Controller
- **Single Source of Truth**: Ein Controller verwaltet den gesamten Ablauf
- **Zustandsverwaltung**: Klare Zustände mit Validierung und Übergängen
- **UI-agnostisch**: Gleiche Logik funktioniert für alle UI-Modi

### View-Interface
Alle UI-Implementierungen müssen das `InstallerView`-Interface erfüllen:
```go
type InstallerView interface {
    ShowWelcome() error
    ShowLicense(license string) (accepted bool, err error)
    ShowComponents(components []core.Component) (selected []core.Component, err error)
    ShowInstallPath(defaultPath string) (path string, err error)
    ShowSummary(config *core.Config, selectedComponents []core.Component, installPath string) (proceed bool, err error)
    ShowProgress(progress *core.Progress) error
    ShowComplete(summary *core.InstallSummary) error
    OnStateChanged(oldState, newState wizard.State) error
}
```

## 🔧 Entwicklung

### Build-System
- **Makefile**: Traditionelle Make-Ziele
- **Magefile**: Go-basierte Build-Automatisierung
- **Direktes Go**: Standard go build Kommandos

### Verfügbare Ziele
```bash
# Alle Beispiele erstellen
make build / mage build

# Tests ausführen
make test / mage test

# Installer-Demo ausführen
make run / mage run

# Build-Artefakte bereinigen
make clean / mage clean
```

## 📚 Beispiele

### Erstellen eines benutzerdefinierten Installers

1. **Konfiguration definieren** (`installer.yml`)
2. **Hauptfunktion erstellen** (siehe `examples/installer-demo/main.go`)
3. **DFA-Controller einrichten**
4. **UI-Modus wählen**
5. **Erstellen und bereitstellen**

### Unbeaufsichtigte Installation

```bash
# Minimales Profil, benutzerdefiniertes Verzeichnis
./installer -profile=minimal -unattended -dir="/opt/myapp"

# Vollständiges Profil, Lizenz automatisch akzeptieren
./installer -profile=full -unattended -accept-license
```

## 🏢 Enterprise-Rollout-Szenarien

### Einzeldatei-Distribution
```bash
# Standard-Rollout - alles eingebettet
MeineApp-Installer.exe
```
- **Keine Abhängigkeiten**: Vollständiger Installer in einzelner Ausführungsdatei
- **Netzwerk-Rollout**: Einfache Verteilung über Dateifreigaben, E-Mail oder Download
- **Offline-Installation**: Keine Internetverbindung erforderlich
- **Virenscanner-freundlich**: Einzelne Datei für Sicherheitsscanning

### Massen-Rollout mit Konfigurationsüberschreibung
```bash
# Unternehmens-Rollout mit benutzerdefinierter Konfiguration
MeineApp-Installer.exe -config=unternehmens-config.yml -silent

# Verschiedene Umgebungen
MeineApp-Installer.exe -config=entwicklung.yml -profile=developer
MeineApp-Installer.exe -config=produktion.yml -profile=minimal
```

### Unbeaufsichtigte Enterprise-Installation
```bash
# Stille Installation für Rollout-Tools (SCCM, Intune, etc.)
MeineApp-Installer.exe -silent -profile=minimal -dir="C:\Program Files\MeineApp"

# Batch-Rollout-Skript
for /f %%i in (computer.txt) do (
    psexec \\%%i -c MeineApp-Installer.exe -silent -unattended
)
```

### Konfigurationsvorlagen
Erstelle umgebungsspezifische YAML-Dateien für verschiedene Rollout-Szenarien:

**entwicklung.yml** - Entwickler-Arbeitsplätze
```yaml
install_dir: "C:\Dev\MeineApp"
profiles:
  developer:
    components: ["core", "docs", "examples", "debug-tools"]
    add_to_path: true
    create_shortcuts: true
```

**produktion.yml** - Produktionsserver
```yaml
mode: "silent"
unattended: true
install_dir: "C:\Program Files\MeineApp"
profiles:
  minimal:
    components: ["core"]
    create_shortcuts: false
```

## 🌐 Plattformübergreifende Unterstützung

- **Windows**: Native Unterstützung mit .exe-Binärdateien
- **macOS**: Native Unterstützung mit ordnungsgemäßer App-Struktur
- **Linux**: Native Unterstützung mit Standard-Verzeichnislayouts

## 📄 Lizenz

MIT-Lizenz - siehe LICENSE-Datei für Details.

## 🤝 Mitwirken

1. Folgen Sie den bestehenden Architekturmustern
2. Stellen Sie sicher, dass alle UI-Modi konsistent funktionieren
3. Fügen Sie Tests für neue Funktionalität hinzu
4. Aktualisieren Sie die Dokumentation

## 🔗 Links

- [Architekturdokumentation](ARCHITECTURE.md)
- [API-Referenz](docs/API.md)
- [Beispiele](examples/)
- [Design-Regeln](design-rules_de.md)
- [Todo-Liste](Todo_de.md)

## 📈 Status

**Aktueller Entwicklungsstand (September 2025):**
- ✅ Alle UI-Modi (Silent, CLI, GUI) vollständig implementiert
- ✅ Komplette Komponentenkette in allen Modi
- ✅ HTML-Rendering-System vollständig
- ✅ DFA-Controller und Zustandsverwaltung
- ✅ Tests und Build-System funktionsfähig

**Nächste Schritte:**
- Verbesserte Fehlerbehandlung
- Cross-Platform-Testing
- Performance-Optimierungen
- Erweiterte GUI-Features

---

*SetupKit - Professionelle Installer mit minimalem Aufwand erstellen.*
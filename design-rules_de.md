# SetupKit - Design Rules & Architecture Principles (Deutsche Version)

## 🏗️ Core Architecture Principles

### DFA-First Design
- **Rule 1**: Alle Installer UI Modi (CLI, GUI, Silent) basieren auf der DFA (Deterministic Finite Automaton) State Machine
- **Rule 2**: Keine Ablaufsteuerung außerhalb der DFA Controller erzeugen - alle State-Transitions laufen über `InstallerController`
- **Rule 3**: UI Implementierungen sind reine "Views" - sie enthalten keine Geschäftslogik oder Flow-Control
- **Rule 4**: State-Validierung erfolgt ausschließlich im DFA Controller, nicht in den UI Views

### MVC Pattern Enforcement
- **Model**: `core.Installer`, `core.Config`, `core.Component` - Business Logic und Daten
- **View**: `ui.Silent`, `ui.CLI`, `ui.GUI` - Nur Darstellung und User Input Collection
- **Controller**: `controller.InstallerController` - Flow Control und State Management
- **Rule 5**: Views dürfen niemals direkt mit dem Model interagieren - nur über Controller
- **Rule 6**: Controller kennt keine UI-spezifischen Details - arbeitet über `InstallerView` Interface

## 🎯 Interface Design Rules

### InstallerView Interface
- **Rule 7**: Alle UI Modi implementieren das `InstallerView` Interface
- **Rule 8**: Interface Methoden sind zustandslos - sie erhalten alle nötigen Parameter
- **Rule 9**: UI-spezifische Methoden gehören nicht ins gemeinsame Interface
- **Rule 10**: Return Values sind konsistent: `(result, error)` Pattern

### Core.UI Compatibility  
- **Rule 11**: Jede UI Implementation muss beide Interfaces unterstützen: `InstallerView` + `core.UI`
- **Rule 12**: Adapter Pattern verwenden für Interface-Kompatibilität, keine Doppel-Implementation
- **Rule 13**: `core.UI` Methoden delegieren an `InstallerView` Methoden wo möglich

## 🔄 State Management Rules

### DFA State Transitions
- **Rule 14**: States sind im `controller` Package als Konstanten definiert
- **Rule 15**: State-Reihenfolge: `Welcome → License → Components → InstallPath → Summary → Progress → Complete`
- **Rule 16**: Conditional States: License kann übersprungen werden wenn `config.License == ""`
- **Rule 17**: State-Transitions nur über DFA Actions: `Next`, `Back`, `Cancel`

### State Validation
- **Rule 18**: Jeder State hat eine Validation Function im Controller
- **Rule 19**: Validation Failures verhindern State Transitions
- **Rule 20**: UI Views sammeln nur Daten - Validierung erfolgt im Controller

## 💾 Data Flow Rules

### Configuration Management
- **Rule 21**: `core.Config` ist Single Source of Truth für Installation Parameter
- **Rule 22**: UI Views lesen Config, schreiben nie direkt - Updates über Controller
- **Rule 23**: YAML Config überschreibt Default Values, CLI Args überschreiben YAML
- **Rule 24**: Sensitive Data (Paths, Passwords) werden validiert bevor sie verwendet werden

### Component Selection
- **Rule 25**: Component State wird im Installer Model verwaltet, nicht in Views
- **Rule 26**: Required Components können nicht deselektiert werden
- **Rule 27**: Component Dependencies werden im Model aufgelöst

## 🎨 UI Implementation Rules

### Silent UI
- **Rule 28**: Silent UI führt keine interaktive Dialoge - alle Parameter müssen vorkonfiguriert sein
- **Rule 29**: Bei fehlenden Required Parameters: Error, nicht Prompt
- **Rule 30**: Logging ist einziger Output - kein Console Output

### CLI UI
- **Rule 31**: CLI wartet auf User Input nur bei interaktiven Methoden
- **Rule 32**: CLI zeigt Progress als Text-basierte Updates
- **Rule 33**: CLI Input Validation erfolgt vor DFA Controller Aufruf
- **Rule 34**: EOF/Interrupt führt zu sauberem Exit, nicht zu Crash

### GUI UI  
- **Rule 35**: GUI State wird über HTTP API synchronisiert
- **Rule 36**: HTTP Handlers delegieren an DFA Controller, implementieren nicht selbst Logic
- **Rule 37**: GUI zeigt immer den aktuellen DFA State - keine independent Navigation
- **Rule 38**: Browser-Events (Next/Back/Cancel) werden zu DFA Actions gemappt

## 🧪 Testing Rules

### Test Structure
- **Rule 39**: Tests verwenden `testify/suite` für strukturierte Test Organization
- **Rule 40**: Mock UIs für Controller Testing - nie echte UI in Unit Tests
- **Rule 41**: Integration Tests testen UI + Controller, Unit Tests testen einzelne Components
- **Rule 42**: DFA Controller Tests verwenden Mock InstallerView, nicht Mock core.UI

### Test Coverage Requirements
- **Rule 43**: Jeder DFA State braucht positive und negative Test Cases
- **Rule 44**: Alle UI Modi brauchen Initialization/Shutdown Tests  
- **Rule 45**: Error Handling Paths müssen getestet werden
- **Rule 46**: Cross-UI Consistency Tests für unified Flow Verification

## 📦 Package Organization Rules

### Import Dependencies
- **Rule 47**: `ui` Package darf `controller` und `core` importieren
- **Rule 48**: `controller` Package importiert nur `core` und `wizard` 
- **Rule 49**: `core` Package hat keine Dependencies auf `ui` oder `controller`
- **Rule 50**: Circular Dependencies sind verboten

### Interface Placement
- **Rule 51**: Shared Interfaces gehören ins `core` Package
- **Rule 52**: UI-spezifische Interfaces gehören ins `ui` Package
- **Rule 53**: Controller Interfaces gehören ins `controller` Package

## 🚫 Anti-Patterns (Vermeiden!)

### Flow Control Anti-Patterns
- ❌ **Hardcoded State Sequences** in UI Implementation
- ❌ **Direct Model Manipulation** von UI Views
- ❌ **Bypass DFA Controller** für State Changes
- ❌ **UI-specific Business Logic** in View Layer

### Error Handling Anti-Patterns  
- ❌ **Silent Error Swallowing** ohne Logging
- ❌ **Inconsistent Error Types** zwischen UI Modi
- ❌ **Error Display in Wrong Layer** (Model showing UI Errors)

### Testing Anti-Patterns
- ❌ **Real UI in Unit Tests** (verwende Mocks)
- ❌ **Hardcoded Test Data** (verwende Test Fixtures)
- ❌ **Test Dependencies on External Resources** (Files, Network)

## 🔧 Development Workflow

### Code Review Checklist
1. ✅ Folgt DFA-First Prinzip?
2. ✅ Keine Flow-Control in UI Views?
3. ✅ Interfaces korrekt implementiert?
4. ✅ Error Handling konsistent?
5. ✅ Tests für neue Features vorhanden?
6. ✅ Keine Circular Dependencies?
7. ✅ Documentation aktualisiert?

### Refactoring Guidelines
- **Before**: Verstehe aktuelle State Flow und Dependencies
- **During**: Ein Interface/Layer nach dem anderen refactoren
- **After**: Tests laufen und Behavior ist identisch
- **Never**: Breaking Changes ohne Migration Path

## 📋 Enforcement

### Compiler-Level Checks
- Go Compiler verhindert Interface Violations
- `go vet` prüft auf häufige Errors
- Import Cycle Detection durch Go Toolchain

### Runtime Validation  
- DFA Controller validates State Transitions
- Interface Implementation wird zur Runtime geprüft
- Config Validation beim Startup

### Code Review Requirements
- Jede Änderung an Core Interfaces braucht Review
- DFA State Changes brauchen explizite Tests
- Performance Impacts müssen dokumentiert werden

---
*Diese Design Rules sind bindend für alle SetupKit Entwicklung. Exceptions nur nach Team Discussion und Dokumentation der Gründe.*
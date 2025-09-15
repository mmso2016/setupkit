# SetupKit - Design Rules & Architecture Principles

## 🏗️ Core Architecture Principles

### DFA-First Design
- **Rule 1**: All installer UI modes (CLI, GUI, Silent) are based on the DFA (Deterministic Finite Automaton) state machine
- **Rule 2**: No flow control outside the DFA controller - all state transitions go through `InstallerController`
- **Rule 3**: UI implementations are pure "Views" - they contain no business logic or flow control
- **Rule 4**: State validation occurs exclusively in the DFA controller, not in UI views

### MVC Pattern Enforcement
- **Model**: `core.Installer`, `core.Config`, `core.Component` - Business logic and data
- **View**: `ui.Silent`, `ui.CLI`, `ui.GUI` - Only presentation and user input collection
- **Controller**: `controller.InstallerController` - Flow control and state management
- **Rule 5**: Views must never interact directly with the model - only through controller
- **Rule 6**: Controller knows no UI-specific details - works through `InstallerView` interface

## 🎯 Interface Design Rules

### InstallerView Interface
- **Rule 7**: All UI modes implement the `InstallerView` interface
- **Rule 8**: Interface methods are stateless - they receive all necessary parameters
- **Rule 9**: UI-specific methods do not belong in the common interface
- **Rule 10**: Return values are consistent: `(result, error)` pattern

### Core.UI Compatibility  
- **Rule 11**: Each UI implementation must support both interfaces: `InstallerView` + `core.UI`
- **Rule 12**: Use adapter pattern for interface compatibility, no duplicate implementation
- **Rule 13**: `core.UI` methods delegate to `InstallerView` methods where possible

## 🔄 State Management Rules

### DFA State Transitions
- **Rule 14**: States are defined as constants in the `controller` package
- **Rule 15**: State order: `Welcome → License → Components → InstallPath → Summary → Progress → Complete`
- **Rule 16**: Conditional states: License can be skipped when `config.License == ""`
- **Rule 17**: State transitions only through DFA actions: `Next`, `Back`, `Cancel`

### State Validation
- **Rule 18**: Each state has a validation function in the controller
- **Rule 19**: Validation failures prevent state transitions
- **Rule 20**: UI views only collect data - validation happens in controller

## 💾 Data Flow Rules

### Embedded Architecture (Core Principle)
- **Rule 21**: All installation files must be embedded using `//go:embed` directives
- **Rule 22**: Configuration (installer.yml) must be embedded with external override capability
- **Rule 23**: Single executable must be completely self-contained - zero external dependencies
- **Rule 24**: External configuration files can override embedded config for enterprise deployments

### Configuration Management
- **Rule 25**: `core.Config` is single source of truth for installation parameters
- **Rule 26**: UI views read config, never write directly - updates through controller
- **Rule 27**: Embedded config is default, external YAML overrides embedded, CLI args override YAML
- **Rule 28**: Sensitive data (paths, passwords) are validated before use

### Component Selection
- **Rule 29**: Component state is managed in the installer model, not in views
- **Rule 30**: Required components cannot be deselected
- **Rule 31**: Component dependencies are resolved in the model

## 🎨 UI Implementation Rules

### Silent UI
- **Rule 32**: Silent UI performs no interactive dialogs - all parameters must be pre-configured
- **Rule 33**: For missing required parameters: error, not prompt
- **Rule 34**: Logging is the only output - no console output

### CLI UI
- **Rule 35**: CLI waits for user input only in interactive methods
- **Rule 36**: CLI shows progress as text-based updates
- **Rule 37**: CLI input validation occurs before DFA controller call
- **Rule 38**: EOF/Interrupt leads to clean exit, not crash

### GUI UI
- **Rule 39**: GUI state is synchronized via HTTP API
- **Rule 40**: HTTP handlers delegate to DFA controller, don't implement logic themselves
- **Rule 41**: GUI always shows current DFA state - no independent navigation
- **Rule 42**: Browser events (Next/Back/Cancel) are mapped to DFA actions

## 🧪 Testing Rules

### Test Structure
- **Rule 43**: Tests use `testify/suite` for structured test organization
- **Rule 44**: Mock UIs for controller testing - never real UI in unit tests
- **Rule 45**: Integration tests test UI + Controller, unit tests test individual components
- **Rule 46**: DFA controller tests use Mock InstallerView, not Mock core.UI

### Test Coverage Requirements
- **Rule 47**: Each DFA state needs positive and negative test cases
- **Rule 48**: All UI modes need initialization/shutdown tests
- **Rule 49**: Error handling paths must be tested
- **Rule 50**: Cross-UI consistency tests for unified flow verification

## 📦 Package Organization Rules

### Import Dependencies
- **Rule 51**: `ui` package may import `controller` and `core`
- **Rule 52**: `controller` package imports only `core` and `wizard`
- **Rule 53**: `core` package has no dependencies on `ui` or `controller`
- **Rule 54**: Circular dependencies are forbidden

### Interface Placement
- **Rule 55**: Shared interfaces belong in the `core` package
- **Rule 56**: UI-specific interfaces belong in the `ui` package
- **Rule 57**: Controller interfaces belong in the `controller` package

## 🚫 Anti-Patterns (Avoid!)

### Embedding Anti-Patterns
- ❌ **External File Dependencies** - requiring separate asset files at runtime
- ❌ **Runtime File Loading** - loading configuration from filesystem during execution
- ❌ **Partial Embedding** - embedding some files but not others
- ❌ **Ignoring External Override** - not allowing configuration customization

### Flow Control Anti-Patterns
- ❌ **Hardcoded State Sequences** in UI implementation
- ❌ **Direct Model Manipulation** by UI views
- ❌ **Bypass DFA Controller** for state changes
- ❌ **UI-specific Business Logic** in view layer

### Error Handling Anti-Patterns
- ❌ **Silent Error Swallowing** without logging
- ❌ **Inconsistent Error Types** between UI modes
- ❌ **Error Display in Wrong Layer** (model showing UI errors)

### Testing Anti-Patterns
- ❌ **Real UI in Unit Tests** (use mocks)
- ❌ **Hardcoded Test Data** (use test fixtures)
- ❌ **Test Dependencies on External Resources** (files, network)

## 🔧 Development Workflow

### Code Review Checklist
1. ✅ All files embedded with `//go:embed` directives?
2. ✅ Follows DFA-First principle?
3. ✅ No flow control in UI views?
4. ✅ Interfaces correctly implemented?
5. ✅ External configuration override supported?
6. ✅ Error handling consistent?
7. ✅ Tests for new features present?
8. ✅ No circular dependencies?
9. ✅ Documentation updated?

### Refactoring Guidelines
- **Before**: Understand current state flow and dependencies
- **During**: Refactor one interface/layer at a time
- **After**: Tests pass and behavior is identical
- **Never**: Breaking changes without migration path

## 📋 Enforcement

### Compiler-Level Checks
- Go compiler prevents interface violations
- `go vet` checks for common errors
- Import cycle detection through Go toolchain

### Runtime Validation  
- DFA controller validates state transitions
- Interface implementation is checked at runtime
- Config validation at startup

### Code Review Requirements
- Every change to core interfaces requires review
- DFA state changes need explicit tests
- Performance impacts must be documented

---
*These design rules are binding for all SetupKit development. Exceptions only after team discussion and documentation of reasons.*
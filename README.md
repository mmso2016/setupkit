# SetupKit - Modern Installer Framework

SetupKit is a powerful, cross-platform installer framework written in Go that enables developers to create professional installers with minimal code. It provides multiple UI modes (GUI, CLI, Silent) and uses a DFA-based controller for consistent installation flows.

## 🚀 Features

### Single-File Architecture
- **Everything embedded**: Configuration, assets, and installation files are embedded in the executable
- **Zero dependencies**: Single .exe file contains complete installer - no external files needed
- **Enterprise ready**: Perfect for corporate environments and mass deployments
- **Optional override**: External YAML files can customize behavior when needed

### Multi-Modal UI Support
- **GUI Mode**: Browser-based interface with HTML/CSS/JavaScript
- **CLI Mode**: Interactive command-line interface
- **Silent Mode**: Unattended installation for automation
- **Auto Mode**: Automatically selects the best UI for the environment

### Enterprise Configuration System
- **Embedded by default**: Configuration and assets are compiled into the executable
- **External file override**: Use `-config=file.yml` to customize for specific deployments
- **Mass deployment support**: Single installer can be configured for different environments
- Component definitions with file lists
- Installation profiles (minimal, full, developer)
- License agreement support
- Advanced settings (shortcuts, PATH, verification)

### DFA-Controlled Flow
- Deterministic Finite Automaton ensures consistent installation flow
- Same flow logic for all UI modes
- States: Welcome → License → Components → Install Path → Summary → Progress → Complete
- **Custom States**: Add custom configuration steps anywhere in the flow

### Custom States Support
- **Extensible Flow**: Add custom configuration steps (database setup, service configuration, etc.)
- **Type-Safe Interface**: Strongly typed handlers with validation
- **UI Integration**: Works with all UI modes (Silent, CLI, GUI)
- **Flexible Positioning**: Insert states anywhere in the installation flow
- **Built-in Examples**: Database configuration, user setup, license validation

### HTML Builder System
- Programmatic HTML generation for installer pages
- Server-side rendering (SSR) for dynamic content
- Responsive design with built-in CSS frameworks

## 📁 Project Structure

```
setupkit/
├── cmd/                         # Command-line tools
├── pkg/
│   ├── installer/
│   │   ├── core/               # Core installation logic
│   │   ├── controller/         # DFA-based flow controller + Custom States
│   │   └── ui/                # UI implementations (CLI, GUI, Silent)
│   ├── html/                  # HTML builder and SSR system
│   └── wizard/                # DFA state machine implementation
├── examples/
│   ├── installer-demo/        # Complete example installer
│   └── custom-state-demo/     # Database configuration example
└── bin/                       # Built binaries
```

## 🛠️ Quick Start

### 1. Build the Examples

```bash
# Build main installer demo
make build                    # or: mage build
go build -o bin/setupkit-installer-demo.exe ./examples/installer-demo

# Build custom state demo (database configuration)
make build-custom-state-demo  # or: mage BuildCustomStateDemo
go build -o bin/setupkit-custom-state-demo.exe ./examples/custom-state-demo

# Build all examples
make build-all               # or: mage BuildAll
```

### 2. Run Different Modes

```bash
# Main installer demo
./bin/setupkit-installer-demo.exe -mode=gui     # GUI Mode (opens browser)
./bin/setupkit-installer-demo.exe -mode=cli     # CLI Mode (interactive terminal)
./bin/setupkit-installer-demo.exe -profile=minimal -unattended -dir="./install"  # Silent Mode

# Custom state demo (database configuration)
./bin/setupkit-custom-state-demo.exe --help              # Show all available options
./bin/setupkit-custom-state-demo.exe -mode=silent        # Automated demo
./bin/setupkit-custom-state-demo.exe -mode=cli           # Interactive CLI with DB config
./bin/setupkit-custom-state-demo.exe -mode=auto          # Auto-select best UI mode

# Using make/mage for custom state demo
make run-custom-state-demo                       # or: mage RunCustomStateDemo (silent mode)
make run-custom-state-demo-cli                   # or: mage RunCustomStateDemoCLI (CLI mode)
make run-custom-state-demo-auto                  # or: mage RunCustomStateDemoAuto (auto mode)
make help-custom-state-demo                      # or: mage HelpCustomStateDemo (show help)

# Using external config file to override embedded
./bin/setupkit-installer-demo.exe -config=custom-installer.yml -mode=gui

# List available profiles
./bin/setupkit-installer-demo.exe -list-profiles
```

## 📝 Configuration

Installation behavior is defined in `installer.yml`. The configuration is **embedded by default** and can be overridden with an external file:

```yaml
app_name: "DemoApp"
version: "1.0.0"
publisher: "SetupKit Framework"
mode: "auto"
unattended: false

# Components to install
components:
  - id: "core"
    name: "Core Application"
    required: true
    files: ["README.txt", "LICENSE.txt", "config.json"]

# Installation profiles
profiles:
  minimal:
    description: "Minimal installation"
    components: ["core"]
  full:
    description: "Full installation"
    components: ["core", "docs", "examples"]
```

## 🏗️ Architecture

### MVC Pattern with DFA Controller

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
         │          (DFA-based Flow Control)             │
         └───────────────────────────────────────────────┘
                                 │
         ┌───────────────────────────────────────────────┐
         │             Core Installer                    │
         │         (Business Logic & File Operations)    │
         └───────────────────────────────────────────────┘
```

### Component Chain Implementation Status

All installer modes implement the complete component chain as defined in design rules (Rule 15):

**✅ Complete Chain**: `Welcome → License → Components → Install Path → Summary → Progress → Complete`

| UI Mode | Welcome | License | Components | Install Path | Summary | Progress | Complete |
|---------|---------|---------|------------|-------------|---------|----------|----------|
| **Silent** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **CLI** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **GUI** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**All UI modes now fully implement the complete component chain with proper HTML rendering and HTTP API interaction.**

### State Flow

```
Welcome → License → Components → Install Path → Summary → Progress → Complete
   ↑         ↑          ↑            ↑            ↑          ↑         ↑
   └─────────┴──────────┴────────────┴────────────┴──────────┴─────────┘
                        (Back navigation supported)
```

## 🎯 Key Concepts

### DFA Controller
- **Single Source of Truth**: One controller manages the entire flow
- **State Management**: Clear states with validation and transitions
- **UI Agnostic**: Same logic works for all UI modes

### View Interface
All UI implementations must satisfy the `InstallerView` interface:
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

## 🔧 Development

### Build System
- **Makefile**: Traditional make targets
- **Magefile**: Go-based build automation
- **Direct Go**: Standard go build commands

### Available Targets
```bash
# Build all examples
make build / mage build

# Run tests
make test / mage test

# Run installer demo
make run / mage run

# Clean build artifacts
make clean / mage clean
```

## 📚 Examples

### Creating a Custom Installer

1. **Define Configuration** (`installer.yml`)
2. **Create Main Function** (see `examples/installer-demo/main.go`)
3. **Set up DFA Controller**
4. **Choose UI Mode**
5. **Build and Deploy**

### Unattended Installation

```bash
# Minimal profile, custom directory
./installer -profile=minimal -unattended -dir="/opt/myapp"

# Full profile, accept license automatically
./installer -profile=full -unattended -accept-license
```

## 🏢 Enterprise Deployment Scenarios

### Single-File Distribution
```bash
# Standard deployment - everything embedded
MyApp-Installer.exe
```
- **Zero dependencies**: Complete installer in single executable
- **Network deployment**: Easy distribution via file shares, email, or download
- **Offline installation**: No internet connection required
- **Virus scanner friendly**: Single file for security scanning

### Mass Deployment with Configuration Override
```bash
# Corporate deployment with custom configuration
MyApp-Installer.exe -config=corporate-config.yml -silent

# Different environments
MyApp-Installer.exe -config=development.yml -profile=developer
MyApp-Installer.exe -config=production.yml -profile=minimal
```

### Unattended Enterprise Installation
```bash
# Silent installation for deployment tools (SCCM, Intune, etc.)
MyApp-Installer.exe -silent -profile=minimal -dir="C:\Program Files\MyApp"

# Batch deployment script
for /f %%i in (computers.txt) do (
    psexec \\%%i -c MyApp-Installer.exe -silent -unattended
)
```

### Configuration Templates
Create environment-specific YAML files for different deployment scenarios:

**development.yml** - Developer workstations
```yaml
install_dir: "C:\Dev\MyApp"
profiles:
  developer:
    components: ["core", "docs", "examples", "debug-tools"]
    add_to_path: true
    create_shortcuts: true
```

**production.yml** - Production servers
```yaml
mode: "silent"
unattended: true
install_dir: "C:\Program Files\MyApp"
profiles:
  minimal:
    components: ["core"]
    create_shortcuts: false
```

## 🔧 Custom States

SetupKit supports **Custom States** for adding custom configuration steps to the installer flow.

### Basic Usage

```go
// 1. Create a custom state handler
type DatabaseConfigHandler struct {
    *controller.BaseCustomStateHandler
}

func NewDatabaseConfigHandler() *DatabaseConfigHandler {
    return &DatabaseConfigHandler{
        BaseCustomStateHandler: &controller.BaseCustomStateHandler{
            StateID:      controller.StateDBConfig,
            Name:         "Database Configuration",
            Description:  "Configure database connection settings",
            InsertPoint:  controller.InsertAfterInstallPath,
            CanGoNext:    true,
            CanGoBack:    true,
            CanCancel:    true,
        },
    }
}

// 2. Register with controller
controller := controller.NewInstallerController(config, installer)
dbHandler := controller.NewDatabaseConfigHandler()
controller.RegisterCustomState(dbHandler)

// 3. The flow becomes:
// Welcome → License → Components → Install Path → [DB Config] → Summary → Complete
```

### Database Configuration Example

The built-in database configuration example supports multiple database types:

```bash
# Run the database configuration demo in different modes
./bin/setupkit-custom-state-demo.exe --help              # Show all options
./bin/setupkit-custom-state-demo.exe -mode=silent        # Automated demo
./bin/setupkit-custom-state-demo.exe -mode=cli           # Interactive CLI
./bin/setupkit-custom-state-demo.exe -mode=auto          # Auto-select UI mode

# Using make/mage shortcuts
make run-custom-state-demo                       # Silent mode (automated)
make run-custom-state-demo-cli                   # CLI mode (interactive)
make help-custom-state-demo                      # Show all options

# With custom directory
./bin/setupkit-custom-state-demo.exe -mode=silent -dir="./my-install"
```

**Supported databases:**
- MySQL
- PostgreSQL
- SQLite
- SQL Server

**Features:**
- **Multiple UI modes**: Silent (automated), CLI (interactive), Auto (best selection)
- **Flexible configuration**: Command-line parameters for all options
- **Interactive CLI configuration**: Full user input with validation
- **Automatic connection validation**: Tests database connectivity (skipped in demo mode)
- **Connection string generation**: Supports all database types
- **Silent mode with defaults**: Unattended installation with sensible defaults

### Custom State Interface

```go
type CustomStateHandler interface {
    GetStateID() wizard.State
    GetConfig() *wizard.StateConfig
    HandleEnter(*InstallerController, map[string]interface{}) error
    HandleLeave(*InstallerController, map[string]interface{}) error
    Validate(*InstallerController, map[string]interface{}) error
    GetInsertionPoint() InsertionPoint
}
```

### Insertion Points

Insert custom states anywhere in the flow:

```go
var (
    InsertAfterWelcome     = InsertionPoint{After: StateWelcome, Before: StateLicense}
    InsertAfterLicense     = InsertionPoint{After: StateLicense, Before: StateComponents}
    InsertAfterComponents  = InsertionPoint{After: StateComponents, Before: StateInstallPath}
    InsertAfterInstallPath = InsertionPoint{After: StateInstallPath, Before: StateSummary}
    InsertAfterSummary     = InsertionPoint{After: StateSummary, Before: StateProgress}
)
```

### Use Cases

- **Database Setup**: Connection configuration, schema initialization
- **Service Configuration**: API keys, URLs, service endpoints
- **User Management**: Admin account creation, permissions
- **License Validation**: Enterprise license server verification
- **Plugin Selection**: Extended component configuration

### Testing

```bash
# Run custom state tests
go test -v ./pkg/installer/controller -run TestCustomState

# Run simple validation tests
go test -v ./pkg/installer/controller -run TestCustomStateSimple
```

## 🌐 Cross-Platform Support

- **Windows**: Native support with .exe binaries
- **macOS**: Native support with proper app structure
- **Linux**: Native support with standard directory layouts

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

1. Follow the existing architecture patterns
2. Ensure all UI modes work consistently
3. Add tests for new functionality
4. Update documentation

## 🔗 Links

- [API Reference](docs/API.md)
- [Main Installer Example](examples/installer-demo/)
- [Custom States Example](examples/custom-state-demo/)
- [Package Documentation](pkg/)
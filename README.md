# SetupKit

> **Native installers for Go applications - no InnoSetup, InstallShield, or NSIS required.**

[![Go Reference](https://pkg.go.dev/badge/github.com/mmso2016/setupkit.svg)](https://pkg.go.dev/github.com/mmso2016/setupkit)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## 🚧 Under Construction 🚧

**Please be patient, the project will be completed shortly.**

This framework is currently being actively developed. Core functionality is being implemented and the API may change. Feel free to watch/star the repository to get notified about updates!

---

## Why SetupKit?

Traditional installer tools like InnoSetup, InstallShield, or NSIS require learning additional scripting languages and managing external dependencies. The Setup-Kit lets you build professional, native installers using pure Go.

### Key Features

- 🚀 **Pure Go**: No external dependencies or scripting languages
- 🎯 **Three Deployment Modes**: GUI (Wails), CLI, and Silent/Unattended
- 📦 **Embedded Assets**: Single binary with all resources included
- 🔄 **Cross-Platform**: Windows, Linux, and macOS support
- 🏢 **Enterprise-Ready**: Unattended installation, proper exit codes, rollback support
- 🧪 **Testable**: Unit test your installation logic
- ⚡ **Fast**: Compiled Go binaries, no runtime overhead

## Quick Start

### Installation

```bash
go get github.com/mmso2016/setupkit
```

### Basic Example

```go
package main

import (
    "context"
    "embed"
    "log"
    
    "github.com/mmso2016/setupkit/installer"
)

//go:embed assets/*
var assets embed.FS

func main() {
    // Create installer with options
    inst, err := installer.New(
        installer.WithAppName("MyApp"),
        installer.WithVersion("1.0.0"),
        installer.WithMode(installer.ModeAuto), // Auto-detect best UI
        installer.WithAssets(assets),
        installer.WithLicense("MIT License..."),
        
        // UI Configuration & Theming (New!)
        installer.WithTheme("corporate-blue"),          // Built-in theme
        installer.WithUIConfig("installer-config.yaml"), // YAML config
        installer.WithBranding("#1e3a8a", "#3b82f6", "logo.png", "'Inter', sans-serif"),
        
        installer.WithComponents(
            installer.Component{
                ID:          "core",
                Name:        "Core Files",
                Description: "Essential application files",
                Required:    true,
                Size:        10 * 1024 * 1024,
                Installer:   installCore,
            },
        ),
        installer.WithInstallDir("/opt/myapp"),
        installer.WithVerbose(true),
    )
    
    if err != nil {
        log.Fatal("Failed to create installer:", err)
    }
    
    // Run the installer
    if err := inst.Run(); err != nil {
        log.Fatal("Installation failed:", err)
    }
}

func installCore(ctx context.Context) error {
    // Your installation logic here
    return nil
}
```

### Build Installers

See **[BUILD.md](BUILD.md)** for complete build instructions.

Using Make:
```bash
make build              # Build all examples
make build-cli          # Build CLI installer
make build-gui          # Build Wails GUI installer
make build-console      # Build console GUI
make build-platform     # Build platform example
make build-ui           # Build UI example (standard)
make build-ui-wails     # Build UI example with Wails
make build-ui-nogui     # Build UI example without GUI
```

Using Mage:
```bash
mage build              # Build all examples  
mage buildCLI           # Build CLI installer
mage buildGUI           # Build Wails GUI installer
mage buildConsole       # Build console GUI
mage buildPlatform      # Build platform example
mage buildUI            # Build UI example (standard)
mage buildUIWails       # Build UI example with Wails
mage buildUINoGUI       # Build UI example without GUI
```

## Installation Modes

### 1. GUI Mode (Cross-platform)
Perfect for desktop users and first-time installations.

```bash
./installer-gui
# or on Windows
installer-gui.exe
```

### 2. CLI Mode (Platform-specific)
Optimized for each platform with native features.

```bash
# Windows - with UAC elevation and service registration
installer-platform.exe --install --dir "C:\Program Files\MyApp" --service

# Linux - with systemd integration
sudo ./installer-platform --install --dir /opt/myapp --service --path
```

### 3. Silent/Unattended Mode
Enterprise-ready for automation and mass deployment.

```bash
# Using UI example with nogui build tag
./installer-ui-nogui --install-dir /opt/myapp --components core,docs

# Platform installer in silent mode
./installer-platform --install --dir /opt/myapp --service
```

## Exit Codes

Professional exit codes for monitoring and automation:

| Code | Meaning | Category |
|------|---------|----------|
| 0 | Success | - |
| 1-19 | General errors | Configuration, permissions |
| 20-39 | Pre-installation check failures | Prerequisites, disk space, ports |
| 40-59 | Installation failures | Extract, copy, service setup |
| 60-79 | Post-installation failures | Service start, health checks |
| 80-89 | Rollback status | Clean rollback or failed rollback |
| 90-99 | User actions | Cancelled, license declined |

## Architecture

```
setupkit/
├── examples/
│   ├── basic/             # Simple CLI installer example
│   ├── gui/               # Wails-based GUI installer
│   ├── gui-console/       # Console GUI example
│   ├── platform/          # Platform-specific features
│   └── ui/                # UI example with multiple modes
├── installer/
│   ├── core/              # Core installer logic
│   ├── components/        # Component definitions
│   └── ui/                # UI abstraction layer
├── BUILD.md               # Complete build instructions
├── go.mod
├── Makefile               # Make-based build system
└── magefile.go            # Mage-based build automation
```

## Platform-Specific Features

### Windows
- UAC elevation handling
- Windows Service registration
- Registry integration
- MSI database creation (optional)

### Linux
- systemd service installation
- Package manager integration (apt/yum/dnf)
- Desktop entry creation
- User/group management

### macOS
- launchd service installation
- Code signing and notarization support
- pkg bundle creation

## Advanced Features

### Component Selection
```go
installer.WithComponents(
    installer.Component{
        ID:       "core",
        Name:     "Core Application",
        Required: true,
        Size:     50 * 1024 * 1024, // 50MB
    },
    installer.Component{
        ID:       "docs",
        Name:     "Documentation",
        Required: false,
        Size:     10 * 1024 * 1024, // 10MB
    },
)
```

### Custom Validation
```go
installer.WithPreCheck(func(ctx *installer.Context) error {
    // Check for specific requirements
    if !hasPostgreSQL() {
        return installer.ErrMissingDependency("PostgreSQL 14+")
    }
    return nil
})
```

### Rollback Support
```go
installer.WithRollback(installer.RollbackFull) // Automatic rollback on failure
```

## Build Tags

The SetupKit supports build tags to control which features are included:

- **`wails`** - Enable Wails GUI support
- **`nogui`** - Disable all GUI features (CLI/Silent only)
- **`nocli`** - Disable CLI support (not recommended)

### Build Examples
```bash
# Standard build with CLI fallback
go build ./examples/ui

# GUI version with Wails
go build -tags wails ./examples/ui

# Minimal CLI-only version
go build -tags nogui ./examples/ui
```

See **[BUILD.md](BUILD.md)** for complete build instructions and tag combinations.

## 🎨 UI Configuration & Theming

Configure your installer's appearance and behavior using YAML files and built-in themes.

### Quick Start with Themes

```bash
# List available themes
go run ./examples/basic -list-themes

# Use a built-in theme
go run ./examples/basic -theme corporate-blue

# Generate a configuration file
go run ./examples/basic -generate-config my-installer.yaml

# Use your configuration
go run ./examples/basic -config my-installer.yaml
```

### Built-in Themes

| Theme | Description | Best For |
|-------|-------------|----------|
| `default` | Clean modern design | General applications |
| `corporate-blue` | Professional blue theme | Business software |
| `medical-green` | Healthcare-friendly green | Medical applications |
| `tech-dark` | Dark theme for developers | Development tools |
| `minimal-light` | Clean minimal design | Simple utilities |

### YAML Configuration

```yaml
# installer-config.yaml
ui:
  theme: "corporate-blue"
  title: "My Application Installer"

branding:
  primary_color: "#1e3a8a"
  secondary_color: "#3b82f6"
  logo: "assets/logo.png"
  company_name: "My Company"

screens:
  welcome:
    enabled: true
    title: "Welcome to My App"
    message: "This installer will guide you through the setup."
  
  license:
    enabled: true
    title: "License Agreement"
  
  components:
    enabled: true
    title: "Select Features"
  
  directory:
    enabled: true
    title: "Installation Directory"
  
  installation:
    enabled: true
    title: "Installing..."
  
  finish:
    enabled: true
    title: "Installation Complete"
```

### Programmatic Configuration

```go
// Use built-in themes
installer.WithTheme("medical-green")

// Custom branding
installer.WithBranding("#059669", "#10b981", "logo.png", "'Inter', sans-serif")

// Load from YAML
installer.WithUIConfig("config.yaml")

// Configure screens
installer.WithScreenConfig(map[string]bool{
    "license": false,     // Skip license screen
    "components": false, // Skip component selection
})

// Custom welcome message
installer.WithWelcomeMessage("Quick Setup", "This will install MyApp quickly.")
```

### Example Configurations

- 📄 **[Default Config](examples/configs/default.yaml)** - Complete example with all options
- 🏢 **[Corporate Config](examples/configs/corporate.yaml)** - Professional business theme
- 🏥 **[Medical Config](examples/configs/medical.yaml)** - Healthcare-focused design
- ⚡ **[Minimal Config](examples/configs/minimal.yaml)** - Streamlined installation

See **[examples/configs/README.md](examples/configs/README.md)** for detailed configuration documentation.

## Examples & Templates

- 📖 **[BUILD.md](BUILD.md)** - Complete build instructions and build tags
- 🚀 **[Basic CLI Example](examples/basic)** - Simple CLI installer
- 🖼️ **[GUI Example](examples/gui)** - Wails-based GUI installer
- 🖥️ **[Console GUI Example](examples/gui-console)** - Console-based interface
- ⚙️ **[Platform Example](examples/platform)** - Platform-specific features
- 🎛️ **[UI Example](examples/ui)** - Multi-mode UI with build tags
- 📋 **[TEMPLATES.md](TEMPLATES.md)** - Real-world projects and installation patterns

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with:
- [Wails](https://wails.io) - GUI framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Mage](https://magefile.org) - Build tool

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/mmso2016/setupkit)
- 🐛 [Issue Tracker](https://github.com/mmso2016/setupkit/issues)
- 💬 [Discussions](https://github.com/mmso2016/setupkit/discussions)

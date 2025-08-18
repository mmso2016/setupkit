# SetupKit Examples

This directory contains various examples demonstrating different features and use cases of the Setup-Kit framework.

## Available Examples

### 📋 [basic/](basic/) - Simple CLI Installer
**Difficulty:** Beginner  
**Features:** Basic installation with components, embedded assets

```bash
# Build and run
make build-cli
./bin/installer-cli --help
```

A straightforward example showing:
- Component definition and selection
- Embedded assets using `//go:embed`
- Basic installer configuration
- Custom installation functions

### 🖼️ [gui/](gui/) - Wails GUI Installer
**Difficulty:** Intermediate  
**Features:** Cross-platform GUI, Wails integration, frontend

```bash
# Requires Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build and run
make build-gui
./bin/installer-gui

# Development mode
mage dev
```

Demonstrates:
- Full Wails-based GUI installer
- Frontend with HTML/CSS/JavaScript
- Real-time installation progress
- Component selection interface
- Directory browsing

### 🖥️ [gui-console/](gui-console/) - Console GUI
**Difficulty:** Beginner  
**Features:** Text-based interface, cross-platform

```bash
# Build and run
make build-console
./bin/installer-console
```

Shows how to:
- Create a text-based "GUI" installer
- Simulate installation steps
- Display progress in console
- Alternative to full GUI when Wails isn't available

### ⚙️ [platform/](platform/) - Platform-Specific Features
**Difficulty:** Advanced  
**Features:** Service installation, UAC elevation, path management

```bash
# Build and run
make build-platform

# Installation example
./bin/installer-platform --install --dir /opt/myapp --service --path

# Uninstallation example
./bin/installer-platform --uninstall --service
```

Advanced features:
- Windows UAC elevation handling
- Service installation (Windows Services, systemd)
- PATH environment variable management
- Registry operations (Windows)
- Proper uninstallation
- Platform-specific APIs

### 🎛️ [ui/](ui/) - Multi-Mode UI with Build Tags
**Difficulty:** Intermediate  
**Features:** Build tags, multiple UI modes, mode detection

```bash
# Standard build (CLI fallback)
make build-ui
./bin/installer-ui

# With Wails GUI support
make build-ui-wails
./bin/installer-ui-wails

# CLI-only (smallest binary)
make build-ui-nogui
./bin/installer-ui-nogui
```

Demonstrates:
- Build tag usage (`wails`, `nogui`, `nocli`)
- Automatic UI mode detection
- Graceful fallback between modes
- Single codebase, multiple builds
- Response file support

## Build All Examples

```bash
# Using Make
make build                 # Build all examples
make test-builds          # Test all build configurations

# Using Mage
mage build                # Build all examples
mage testAllBuilds        # Test all build configurations
```

## Build Tags Overview

| Tag | Effect | Use Case |
|-----|--------|----------|
| (none) | CLI fallback | Universal binary with GUI fallback |
| `wails` | Enable Wails GUI | Full GUI experience |
| `nogui` | CLI/Silent only | Server deployments, minimal size |
| `nocli` | GUI only | Desktop-only applications |

See **[../BUILD.md](../BUILD.md)** for complete build instructions.

## Getting Started

1. **Start with [basic/](basic/)** - Learn the core concepts
2. **Try [gui-console/](gui-console/)** - See text-based interface
3. **Explore [platform/](platform/)** - Platform-specific features
4. **Advanced: [gui/](gui/)** - Full GUI with Wails
5. **Master: [ui/](ui/)** - Build tags and multi-mode

## Dependencies

### Required
- Go 1.23.5+
- Make or Mage build tool

### Optional
- **Wails CLI** (for GUI examples):
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

### Platform-Specific
- **Windows**: No additional requirements
- **Linux**: May need elevated privileges for service installation
- **macOS**: Xcode command line tools for native builds

## Tips

- **Start Simple**: Begin with the `basic/` example
- **Read the Code**: Each example is heavily commented
- **Check Build Output**: All binaries go to `../bin/`
- **Test Builds**: Use `make test-builds` to verify all configurations
- **Development Mode**: Use `mage dev` for Wails hot-reload

## 🌟 Real-World Examples

Check out **[../TEMPLATES.md](../TEMPLATES.md)** for:
- Production projects using Setup-Kit
- Enterprise deployment patterns
- Community examples and case studies

## 🤝 Contributing

Have an idea for a new example? Found an issue? 
- Open an issue or pull request
- Add your project to **[../TEMPLATES.md](../TEMPLATES.md)**
- Help improve documentation

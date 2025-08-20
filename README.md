# SetupKit - Installer Framework for Go

> **Native installer FRAMEWORK for Go applications - no InnoSetup, InstallShield, or NSIS required.**

[![Go Reference](https://pkg.go.dev/badge/github.com/setupkit.svg)](https://pkg.go.dev/github.com/setupkit)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-available-green.svg)](UI_SYSTEM.md)

---

## 🚧 Under Construction 🚧

This framework is currently being actively developed. Core functionality is working but the API may still change. Watch/star the repository to get notified about updates!

---

## Why SetupKit?

Traditional installer tools require learning scripting languages (Pascal for InnoSetup, NSIS Script, etc.) and managing external tools. SetupKit is different:

- **Write in Go**: Use the language you already know
- **Single Binary**: Everything embedded in one .exe file
- **Framework, not Library**: You write configuration, we handle the complexity
- **Professional UI**: Modern installer with gradient design and animations
- **Native Experience**: Uses Wails/WebView2 for native Windows feel
- **🧪 UNIT TESTABLE**: Test your installer logic with standard Go tests - no other tools needed!

## Quick Start

### Installation

```bash
go get github.com/setupkit
```

### Minimal Example (5 lines!)

```go
package main

import "github.com/setupkit"

func main() {
    setupkit.Install(setupkit.Config{
        AppName: "My App",
        Version: "1.0.0",
    })
}
```

### A More Complex Example

```go
package main

import (
    "log"
    "github.com/setupkit"
)

func main() {
    err := setupkit.Install(setupkit.Config{
        AppName:   "My Application",
        Version:   "2.0.0",
        Publisher: "Your Company",
        Website:   "https://example.com",
        License:   setupkit.LicenseMIT,
        
        Components: []setupkit.Component{
            {
                ID:       "core",
                Name:     "Core Application",
                Size:     45 * setupkit.MB,
                Required: true,
                Selected: true,
            },
            {
                ID:          "docs",
                Name:        "Documentation",
                Description: "User manual and API documentation",
                Size:        12 * setupkit.MB,
                Selected:    true,
            },
        },
        
        // Optional callbacks
        BeforeInstall: func() error {
            // Pre-installation checks
            return nil
        },
        AfterInstall: func() error {
            // Post-installation tasks
            return nil
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

## Building

```bash
# Standard build
go build -o installer.exe main.go

# With build tools (includes Wails tags)
mage build
# or
make build
```

## Unit Testing - The Game Changer! 🧪

**Unlike EVERY other installer tool**, SetupKit installers are fully testable with standard Go tests:

```go
func TestMyInstaller(t *testing.T) {
    config := setupkit.Config{
        AppName: "TestApp",
        Version: "1.0.0",
        Components: []setupkit.Component{
            {ID: "core", Name: "Core", Size: 10*setupkit.MB},
        },
    }
    
    // Test configuration validation
    assert.NoError(t, config.Validate())
    
    // Test component selection logic
    selected := config.GetSelectedComponents()
    assert.Equal(t, 1, len(selected))
    
    // Test your callbacks
    err := config.BeforeInstall()
    assert.NoError(t, err)
}
```

**No special tools needed** - just `go test` like any other Go project! Try that with InnoSetup or NSIS! 😎

## Features

### Framework Provides

- ✅ **Complete UI** - Welcome, License, Components, Progress, Completion pages
- ✅ **Native Window** - Using Wails/WebView2 (no browser)
- ✅ **Progress Tracking** - Real-time installation progress
- ✅ **Component Selection** - Tree view with dependencies
- ✅ **Themes** - Default, Dark, Corporate, or custom
- ✅ **Single Executable** - Everything embedded

### You Provide

- ✅ **Configuration** - App name, version, components
- ✅ **Installation Logic** - What to actually install (optional)
- ✅ **Callbacks** - Pre/post install hooks (optional)

## Customization

### Themes

```go
Theme: setupkit.ThemeDefault    // Professional blue gradient
Theme: setupkit.ThemeDark       // Dark mode
Theme: setupkit.ThemeCorporate  // Corporate style

// Custom theme
Theme: setupkit.Theme{
    PrimaryColor: "#FF6B00",
    CustomCSS: `/* your styles */`,
}
```

### Predefined Constants

```go
// Licenses
License: setupkit.LicenseMIT
License: setupkit.LicenseApache2
License: setupkit.LicenseGPL3

// Sizes
Size: 100 * setupkit.KB
Size: 50 * setupkit.MB
Size: 2 * setupkit.GB
```

## Examples

### Enterprise Installer

```go
setupkit.Install(setupkit.Config{
    AppName:   "Enterprise Suite",
    Publisher: "BigCorp Inc.",
    Theme:     setupkit.ThemeCorporate,
    
    Components: []setupkit.Component{
        {ID: "server", Name: "Server", Size: 500*setupkit.MB, Required: true},
        {ID: "client", Name: "Client", Size: 100*setupkit.MB},
        {ID: "admin", Name: "Admin Tools", Size: 50*setupkit.MB},
    },
    
    BeforeInstall: func() error {
        // Stop services, check prerequisites
        return nil
    },
    AfterInstall: func() error {
        // Start services, create shortcuts
        return nil
    },
})
```

### Game Installer

```go
setupkit.Install(setupkit.Config{
    AppName: "Amazing Game",
    Version: "1.0.0",
    Theme:   setupkit.ThemeDark,
    
    Components: []setupkit.Component{
        {ID: "game", Name: "Game Files", Size: 10*setupkit.GB, Required: true},
        {ID: "hd", Name: "HD Textures", Size: 5*setupkit.GB},
        {ID: "soundtrack", Name: "Soundtrack", Size: 500*setupkit.MB},
    },
})
```

## Project Structure

```
setupkit/
├── setupkit.go              # Main API entry point
├── pkg/
│   └── installer/           # Core installer logic
│       ├── installer.go    # Framework implementation
│       └── assets/          # Embedded UI (HTML/CSS/JS)
├── examples/
│   ├── minimal/            # Minimal example (50 lines)
│   ├── branded/            # Corporate branding example
│   └── simplest/           # Absolute minimum (5 lines)
└── installer/              # Legacy components (being refactored)
```

## Framework vs Library

| Aspect | Traditional Approach | SetupKit Framework |
|--------|---------------------|-------------------|
| UI Code | You write 1000+ lines | Framework provides |
| Build Process | Complex, multiple tools | Simple `go build` |
| Output | Multiple files | Single .exe |
| Dependencies | Node.js, npm, etc. | Just Go |
| Learning Curve | New scripting language | Go configuration |

## Requirements

- Go 1.18+ (for generics and embed)
- Windows 7+ with WebView2 runtime
- For development: Wails CLI (optional, auto-handled by mage/make)

## Documentation

- [UI System](UI_SYSTEM.md) - Understanding the UI framework
- [Examples](examples/README.md) - Complete working examples
- [Contributing](CONTRIBUTING.md) - How to contribute

## Comparison with Other Tools

| Feature | SetupKit | InnoSetup | NSIS | WiX | Electron |
|---------|----------|-----------|------|-----|----------|
| Language | Go | Pascal Script | NSIS Script | XML | JavaScript |
| Single Binary | ✅ | ❌ | ❌ | ❌ | ❌ |
| Cross-Platform | ✅ | ❌ | ❌ | ❌ | ✅ |
| Native UI | ✅ | ✅ | ✅ | ✅ | ❌ |
| File Size | ~10MB | ~5MB | ~3MB | ~5MB | ~100MB+ |
| Learning Curve | Low | Medium | High | High | Medium |
| **Unit Testing** | **✅ Native** | **❌ None** | **❌ None** | **❌ None** | ✅ Complex |

## Roadmap

- [x] Core framework architecture
- [x] Wails/WebView2 integration
- [x] Component selection
- [x] Progress tracking
- [x] Theme system
- [ ] Auto-update support
- [ ] Digital signatures
- [ ] Rollback support
- [ ] macOS/Linux support
- [ ] Cloud analytics

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) file.

## Support

- 🐛 [Report bugs](https://github.com/setupkit/setupkit/issues)
- 💡 [Request features](https://github.com/setupkit/setupkit/issues)
- 📖 [Read documentation](https://github.com/setupkit/setupkit/wiki)
- ⭐ Star the project if you find it useful!

---

**SetupKit** - Modern installer framework for Go applications. Write configuration, ship professional installers.

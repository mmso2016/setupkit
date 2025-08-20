# SetupKit UI System

## What is SetupKit?

SetupKit is a **framework** (not a library!) for creating professional Windows installers with minimal code.

## The Framework Philosophy

```
Your Code (50 lines) → SetupKit Framework → Professional Installer (.exe)
```

You provide configuration, the framework provides everything else:
- Complete UI with multiple pages
- Wails/WebView2 integration
- Installation logic
- Progress tracking
- Single executable output

## Quick Start

### Minimal Example (5 lines)

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

### Full Example with Components

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
        License:   setupkit.LicenseMIT,  // or Apache2, GPL3, or custom
        
        Components: []setupkit.Component{
            {
                ID:       "core",
                Name:     "Core Application",
                Size:     45 * setupkit.MB,  // Convenience constants: KB, MB, GB
                Required: true,
                Selected: true,
            },
            {
                ID:          "docs",
                Name:        "Documentation",
                Description: "User manual and API docs",
                Size:        12 * setupkit.MB,
                Selected:    true,
            },
        },
        
        // Optional callbacks
        OnProgress: func(percent int, status string) {
            // Log progress
        },
        OnComplete: func(path string) {
            // Post-install actions
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

## UI Pages (Provided by Framework)

The framework automatically provides these pages:

1. **Welcome** - Introduction with app name and version
2. **License Agreement** - Show license with accept checkbox
3. **Component Selection** - Choose what to install
4. **Installation Progress** - Real-time progress with animations
5. **Completion** - Success confirmation

All pages are professionally styled with gradient headers and smooth transitions.

## Customization Options

### Themes

```go
Theme: setupkit.ThemeDefault    // Professional blue
Theme: setupkit.ThemeDark       // Dark mode
Theme: setupkit.ThemeCorporate  // Corporate style

// Or custom:
Theme: setupkit.Theme{
    PrimaryColor: "#FF6B00",
    CustomCSS: "/* your styles */",
}
```

### Predefined Licenses

```go
License: setupkit.LicenseMIT
License: setupkit.LicenseApache2
License: setupkit.LicenseGPL3
// Or provide your own string
```

### Size Constants

```go
Size: 100 * setupkit.KB  // Kilobytes
Size: 50 * setupkit.MB   // Megabytes
Size: 2 * setupkit.GB    // Gigabytes
```

### Callbacks

```go
BeforeInstall: func() error {
    // Pre-installation checks
    return nil
}

OnProgress: func(percent int, status string) {
    // Track progress
}

AfterInstall: func() error {
    // Post-installation tasks
    return nil
}

OnComplete: func(path string) {
    // Installation finished
}

OnError: func(err error) {
    // Handle errors
}
```

## Building

```bash
# Standard build
go build -o installer.exe main.go

# With Mage (includes Wails tags)
mage build

# With Make
make build

# Optimized build
go build -ldflags "-s -w" -o installer.exe main.go
```

## How It Works

1. **You write:** Simple configuration in Go
2. **Framework provides:** 
   - Embedded HTML/CSS/JS UI
   - Wails integration with WebView2
   - Native Windows application
   - All installation logic
3. **Result:** Single .exe installer

## Framework vs Library

| Aspect | Traditional Library | SetupKit Framework |
|--------|-------------------|-------------------|
| UI Code | You write 1000+ lines HTML/CSS/JS | Framework provides everything |
| Build Process | Complex, npm, webpack | Simple: `go build` |
| Output | Multiple files and folders | Single .exe |
| Dependencies | Node.js, npm packages | Just Go |
| Learning Curve | HTML, CSS, JS, build tools | Just Go configuration |

## What You DON'T Need

- ❌ No HTML/CSS/JavaScript knowledge
- ❌ No Node.js or npm
- ❌ No frontend framework
- ❌ No wails.json configuration
- ❌ No build directories
- ❌ No asset folders

## What You GET

- ✅ Professional installer UI
- ✅ Single executable file
- ✅ Native Windows experience
- ✅ WebView2 rendering (modern)
- ✅ Smooth animations
- ✅ ~50 lines of code total

## Common Patterns

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
})
```

### Game Installer
```go
setupkit.Install(setupkit.Config{
    AppName: "Awesome Game",
    Theme:   setupkit.ThemeDark,
    Components: []setupkit.Component{
        {ID: "game", Name: "Game Files", Size: 10*setupkit.GB, Required: true},
        {ID: "hd", Name: "HD Textures", Size: 5*setupkit.GB},
    },
})
```

## Requirements

- Go 1.18+
- Windows 7+ (WebView2 runtime)
- For building: Wails CLI (optional, handled by mage/make)

## License

MIT - See LICENSE file

---

**Remember: SetupKit is a framework, not a library. You write configuration, we handle everything else!**

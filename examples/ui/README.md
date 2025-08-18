# UI System Example

This example demonstrates the complete UI system of the Setup-Kit library with multiple interface options.

## Features

The UI system supports multiple modes:

- **CLI Mode**: Command-line interface with interactive prompts
- **GUI Mode**: Native GUI using Wails framework
- **Web Mode**: Browser-based interface
- **Silent Mode**: Unattended installation with response files
- **Auto Mode**: Automatically selects the best available UI

## Building

```bash
# Build the example
go build -o ui-example

# Build with GUI support (requires Wails)
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails build
```

## Usage

### CLI Mode

Interactive command-line installation:

```bash
./ui-example -mode cli
```

Features:
- Step-by-step wizard
- Progress bars with colors
- Component selection
- Directory browsing
- Input validation

### Web Mode

Browser-based installation:

```bash
./ui-example -mode web -port 8080
```

Then open http://localhost:8080 in your browser.

Features:
- Modern responsive design
- Real-time progress updates
- WebSocket communication
- Mobile-friendly interface

### Silent Mode

Unattended installation using response file:

```bash
# Generate example response file
./ui-example -mode silent -response example.rsp -generate

# Run silent installation
./ui-example -mode silent -response install.rsp -log install.log
```

Example response file (`install.rsp`):
```
accept_license=true
install_type=typical
install_dir=C:\Program Files\MyApp
components=core,docs,examples
desktop_shortcut=true
start_menu=true
add_path=false
confirm_install=true
```

### GUI Mode

Native desktop application (requires Wails):

```bash
./ui-example -mode gui
```

Features:
- Native look and feel
- File dialogs
- System notifications
- Smooth animations
- Dark mode support

### Demo Mode

Run a demonstration of all UI features:

```bash
./ui-example -demo

# With specific mode
./ui-example -demo -mode web
```

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-mode` | UI mode (auto, cli, gui, web, silent) | auto |
| `-title` | Application title | "My Application" |
| `-response` | Response file for silent mode | "" |
| `-log` | Log file path | "" |
| `-port` | Port for web UI | 8080 |
| `-v` | Verbose output | false |
| `-demo` | Run in demo mode | false |

## UI Flow

The standard installation flow:

1. **Welcome Screen** - Introduction and overview
2. **License Agreement** - Display and accept license
3. **Installation Type** - Typical, Custom, Minimal, Complete
4. **Component Selection** - Choose components to install
5. **Installation Directory** - Select target directory
6. **Configuration Options** - Additional settings
7. **Summary & Confirmation** - Review selections
8. **Installation Progress** - Real-time progress display
9. **Completion** - Success or failure message

## Customization

### Themes

The UI supports custom themes:

```go
config := &ui.Config{
    Theme: ui.Theme{
        Name:            "dark",
        PrimaryColor:    "#1a202c",
        SecondaryColor:  "#2d3748",
        BackgroundColor: "#0f1419",
        TextColor:       "#e2e8f0",
        Dark:            true,
    },
}
```

### Custom Components

Add your own components:

```go
components := []ui.Component{
    {
        ID:          "custom",
        Name:        "Custom Component",
        Description: "My custom component",
        Size:        1024 * 1024,
        Children: []ui.Component{
            // Sub-components
        },
    },
}
```

### Custom Options

Define installation options:

```go
options := []ui.Option{
    {
        ID:   "database",
        Name: "Database Type",
        Type: ui.OptionTypeSelect,
        Choices: []ui.Choice{
            {Value: "sqlite", Label: "SQLite"},
            {Value: "postgres", Label: "PostgreSQL"},
            {Value: "mysql", Label: "MySQL"},
        },
        Default: "sqlite",
    },
}
```

## Response File Format

Response files can be in two formats:

### Key-Value Format
```
accept_license=true
install_type=custom
install_dir=/opt/myapp
components=core,plugins,docs
```

### JSON Format
```json
{
    "accept_license": true,
    "install_type": "custom",
    "install_dir": "/opt/myapp",
    "components": ["core", "plugins", "docs"],
    "options": {
        "desktop_shortcut": true,
        "add_path": true
    }
}
```

## Progress Reporting

The UI system provides detailed progress information:

```go
progress := ui.Progress{
    Percentage:    75.5,
    CurrentAction: "Copying files",
    CurrentFile:   "app.exe",
    BytesCompleted: 75497472,
    BytesTotal:     100000000,
    TimeElapsed:   120,
    TimeRemaining: 40,
    Speed:         629145,
}
```

## Error Handling

All UI modes handle errors gracefully:

```go
if err := installer.ShowError(err); err != nil {
    // Error dialog shown to user
}
```

## Platform Differences

### Windows
- Uses Windows color console for CLI
- Native file dialogs in GUI
- Registry integration

### Linux
- ANSI colors in terminal
- GTK dialogs in GUI
- Desktop file creation

### macOS
- Terminal colors
- Native Cocoa dialogs
- App bundle support

## Testing

```bash
# Test CLI mode
go test -run TestCLIUI

# Test Web mode
go test -run TestWebUI

# Test Silent mode
go test -run TestSilentUI

# Test all modes
go test ./...
```

## Troubleshooting

### CLI Mode Issues

**Problem**: No colors in terminal
**Solution**: Check if terminal supports ANSI colors

**Problem**: Input not working
**Solution**: Ensure stdin is available (not redirected)

### Web Mode Issues

**Problem**: Port already in use
**Solution**: Use different port with `-port` flag

**Problem**: Browser doesn't open
**Solution**: Manually navigate to http://localhost:8080

### GUI Mode Issues

**Problem**: GUI not available
**Solution**: Install Wails and rebuild with GUI support

**Problem**: Window doesn't appear
**Solution**: Check display server (X11/Wayland on Linux)

### Silent Mode Issues

**Problem**: Installation fails silently
**Solution**: Check log file specified with `-log`

**Problem**: Response file not found
**Solution**: Use absolute path or check working directory

## Architecture

```
ui/
├── ui.go           # Main interface and types
├── ui_cli.go       # CLI implementation
├── ui_web.go       # Web implementation  
├── ui_gui.go       # GUI implementation (Wails)
├── ui_silent.go    # Silent mode implementation
└── web/            # Web assets
    ├── index.html
    ├── app.js
    └── style.css
```

## Contributing

To add a new UI mode:

1. Implement the `UI` interface
2. Add mode to `Mode` enum
3. Update `CreateUI` factory function
4. Add tests for new mode

## License

MIT License - See LICENSE file in root directory

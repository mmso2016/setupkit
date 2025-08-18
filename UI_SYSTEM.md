# UI System Documentation

## Overview

The Setup-Kit UI system provides a flexible, multi-mode user interface framework for creating professional installation wizards. It supports multiple interface types to accommodate different deployment scenarios and user preferences.

## Architecture

```
┌─────────────────────────────────────────────────┐
│                 UI Interface                    │
├───────────┬──────────┬──────────┬───────────────┤
│    CLI    │   GUI    │   Web    │    Silent     │
├───────────┴──────────┴──────────┴───────────────┤
│              Base UI Components                 │
├─────────────────────────────────────────────────┤
│              Platform Abstraction               │
└─────────────────────────────────────────────────┘
```

## UI Modes

### 1. CLI Mode (Command Line Interface)
- **Use Case**: Server installations, SSH sessions, Docker containers
- **Features**:
  - Interactive prompts
  - Progress bars with ANSI colors
  - Component selection with arrow keys
  - Works over SSH and in containers
- **Requirements**: Terminal with basic ANSI support

### 2. GUI Mode (Graphical User Interface)
- **Use Case**: Desktop applications, end-user software
- **Technology**: Wails v2 framework
- **Features**:
  - Native look and feel
  - System dialogs
  - Smooth animations
  - Dark mode support
- **Requirements**: Display server, Wails runtime

### 3. Web Mode (Browser-based)
- **Use Case**: Remote installations, cloud deployments
- **Features**:
  - Responsive design
  - Real-time updates via WebSocket
  - Mobile-friendly
  - No client installation required
- **Requirements**: Modern web browser

### 4. Silent Mode (Unattended)
- **Use Case**: Automated deployments, CI/CD pipelines
- **Features**:
  - Response file support (text/JSON)
  - Detailed logging
  - Exit codes for scripting
  - No user interaction required
- **Requirements**: Response file with configuration

## Core Components

### UI Interface

```go
type UI interface {
    Initialize(ctx context.Context, config *Config) error
    ShowWelcome() error
    ShowLicense(license string) (bool, error)
    SelectInstallType() (InstallType, error)
    SelectComponents(components []Component) ([]Component, error)
    SelectInstallDirectory(defaultPath string) (string, error)
    ConfigureOptions(options []Option) (map[string]interface{}, error)
    ConfirmInstallation(summary Summary) (bool, error)
    ShowProgress(progress Progress) error
    ShowCompletion(success bool, message string) error
    ShowError(err error) error
    Prompt(message string, defaultValue string) (string, error)
    Confirm(message string, defaultValue bool) (bool, error)
    Close() error
}
```

### Configuration

```go
type Config struct {
    Title        string
    Icon         []byte
    Theme        Theme
    Language     string
    Width        int    // GUI window width
    Height       int    // GUI window height
    WebPort      int    // Web server port
    WebHost      string // Web server host
    ResponseFile string // Silent mode response file
    LogFile      string // Log file path
}
```

### Progress Tracking

```go
type Progress struct {
    Current        int     // Current step
    Total          int     // Total steps
    Percentage     float64 // Overall percentage
    CurrentAction  string  // Current action description
    CurrentFile    string  // Current file being processed
    BytesCompleted int64   // Bytes completed
    BytesTotal     int64   // Total bytes
    TimeElapsed    int     // Seconds elapsed
    TimeRemaining  int     // Estimated seconds remaining
    Speed          int64   // Bytes per second
}
```

## Usage Examples

### Basic Usage

```go
// Create UI with auto-detection
ui, err := ui.CreateUI(ui.ModeAuto, &ui.Config{
    Title: "My Application Installer",
})

// Initialize
ctx := context.Background()
ui.Initialize(ctx, config)

// Show welcome
ui.ShowWelcome()

// Show license
accepted, _ := ui.ShowLicense(licenseText)
if !accepted {
    return
}

// Select components
components := getAvailableComponents()
selected, _ := ui.SelectComponents(components)

// Install with progress
for i := 0; i < 100; i++ {
    ui.ShowProgress(ui.Progress{
        Percentage: float64(i),
        CurrentAction: "Installing...",
    })
}

// Show completion
ui.ShowCompletion(true, "Installation complete!")
```

### Mode-Specific Usage

#### CLI Mode
```go
ui, _ := ui.CreateUI(ui.ModeCLI, config)
// Automatically uses terminal interface
```

#### Web Mode
```go
config := &ui.Config{
    WebPort: 8080,
    WebHost: "0.0.0.0", // Listen on all interfaces
}
ui, _ := ui.CreateUI(ui.ModeWeb, config)
// Opens browser automatically
```

#### Silent Mode
```go
config := &ui.Config{
    ResponseFile: "install.rsp",
    LogFile: "install.log",
}
ui, _ := ui.CreateUI(ui.ModeSilent, config)
// Reads configuration from response file
```

## Response File Format

### Text Format
```
accept_license=true
install_type=typical
install_dir=C:\Program Files\MyApp
components=core,docs,examples
desktop_shortcut=true
add_to_path=true
confirm_install=true
```

### JSON Format
```json
{
    "accept_license": true,
    "install_type": "custom",
    "install_dir": "/opt/myapp",
    "components": ["core", "plugins"],
    "options": {
        "desktop_shortcut": true,
        "add_to_path": true
    }
}
```

## Customization

### Custom Theme

```go
theme := ui.Theme{
    Name:            "dark",
    PrimaryColor:    "#1a202c",
    SecondaryColor:  "#2d3748",
    BackgroundColor: "#0f1419",
    TextColor:       "#e2e8f0",
    Dark:            true,
}

config := &ui.Config{
    Theme: theme,
}
```

### Custom Components

```go
components := []ui.Component{
    {
        ID:          "database",
        Name:        "Database Server",
        Description: "PostgreSQL database server",
        Size:        150 * 1024 * 1024,
        Required:    false,
        Selected:    true,
        Children: []ui.Component{
            {ID: "pgadmin", Name: "pgAdmin"},
            {ID: "samples", Name: "Sample Database"},
        },
    },
}
```

### Custom Options

```go
options := []ui.Option{
    {
        ID:   "port",
        Name: "Server Port",
        Type: ui.OptionTypeNumber,
        Default: 8080,
        Validation: "^[0-9]{1,5}$",
    },
    {
        ID:   "theme",
        Name: "UI Theme",
        Type: ui.OptionTypeSelect,
        Choices: []ui.Choice{
            {Value: "light", Label: "Light"},
            {Value: "dark", Label: "Dark"},
        },
    },
}
```

## Internationalization

```go
// Language support (future enhancement)
config := &ui.Config{
    Language: "de-DE", // German
}

// Provide translations
translations := map[string]map[string]string{
    "de-DE": {
        "welcome": "Willkommen",
        "next": "Weiter",
        "back": "Zurück",
    },
}
```

## Error Handling

All UI implementations handle errors gracefully:

```go
err := ui.ShowError(fmt.Errorf("disk space insufficient"))
// Shows appropriate error dialog/message

// Check for user cancellation
if err == ui.ErrUserCancelled {
    // Handle cancellation
}
```

## Testing

### Unit Tests
```go
func TestUIFlow(t *testing.T) {
    ui := NewMockUI()
    
    // Set expected responses
    ui.SetResponse("license", true)
    ui.SetResponse("components", []string{"core", "docs"})
    
    // Run installation flow
    err := runInstallation(ui)
    assert.NoError(t, err)
}
```

### Integration Tests
```go
func TestCLIIntegration(t *testing.T) {
    ui, _ := ui.CreateUI(ui.ModeCLI, config)
    // Test with actual terminal
}
```

## Platform Considerations

### Windows
- CLI: Windows Terminal or PowerShell for best color support
- GUI: Native Windows look with WinUI styles
- Paths: Use backslashes, Program Files detection

### Linux
- CLI: Full ANSI color support
- GUI: GTK theme integration
- Paths: Follow FHS (Filesystem Hierarchy Standard)

### macOS
- CLI: Terminal.app compatibility
- GUI: Native Cocoa widgets
- Paths: /Applications for GUI apps

## Performance

- **CLI**: Minimal overhead, instant response
- **GUI**: ~30MB memory usage with Wails
- **Web**: ~5MB server memory + browser resources
- **Silent**: Fastest mode, no UI overhead

## Security

- **Input Validation**: All user inputs are validated
- **Path Traversal**: Protected against directory traversal attacks
- **XSS Protection**: Web mode sanitizes all outputs
- **Privilege Escalation**: Proper elevation handling per platform

## Best Practices

1. **Mode Selection**:
   - Use Auto mode for maximum compatibility
   - Provide fallback options
   - Document mode requirements

2. **Progress Reporting**:
   - Update at meaningful intervals (not too frequent)
   - Provide estimated time remaining
   - Show current action clearly

3. **Error Messages**:
   - Be specific about what went wrong
   - Suggest solutions when possible
   - Log detailed errors for debugging

4. **Response Files**:
   - Provide example files
   - Validate all values
   - Support environment variables

5. **Accessibility**:
   - Support keyboard navigation
   - Provide clear labels
   - Use sufficient color contrast

## Troubleshooting

### Common Issues

1. **CLI colors not working**
   - Check terminal capabilities
   - Try setting `TERM=xterm-256color`

2. **Web UI not opening**
   - Check firewall settings
   - Verify port availability
   - Try manual browser navigation

3. **GUI not starting**
   - Verify display server (X11/Wayland)
   - Check Wails installation
   - Review system logs

4. **Silent mode failing**
   - Check response file syntax
   - Review log file for errors
   - Validate file paths

## Future Enhancements

- [ ] Accessibility improvements (screen reader support)
- [ ] More themes and customization options
- [ ] Plugin system for custom UI components
- [ ] Mobile app support
- [ ] Cloud-based installation management
- [ ] Real-time collaboration features
- [ ] AI-assisted configuration

## Contributing

See CONTRIBUTING.md for guidelines on adding new UI modes or enhancing existing ones.

## License

MIT License - See LICENSE file

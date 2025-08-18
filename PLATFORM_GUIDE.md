# Platform-Specific Implementation Guide

## Overview

The Setup-Kit provides comprehensive platform-specific implementations for Windows, Linux, and macOS. This guide covers the features and usage for each platform.

## Supported Platforms

| Platform | Architecture | Service Management | PATH Management | Elevation | Registry |
|----------|-------------|-------------------|-----------------|-----------|----------|
| Windows  | amd64, 386  | ✅ Windows Services | ✅ Registry-based | ✅ UAC | ✅ Full |
| Linux    | amd64, 386, arm, arm64 | ✅ systemd | ✅ Shell profiles | ✅ sudo/pkexec | ❌ N/A |
| macOS    | amd64, arm64 | ✅ launchd | ✅ Shell profiles | ✅ osascript | ❌ N/A |

## Windows Features

### Service Management
- Full Windows Service API integration
- Alternative sc.exe command support
- Service recovery and restart policies
- Event log integration

```go
// Example: Install Windows service
config := &ServiceConfig{
    Name:        "MyService",
    DisplayName: "My Application Service",
    Description: "Runs My Application as a Windows service",
    Executable:  "C:\\Program Files\\MyApp\\myapp.exe",
    AutoStart:   true,
    RestartMode: RestartOnFailure,
}

manager := GetServiceManager(logger)
err := manager.Install(config)
```

### Registry Operations
```go
// Write to registry
installer.WriteRegistryString("HKLM", 
    "SOFTWARE\\MyCompany\\MyApp", 
    "Version", "1.0.0")

// Read from registry
value, err := installer.ReadRegistryString("HKLM",
    "SOFTWARE\\MyCompany\\MyApp",
    "Version")
```

### UAC Elevation
```go
if installer.RequiresElevation() && !installer.IsElevated() {
    err := installer.RequestElevation() // Restarts with admin rights
}
```

### PATH Management
- System PATH via registry (requires admin)
- User PATH via registry
- Automatic environment broadcast

## Linux Features

### systemd Service Management
- Generates systemd unit files
- Supports all systemd directives
- Journal integration for logging

```go
// Example: Install systemd service
config := &ServiceConfig{
    Name:        "myapp",
    Description: "My Application Service",
    Executable:  "/usr/local/bin/myapp",
    User:        "myapp",
    WorkingDir:  "/var/lib/myapp",
    Environment: map[string]string{
        "APP_ENV": "production",
    },
    RestartMode: RestartAlways,
    LimitNOFILE: 65536,
}
```

### PATH Management
- System: `/etc/profile.d/` scripts
- User: `.bashrc`, `.zshrc`, `.profile`
- Automatic shell detection

### Privilege Elevation
- sudo support with fallback to pkexec
- Automatic detection of elevation requirements

## macOS Features

### launchd Service Management
- Generates launchd plist files
- Supports both LaunchDaemons and LaunchAgents
- Automatic domain detection (system/user)

```go
// Example: Install launchd service
config := &ServiceConfig{
    Name:        "com.mycompany.myapp",
    Description: "My Application",
    Executable:  "/Applications/MyApp.app/Contents/MacOS/MyApp",
    AutoStart:   true,
    RestartMode: RestartAlways,
}
```

### App Bundle Support
- Creates proper .app bundle structure
- Info.plist generation
- Launch Services registration

### PATH Management
- System: `/etc/paths.d/` files
- User: `.zshrc`, `.bash_profile`
- Automatic shell detection (zsh default since Catalina)

## Cross-Platform Considerations

### Build Tags
Each platform-specific file uses appropriate build tags:

```go
//go:build windows
// +build windows

//go:build linux
// +build linux

//go:build darwin
// +build darwin

//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin
```

### Interface Consistency
All platforms implement the same interfaces:

```go
type PlatformInstaller interface {
    Initialize() error
    CheckRequirements() error
    InstallService(name, path string) error
    CreateShortcuts() error
    RegisterWithOS() error
    RegisterUninstaller() error
    // ... elevation methods
    // ... PATH methods
    // ... environment methods
}

type ServiceManager interface {
    Install(config *ServiceConfig) error
    Uninstall(name string) error
    Start(name string) error
    Stop(name string) error
    Status(name string) (ServiceStatus, error)
    // ... other service methods
}
```

### Error Handling
Platform-specific operations return `ErrNotSupported` when not available:

```go
// On Linux/macOS
err := installer.WriteRegistryString(...) // Returns ErrNotSupported

// Check for support
if err == ErrNotSupported {
    // Use alternative method
}
```

## Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Platform-specific tests
GOOS=windows go test ./...
GOOS=linux go test ./...
GOOS=darwin go test ./...
```

### Integration Tests
Require elevated privileges:

```bash
# Windows (as Administrator)
go test -tags=integration ./...

# Linux/macOS (as root)
sudo go test -tags=integration ./...
```

### Cross-Compilation Testing
```bash
# Build for all platforms
make cross-compile

# Or manually
GOOS=windows GOARCH=amd64 go build ./...
GOOS=linux GOARCH=amd64 go build ./...
GOOS=darwin GOARCH=amd64 go build ./...
```

## Best Practices

### 1. Always Check Platform Support
```go
if runtime.GOOS == "windows" {
    // Windows-specific code
}

// Or use interface methods
err := installer.SomeMethod()
if err == ErrNotSupported {
    // Handle unsupported platform
}
```

### 2. Handle Elevation Gracefully
```go
if installer.RequiresElevation() {
    if installer.CanElevate() {
        fmt.Println("This operation requires administrator privileges.")
        if confirmFromUser() {
            installer.RequestElevation()
        }
    } else {
        fmt.Println("Please run as administrator/root")
        os.Exit(1)
    }
}
```

### 3. Use Appropriate Service Names
- Windows: `MyAppService` or `MyCompany.MyApp`
- Linux: `myapp` or `myapp.service`
- macOS: `com.mycompany.myapp` (reverse domain notation)

### 4. Test on Target Platforms
Always test on actual target platforms, not just cross-compile:
- Windows: Test UAC, service installation, registry
- Linux: Test different distributions, systemd versions
- macOS: Test different macOS versions, app signing

### 5. Provide Fallbacks
```go
// Try native method first
err := manager.Install(config)
if err != nil {
    // Fall back to alternative method
    if runtime.GOOS == "windows" {
        err = installWithSC(config) // Use sc.exe
    }
}
```

## Troubleshooting

### Windows Issues

**Problem**: "Access denied" when installing service
**Solution**: Run as Administrator

**Problem**: PATH changes not visible immediately
**Solution**: Environment broadcast is sent, but some applications need restart

### Linux Issues

**Problem**: systemd service fails to start
**Solution**: Check `journalctl -u servicename -f` for logs

**Problem**: PATH not updated in current shell
**Solution**: Source the profile: `source ~/.bashrc`

### macOS Issues

**Problem**: launchd service not starting
**Solution**: Check Console.app for error messages

**Problem**: App not appearing in Launchpad
**Solution**: Register with Launch Services: `lsregister -f /path/to/app`

## Examples

See the `examples/` directory for complete working examples:
- `basic/` - Simple installer with all platform features
- Service installation examples
- PATH management examples
- Elevation handling examples

## Contributing

When adding platform-specific features:
1. Implement the interface method for all platforms
2. Return `ErrNotSupported` for unsupported platforms
3. Add appropriate build tags
4. Include tests for each platform
5. Update this documentation

## License

MIT License - See LICENSE file for details

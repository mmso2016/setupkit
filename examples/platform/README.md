# Platform-Specific Example

This example demonstrates all platform-specific features of the Setup-Kit library.

## Features Demonstrated

- **Service Management**: Install/uninstall system services (Windows Services, systemd, launchd)
- **PATH Management**: Add/remove directories from system or user PATH
- **Elevation Handling**: Request administrator/root privileges when needed
- **Registry Operations** (Windows): Read/write registry values
- **Environment Variables**: Set/unset environment variables
- **OS Registration**: Register with Add/Remove Programs (Windows), create desktop entries (Linux)
- **Shortcuts**: Create Start Menu shortcuts (Windows), desktop entries (Linux), app bundles (macOS)

## Building

```bash
# Build for current platform
go build -o platform-example

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o platform-example.exe

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o platform-example-linux

# Cross-compile for macOS
GOOS=darwin GOARCH=amd64 go build -o platform-example-mac
```

## Usage

### Installation

```bash
# Basic installation
./platform-example -install

# Install with custom directory
./platform-example -install -dir /opt/myapp

# Install with service
./platform-example -install -service

# Install with custom service name
./platform-example -install -service -service-name myapp-service

# Install without adding to PATH
./platform-example -install -path=false

# Verbose installation
./platform-example -install -v
```

### Uninstallation

```bash
# Basic uninstallation
./platform-example -uninstall

# Uninstall including service
./platform-example -uninstall -service

# Uninstall with custom service name
./platform-example -uninstall -service -service-name myapp-service
```

## Platform-Specific Behavior

### Windows

- Requests UAC elevation when installing to Program Files
- Creates Windows service using Windows Service API
- Adds to PATH via registry
- Registers in Add/Remove Programs
- Creates Start Menu shortcuts

**Run as Administrator:**
```cmd
platform-example.exe -install -service
```

### Linux

- Requests sudo/pkexec elevation for system directories
- Creates systemd service
- Adds to PATH via /etc/profile.d or shell RC files
- Creates .desktop files for desktop environments

**Run with sudo:**
```bash
sudo ./platform-example -install -service
```

### macOS

- Requests osascript elevation for system directories
- Creates launchd service
- Adds to PATH via /etc/paths.d or shell profiles
- Creates app bundles and registers with Launch Services

**Run with sudo:**
```bash
sudo ./platform-example -install -service
```

## Testing Service Installation

After installing with `-service` flag:

### Windows
```cmd
# Check service status
sc query gosetupkit-example

# Start/stop service
sc start gosetupkit-example
sc stop gosetupkit-example

# View in Services Manager
services.msc
```

### Linux
```bash
# Check service status
systemctl status gosetupkit-example

# Start/stop service
sudo systemctl start gosetupkit-example
sudo systemctl stop gosetupkit-example

# View logs
journalctl -u gosetupkit-example -f
```

### macOS
```bash
# Check service status
launchctl list | grep gosetupkit

# Start/stop service
sudo launchctl start com.gosetupkit.gosetupkit-example
sudo launchctl stop com.gosetupkit.gosetupkit-example

# View logs
tail -f /var/log/gosetupkit-example.log
```

## Testing PATH Management

After installation with PATH enabled:

### Windows
```cmd
# Check if in PATH (new command prompt required)
echo %PATH%

# Or check registry
reg query "HKCU\Environment" /v Path
```

### Linux/macOS
```bash
# Check if in PATH (new shell required)
echo $PATH

# Or source profile
source ~/.bashrc  # Linux
source ~/.zshrc   # macOS
```

## Troubleshooting

### Permission Denied
- **Windows**: Right-click and "Run as Administrator"
- **Linux/macOS**: Use `sudo` prefix

### Service Installation Failed
- Check if you have sufficient privileges
- Check if service name already exists
- View system logs for detailed error

### PATH Not Updated
- Open a new terminal/command prompt
- On Linux/macOS, source your shell profile
- On Windows, restart applications or log out/in

### Service Not Starting
- Check service logs
- Verify executable path is correct
- Check file permissions

## Code Structure

The example demonstrates:

1. **Configuration Setup**: Creating installer config
2. **Platform Detection**: Getting appropriate platform installer
3. **Elevation Handling**: Checking and requesting privileges
4. **Installation Process**:
   - Create directories
   - Copy files
   - PATH management
   - Service installation
   - Shortcut creation
   - OS registration
5. **Uninstallation Process**:
   - Service removal
   - PATH cleanup
   - File removal
   - Registry cleanup (Windows)

## Extending the Example

To adapt for your application:

1. Replace `createDummyExecutable()` with actual file copying
2. Modify service configuration in `performInstallation()`
3. Add custom registry values or environment variables
4. Implement proper component selection
5. Add configuration file handling

## License

MIT License - See LICENSE file in the root directory

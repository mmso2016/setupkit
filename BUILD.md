# Setup-Kit Build System

## Overview

The Setup-Kit project uses both **Make** and **Mage** for building. Both tools provide the same functionality, so you can use whichever you prefer.

## Prerequisites

- Go 1.21+
- Make (optional, for Makefile)
- Mage (optional, for Magefile)
- Wails v2 (optional, for GUI)

## Installing Build Tools

### Install Mage
```bash
go install github.com/magefile/mage@latest
```

### Install Wails (for GUI)
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Build Tags

The project supports several build tags to control which features are included:

### Available Tags
- **`wails`** - Enable Wails GUI support
- **`nogui`** - Disable all GUI features (CLI/Silent only)
- **`nocli`** - Disable CLI support (not recommended)

### Build Tag Matrix

| Datei | Build-Tags | Funktion |
|-------|------------|----------|
| `factory_gui.go` | `!wails && !nogui` | Standard-GUI (CLI-Fallback) |
| `factory_wails.go` | `wails` | Wails-basierte GUI |
| `factory_gui_stub.go` | `nogui` | GUI-Unterstützung deaktiviert |
| `factory_cli.go` | `!nocli` | CLI-Unterstützung |

### Build Variants
```bash
# Using Make
make build-ui           # Standard build (CLI fallback)
make build-ui-wails     # With Wails GUI support
make build-ui-nogui     # CLI/Silent only

# Using Mage
mage buildUI            # Standard build
mage buildUIWails       # With Wails support
mage buildUINoGUI       # CLI/Silent only
```

### Manual Build with Tags
```bash
# Standard build (no tags) - CLI fallback
go build ./examples/ui

# With Wails support
go build -tags wails ./examples/ui

# CLI only (smallest binary)
go build -tags nogui ./examples/ui
```

### Development Workflow for Different Build Types

#### For Regular CLI Development:
```bash
make build-cli
# or
mage buildCLI
```

#### For GUI Development with Wails:
```bash
# Install Wails (if not already installed)
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build GUI
make build-gui
# or
mage buildGUI

# Development mode with hot reload
mage dev
```

#### For Platform-specific Development:
```bash
make build-platform
# or
mage buildPlatform
```

### Recommended Distribution Builds

#### Minimal CLI Version:
```bash
go build -tags nogui -ldflags "-s -w" ./examples/ui
```

#### Full GUI Version:
```bash
go build -tags wails -ldflags "-s -w" ./examples/ui
```

#### Universal Version (with fallback):
```bash
go build ./examples/ui
```

## Using Make

### Basic Commands
```bash
# Show help
make help

# Build everything
make all

# Build specific targets
make build-cli      # Build CLI installer
make build-console  # Build console GUI
make build-gui      # Build Wails GUI (requires Wails)
make build-platform # Build platform example

# Testing
make test           # Run tests
make test-verbose   # Run tests with verbose output
make test-race      # Run tests with race detector
make coverage       # Generate coverage report

# Code quality
make fmt            # Format code
make vet            # Run go vet
make lint           # Run linters (requires golangci-lint)

# Clean
make clean          # Remove build artifacts

# Dependencies
make deps           # Download dependencies
make tidy           # Tidy go.mod
```

## Using Mage

### Basic Commands
```bash
# List all targets
mage -l

# Build everything
mage all

# Build specific targets
mage buildCLI       # Build CLI installer
mage buildConsole   # Build console GUI
mage buildGUI       # Build Wails GUI
mage buildPlatform  # Build platform example

# Testing
mage test           # Run tests
mage testVerbose    # Run tests with verbose output
mage testRace       # Run tests with race detector
mage bench          # Run benchmarks
mage coverage       # Generate coverage report

# Code quality
mage fmt            # Format code
mage vet            # Run go vet
mage lint           # Run linters

# Clean
mage clean          # Remove build artifacts

# Dependencies
mage deps           # Download dependencies
mage tidy           # Tidy go.mod
mage verify         # Verify dependencies

# Run examples
mage runCLI         # Build and run CLI
mage runConsole     # Build and run console GUI
mage runGUI         # Build and run Wails GUI

# Development
mage dev            # Start Wails in development mode

# Utilities
mage version        # Show version information
mage wailsInstall   # Install Wails CLI
mage wailsDoctor    # Run Wails doctor
```

## Build Output

All binaries are placed in the `bin/` directory:
- `bin/installer-cli[.exe]` - CLI installer
- `bin/installer-console[.exe]` - Console GUI
- `bin/installer-gui[.exe]` - Wails GUI
- `bin/installer-platform[.exe]` - Platform example
- `bin/installer-ui[.exe]` - UI example (standard)
- `bin/installer-ui-wails[.exe]` - UI example with Wails
- `bin/installer-ui-nogui[.exe]` - UI example without GUI

## Testing

### Run All Tests
```bash
# Using Make
make test

# Using Mage
mage test
```

### Test All Build Configurations
```bash
# Using Make (requires Mage)
make test-builds

# Using Mage directly
mage testAllBuilds
mage testBuildTags
mage validateBuildTags
```

### Test UI Configuration System
```bash
# Using Make
make test-configs

# Using Mage
mage testConfigs
```

The configuration tests will:
- Test all built-in themes
- Verify config file generation
- Validate example configurations
- Test theme and config loading

The `testAllBuilds` target will:
- Test all build configurations (CLI, Console, Platform, UI variants)
- Report success/failure for each build
- List all created binaries
- Test specific build tag combinations

### Generate Coverage Report
```bash
# Using Make
make coverage

# Using Mage
mage coverage
```

The coverage report will be generated in `coverage/coverage.html`.

## Development Workflow

### 1. Make Changes
Edit your code in your favorite editor.

### 2. Format and Check
```bash
# Using Make
make fmt vet

# Using Mage
mage fmt vet
```

### 3. Run Tests
```bash
# Using Make
make test

# Using Mage
mage test
```

### 4. Build
```bash
# Using Make
make build

# Using Mage
mage build
```

### 5. Run
```bash
# Direct execution
./bin/installer-cli --help
./bin/installer-console

# Or using Mage
mage runCLI
mage runConsole
```

## GUI Development

### Development Mode
```bash
# Using Make
make dev

# Using Mage
mage dev
```

This starts Wails in development mode with hot reload.

### Build GUI
```bash
# Using Make
make build-gui

# Using Mage
mage buildGUI
```

## Cross-Platform Notes

Both Makefile and Magefile are designed to work on:
- Windows
- Linux
- macOS

The build system automatically detects the platform and adjusts paths and binary extensions accordingly.

## Troubleshooting

### Build Tag Related Issues

#### "GUI support not compiled in"
- Wails not available or `nogui` tag used
- Solution: Install Wails or use `build-ui-wails` target

#### Build Errors with Tags
- Check tag syntax: `-tags "tag1,tag2"`
- No spaces between tags in comma-separated list
- Ensure build tags are correctly specified in source files

### Wails-Specific Issues

#### Wails Build Errors
- Check Wails installation: `wails doctor`
- Install/update frontend dependencies: `cd examples/gui && npm install`
- Ensure Wails CLI is up to date: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### General Build Issues

### Wails Not Found
```bash
# Install Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Or using Mage
mage wailsInstall

# Check Wails installation
wails doctor

# Or using Mage
mage wailsDoctor
```

### Tests Failing with File Lock
Make sure all loggers are properly closed. The logger implementation now includes a `Close()` method that must be called:
```go
logger := core.NewLogger("info", "logfile.log")
defer logger.Close() // Important!
```

### Build Fails
```bash
# Clean and rebuild
make clean all

# Or using Mage
mage clean all

# Fix dependencies
go mod tidy
go mod download
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Install Mage
      run: go install github.com/magefile/mage@latest
    
    - name: Build
      run: mage build
    
    - name: Test
      run: mage test
```

## License

MIT

# SetupKit UI - Wails Installer Application

This is a full-featured Wails-based installer GUI for the SetupKit project.

## Features

- Modern, responsive GUI with step-by-step wizard
- Multiple theme support (Default, Corporate Blue, Medical Green, Tech Dark, Minimal Light)
- Component selection with size calculation
- Installation progress tracking
- Browse for installation directory
- License agreement display
- Installation summary

## Building

### Prerequisites

1. Install Wails:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

2. Install Node.js (for frontend build)

### Build Commands

#### Build the Wails GUI:
```bash
# From this directory (examples/ui)
wails build

# The executable will be in build/bin/
```

#### Development Mode:
```bash
# Run in development mode with hot reload
wails dev
```

#### Build for specific platform:
```bash
# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/universal

# Linux
wails build -platform linux/amd64
```

### Command Line Options

The installer supports various command-line flags:

```bash
# Run with specific theme
./installer-ui -theme corporate-blue

# Specify application details
./installer-ui -title "My App" -version "2.0.0" -publisher "My Company"

# List available themes
./installer-ui -list-themes

# Generate configuration file
./installer-ui -generate-config myconfig.yaml

# Use configuration file
./installer-ui -config myconfig.yaml

# CLI mode fallback (when not using GUI)
./installer-ui -mode cli
```

## Project Structure

```
examples/ui/
├── main.go              # Main Wails application
├── wails.json           # Wails configuration
├── frontend/            # Frontend files
│   ├── package.json     # Frontend build configuration
│   ├── src/             # Source files
│   │   ├── index.html   # Main HTML
│   │   ├── style.css    # Styles with theme support
│   │   └── app.js       # Frontend logic
│   └── dist/            # Built frontend files
└── build/               # Build output (created by Wails)
    └── bin/             # Compiled executables
```

## Themes

The installer includes 5 built-in themes:

1. **Default** - Purple gradient theme
2. **Corporate Blue** - Professional blue theme
3. **Medical Green** - Healthcare-oriented green theme
4. **Tech Dark** - Dark theme for technical applications
5. **Minimal Light** - Clean, minimal light theme

Themes can be changed at runtime from the welcome screen or specified via command line.

## Development

### Modifying the Frontend

1. Edit files in `frontend/src/`
2. Run `npm run build` in the frontend directory to copy to dist
3. Run `wails build` to rebuild the application

### Adding New Features

The main.go file contains the backend logic with these key methods:

- `GetConfig()` - Returns configuration for the frontend
- `BrowseFolder()` - Opens directory selection dialog
- `SetSelectedComponents()` - Updates component selection
- `SetInstallPath()` - Sets installation directory
- `StartInstallation()` - Begins the installation process
- `FinishInstallation()` - Completes installation and optionally launches the app

## Troubleshooting

### Build Issues

If you encounter build issues:

1. Ensure Wails is properly installed:
```bash
wails doctor
```

2. Clear the build cache:
```bash
wails build -clean
```

3. Ensure frontend files are in dist:
```bash
cd frontend
npm run build
```

### Runtime Issues

- If the GUI doesn't appear, check that you're running in GUI mode (default)
- For debugging, run with verbose flag: `-v`
- Check the console output for any error messages

## License

MIT License - See the main project LICENSE file for details.

# Contributing to Setup-Kit

Thank you for your interest in contributing to Setup-Kit! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and constructive in all interactions. We aim to create a welcoming environment for all contributors.

## How to Contribute

### Reporting Issues

- Check if the issue already exists
- Include a clear description of the problem
- Provide steps to reproduce the issue
- Include your OS, Go version, and relevant configuration
- Add code samples or error messages if applicable

### Suggesting Features

- Check if the feature has already been requested
- Clearly describe the use case
- Explain why this would be valuable for the project
- Consider providing a rough implementation plan

### Submitting Pull Requests

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow the coding standards below
   - Add tests for new functionality
   - Update documentation as needed

4. **Test your changes**
   ```bash
   mage test
   mage lint
   ```

5. **Commit with meaningful messages**
   ```bash
   git commit -m "Add feature: description of what you added"
   ```

6. **Push and create a Pull Request**
   - Provide a clear description of the changes
   - Reference any related issues
   - Include examples if applicable

## Development Setup

### Prerequisites

- Go 1.21 or later
- Mage (install with `go install github.com/magefile/mage@latest`)
- golangci-lint (optional, for linting)

### Building

```bash
# Build for current platform
mage build

# Build for all platforms
mage buildAll

# Run tests
mage test

# Run linter
mage lint
```

## Coding Standards

### Go Code Style

- Follow standard Go formatting (use `gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Handle errors appropriately

### Project Structure

```
Setup-Kit/
├── installer/          # Core library code
│   ├── core/          # Core functionality
│   ├── platform/      # Platform-specific implementations
│   └── ui/            # User interface components
├── cmd/               # Command-line applications
├── examples/          # Example implementations
├── docs/              # Documentation
└── tests/             # Integration tests
```

### Testing

- Write unit tests for new functionality
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Test error conditions

Example test:

```go
func TestExtractAssets(t *testing.T) {
    tests := []struct {
        name    string
        assets  embed.FS
        target  string
        wantErr bool
    }{
        {
            name:    "valid extraction",
            assets:  testAssets,
            target:  t.TempDir(),
            wantErr: false,
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ExtractAssets(tt.assets, tt.target)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExtractAssets() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Add GoDoc comments for exported functions
- Include examples in documentation
- Keep documentation concise and clear

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions or changes
- `chore`: Maintenance tasks

Examples:
```
feat(installer): add rollback support for failed installations
fix(windows): correct service installation on Windows 11
docs(examples): add enterprise deployment example
```

## Platform-Specific Guidelines

### Windows

- Test on Windows 10 and 11
- Ensure UAC elevation works correctly
- Verify registry operations
- Test service installation

### Linux

- Test on major distributions (Ubuntu, Fedora, Debian)
- Verify systemd integration
- Check package manager compatibility
- Test with different privilege levels

### macOS

- Test on recent macOS versions
- Verify code signing requirements
- Test launchd integration
- Check Gatekeeper compatibility

## Release Process

1. Update version numbers
2. Update CHANGELOG.md
3. Run full test suite
4. Build release artifacts
5. Create GitHub release
6. Update documentation

## Getting Help

- Check existing documentation
- Look through existing issues
- Ask questions in discussions
- Join our community chat (if available)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:
- The AUTHORS file
- Release notes
- Project documentation

Thank you for contributing to Setup-Kit!

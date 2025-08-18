# Setup-Kit Templates & Examples

A curated list of real-world projects using Setup-Kit or serving as installation templates.

## 🎯 Official Templates

### comming soon
- **Repository**: ... 
- **Description**: ...
- **Features**: 
  - show use of SetupKit
- **Use Case**: starting point for own projects

## 📦 Community Projects

### Your Project Here!
*Submit a PR to add your project that uses Setup-Kit*

## 🔧 Installer Patterns by Category

### Web Services
| Project | Description | Platform Support | Key Features |
|---------|-------------|-----------------|--------------|
| *pending* | | | |

### CLI Tools
| Project | Description | Platform Support | Key Features |
|---------|-------------|-----------------|--------------|
| [example-cli](examples/simple-cli) | Basic CLI tool installer | Win/Linux/Mac | PATH integration, minimal deps |

### Desktop Applications
| Project | Description | Platform Support | Key Features |
|---------|-------------|-----------------|--------------|
| *pending* | | | |

### Database Applications
| Project | Description | Platform Support | Key Features |
|---------|-------------|-----------------|--------------|
| *pending* | | | |

## 🌟 Inspirational Projects
*These projects don't use Setup-Kit but demonstrate excellent installation patterns*


### Installation Frameworks
- [**WiX Toolset**](https://github.com/wixtoolset/wix3) - Windows Installer XML toolset
- [**fpm**](https://github.com/jordansissel/fpm) - Multi-format package building
- [**nfpm**](https://github.com/goreleaser/nfpm) - Go-based rpm/deb/apk packager
- [**GoReleaser**](https://github.com/goreleaser/goreleaser) - Release automation tool

## 📝 Best Practices from Templates

### Binary Embedding
```go
//go:embed assets/* postgresql/bin/* web/dist/*
var embeddedFiles embed.FS
```

### Service Installation
- **Windows**: Use `sc.exe` or Windows Service API
- **Linux**: Generate systemd unit files
- **macOS**: Create launchd plists

### Unattended Installation
```yaml
# response.yaml example from enterprise templates
installation:
  mode: unattended
  accept_license: true
  components:
    - core
    - database
    - web
  service:
    auto_start: true
    run_as: "LOCAL_SYSTEM"
```

### Platform Detection
```go
switch runtime.GOOS {
case "windows":
    return installWindowsService()
case "linux":
    return installSystemdService()
case "darwin":
    return installLaunchdService()
}
```

## 🤝 Contributing a Template

To add your project to this list:

1. Use Setup-Kit in your project
2. Ensure your installer demonstrates good practices
3. Submit a PR with:
   - Project link
   - Brief description
   - Key features
   - Platform support

## 📚 Additional Resources

- [Installer Best Practices](docs/best-practices.md) *(coming soon)*
- [Platform-Specific Guides](docs/platforms/) *(coming soon)*
- [Security Considerations](docs/security.md) *(coming soon)*

---

*Last updated: August 2025*

# SetupKit Makefile - Embedded Architecture
# Single-file installer with embedded configuration and assets

# Variables
MODULE := github.com/mmso2016/setupkit
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GOFLAGS := -v

# Directories
BIN_DIR := bin
INSTALLER_DIR := examples/installer-demo
CUSTOM_STATE_DEMO_DIR := examples/custom-state-demo

# Platform detection
ifeq ($(OS),Windows_NT)
    BINARY_EXT := .exe
    RM := del /Q
    MKDIR := mkdir
    PATH_SEP := \\
else
    BINARY_EXT :=
    RM := rm -f
    MKDIR := mkdir -p
    PATH_SEP := /
endif

# Default target
.PHONY: all
all: clean test build

# Help target
.PHONY: help
help:
	@echo "SetupKit - Modern Installer Framework"
	@echo "Embedded Architecture - Single-file installer"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build Targets:"
	@echo "  all          - Clean, test and build embedded installer"
	@echo "  build        - Build embedded installer with all assets"
	@echo "  build-custom-state-demo - Build custom state demo"
	@echo "  build-all    - Build all demo applications"
	@echo "  clean        - Clean build artifacts"
	@echo ""
	@echo "Test Targets:"
	@echo "  test         - Run all tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  test-race    - Run tests with race detector"
	@echo "  coverage     - Generate test coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run linters"
	@echo ""
	@echo "Run Targets:"
	@echo "  run          - Run embedded installer (auto mode)"
	@echo "  run-gui      - Run installer (GUI mode)"
	@echo "  run-cli      - Run installer (CLI mode)"
	@echo "  run-silent   - Run installer (silent mode)"
	@echo "  run-custom-state-demo - Run database configuration demo (silent mode)"
	@echo "  run-custom-state-demo-cli - Run database configuration demo (CLI mode)"
	@echo "  run-custom-state-demo-auto - Run database configuration demo (auto mode)"
	@echo "  help-custom-state-demo - Show custom state demo help and options"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy go.mod"
	@echo "  verify       - Verify dependencies"
	@echo ""
	@echo "Utilities:"
	@echo "  clean-install - Clean test installation directories"
	@echo "  version      - Show version information"
	@echo "  help         - Show this help"

# Build target
.PHONY: build
build:
	@echo "Building SetupKit embedded installer..."
	@echo "All configuration and assets are embedded in the executable."
	@echo ""
	@$(MKDIR) $(BIN_DIR)
	go build $(GOFLAGS) -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)" -o $(BIN_DIR)$(PATH_SEP)setupkit-installer-demo$(BINARY_EXT) $(INSTALLER_DIR)
	@echo "✅ Embedded installer built: $(BIN_DIR)$(PATH_SEP)setupkit-installer-demo$(BINARY_EXT)"
	@echo "✅ Single-file installer ready - no external dependencies needed!"

# Build custom state demo
.PHONY: build-custom-state-demo
build-custom-state-demo:
	@echo "Building SetupKit custom state demo..."
	@echo "Demonstrates database configuration custom state."
	@echo ""
	@$(MKDIR) $(BIN_DIR)
	go build $(GOFLAGS) -o $(BIN_DIR)$(PATH_SEP)setupkit-custom-state-demo$(BINARY_EXT) $(CUSTOM_STATE_DEMO_DIR)
	@echo "✅ Custom state demo built: $(BIN_DIR)$(PATH_SEP)setupkit-custom-state-demo$(BINARY_EXT)"
	@echo "✅ Database configuration demo ready!"

# Build all demo applications
.PHONY: build-all
build-all: build build-custom-state-demo

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	go test -short ./pkg/...

.PHONY: test-verbose
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./pkg/...

.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	go test -race -short ./pkg/...

.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	@$(MKDIR) coverage
	go test -coverprofile=coverage/coverage.out ./pkg/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report: coverage/coverage.html"

# Code quality targets
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not found. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

# Dependency management
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

.PHONY: tidy
tidy:
	@echo "Tidying go.mod..."
	go mod tidy

.PHONY: verify
verify:
	@echo "Verifying dependencies..."
	go mod verify

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@$(RM) -rf $(BIN_DIR)
	@$(RM) -rf coverage
	@find . -name "*.test" -delete 2>/dev/null || true
	@find . -name "*.out" -delete 2>/dev/null || true
	@echo "✅ Cleanup complete!"

# Run targets - embedded installer
.PHONY: run
run: build
	@echo "Running embedded installer (auto mode)..."
	@echo "Uses embedded configuration and assets"
	@echo ""
	$(BIN_DIR)$(PATH_SEP)setupkit-installer-demo$(BINARY_EXT)

.PHONY: run-gui
run-gui: build
	@echo "Starting embedded installer (GUI mode)..."
	@echo "Opens browser-based interface with embedded assets"
	@echo ""
	$(BIN_DIR)$(PATH_SEP)setupkit-installer-demo$(BINARY_EXT) -mode=gui

.PHONY: run-cli
run-cli: build
	@echo "Starting embedded installer (CLI mode)..."
	@echo "Interactive CLI with embedded configuration"
	@echo ""
	$(BIN_DIR)$(PATH_SEP)setupkit-installer-demo$(BINARY_EXT) -mode=cli

.PHONY: run-silent
run-silent: build
	@echo "Starting embedded installer (silent mode)..."
	@echo "Unattended installation with embedded assets"
	@echo ""
	$(BIN_DIR)$(PATH_SEP)setupkit-installer-demo$(BINARY_EXT) -silent -profile=minimal

# Run custom state demo
.PHONY: run-custom-state-demo
run-custom-state-demo: build-custom-state-demo
	@echo "Starting custom state demo (database configuration)..."
	@echo "Demonstrates: Welcome → License → Components → Install Path → DB Config → Summary → Complete"
	@echo ""
	$(BIN_DIR)$(PATH_SEP)setupkit-custom-state-demo$(BINARY_EXT) -mode=silent

# Run custom state demo in CLI mode
.PHONY: run-custom-state-demo-cli
run-custom-state-demo-cli: build-custom-state-demo
	@echo "Starting custom state demo (CLI mode)..."
	@echo "Interactive CLI with database configuration state"
	@echo ""
	$(BIN_DIR)$(PATH_SEP)setupkit-custom-state-demo$(BINARY_EXT) -mode=cli

# Run custom state demo in auto mode
.PHONY: run-custom-state-demo-auto
run-custom-state-demo-auto: build-custom-state-demo
	@echo "Starting custom state demo (auto mode)..."
	@echo "Auto-selects best UI mode for database configuration"
	@echo ""
	$(BIN_DIR)$(PATH_SEP)setupkit-custom-state-demo$(BINARY_EXT) -mode=auto

# Show custom state demo help
.PHONY: help-custom-state-demo
help-custom-state-demo: build-custom-state-demo
	@echo "Custom State Demo Help:"
	@echo ""
	$(BIN_DIR)$(PATH_SEP)setupkit-custom-state-demo$(BINARY_EXT) --help

# Clean test installation directories
.PHONY: clean-install
clean-install:
	@echo "Cleaning test installation directories..."
	@rm -rf /tmp/DemoApp "C:\Program Files\DemoApp" ~/Applications/DemoApp /opt/demoapp 2>/dev/null || true
	@echo "✅ Test installations cleaned"

# Version info
.PHONY: version
version:
	@echo "SetupKit Framework"
	@echo "=================="
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Module: $(MODULE)"

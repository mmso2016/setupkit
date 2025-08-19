# Setup-Kit Makefile
# Cross-platform build system

# Variables
BINARY_NAME := installer
MODULE := github.com/mmso2016/setupkit
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GOFLAGS := -v
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

# Directories
BIN_DIR := bin
EXAMPLES_DIR := examples
INSTALLER_DIR := installer

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
	@echo "Setup-Kit Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Clean, test and build everything"
	@echo "  build        - Build all examples (non-Wails)"
	@echo "  build-all    - Build everything including Wails GUIs"
	@echo "  build-cli    - Build CLI installer"
	@echo "  build-gui    - Build Wails GUI installer (requires Wails)"
	@echo "  build-console - Build console GUI"
	@echo "  build-platform - Build platform example"
	@echo "  build-ui     - Build UI example (CLI mode)"
	@echo "  build-ui-wails - Build UI example with Wails GUI (requires Wails)"
	@echo "  build-ui-nogui - Build UI example without GUI"
	@echo "  test         - Run all tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  test-builds  - Test all build configurations"
	@echo "  test-configs - Test UI configuration system"
	@echo "  bench        - Run benchmarks"
	@echo "  coverage     - Generate test coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy go.mod"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run linters (requires golangci-lint)"
	@echo "  install      - Install to GOPATH/bin"
	@echo "  wails-check  - Check Wails installation"
	@echo "  help         - Show this help"

# Build targets
.PHONY: build
build: build-cli build-console build-platform build-ui build-ui-nogui
ifeq ($(OS),Windows_NT)
	@where wails >nul 2>&1 && $(MAKE) build-wails-targets || echo Skipping Wails targets (Wails not installed)
else
	@which wails > /dev/null 2>&1 && $(MAKE) build-wails-targets || echo "Skipping Wails targets (Wails not installed)"
endif

.PHONY: build-wails-targets
build-wails-targets: build-gui build-ui-wails

.PHONY: build-all
build-all: build build-wails-targets

.PHONY: build-cli
build-cli:
	@echo "Building CLI installer..."
	@$(MKDIR) $(BIN_DIR)
	go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)$(PATH_SEP)installer-cli$(BINARY_EXT) $(EXAMPLES_DIR)/basic

.PHONY: build-console
build-console:
	@echo "Building console GUI..."
	@$(MKDIR) $(BIN_DIR)
	go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)$(PATH_SEP)installer-console$(BINARY_EXT) $(EXAMPLES_DIR)/gui-console

.PHONY: build-platform
build-platform:
	@echo "Building platform example..."
	@$(MKDIR) $(BIN_DIR)
	go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)$(PATH_SEP)installer-platform$(BINARY_EXT) $(EXAMPLES_DIR)/platform

.PHONY: build-ui
build-ui:
	@echo "Building UI example (CLI mode)..."
	@$(MKDIR) $(BIN_DIR)
	go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)$(PATH_SEP)installer-ui-cli$(BINARY_EXT) $(EXAMPLES_DIR)/ui

.PHONY: build-ui-wails
build-ui-wails: wails-check
	@echo "Building UI example with Wails GUI..."
ifeq ($(OS),Windows_NT)
	@cd $(EXAMPLES_DIR)\ui && \
	(if not exist frontend\dist mkdir frontend\dist) && \
	(if not exist frontend\dist\app.js ( \
		echo Copying frontend files to dist... && \
		xcopy /E /I /Y frontend\src\*.* frontend\dist\ >nul 2>&1 \
	)) && \
	wails build -clean && \
	(if exist build\bin\installer-ui.exe ( \
		copy /Y build\bin\installer-ui.exe ..\..\$(BIN_DIR)\installer-ui$(BINARY_EXT) >nul \
	))
else
	@cd $(EXAMPLES_DIR)/ui && \
	if [ ! -d "frontend/dist" ] || [ ! -f "frontend/dist/app.js" ]; then \
		echo "Copying frontend files to dist..."; \
		cp -r frontend/src/* frontend/dist/ 2>/dev/null || true; \
	fi && \
	wails build -clean && \
	if [ -f "build/bin/installer-ui.exe" ]; then \
		cp build/bin/installer-ui.exe ../../$(BIN_DIR)/installer-ui$(BINARY_EXT); \
	elif [ -f "build/bin/installer-ui" ]; then \
		cp build/bin/installer-ui ../../$(BIN_DIR)/installer-ui; \
	fi
endif

.PHONY: build-ui-nogui
build-ui-nogui:
	@echo "Building UI example without GUI support..."
	@$(MKDIR) $(BIN_DIR)
	go build $(GOFLAGS) $(LDFLAGS) -tags nogui -o $(BIN_DIR)$(PATH_SEP)installer-ui-nogui$(BINARY_EXT) $(EXAMPLES_DIR)/ui

.PHONY: build-gui
build-gui: wails-check
	@echo "Building Wails GUI..."
ifeq ($(OS),Windows_NT)
	@cd $(EXAMPLES_DIR)\gui && \
	(if not exist frontend\dist mkdir frontend\dist) && \
	(if exist frontend\src\index.html ( \
		copy /Y frontend\src\index.html frontend\dist\ >nul 2>&1 \
	)) && \
	(if exist frontend\src\assets ( \
		if not exist frontend\dist\assets mkdir frontend\dist\assets && \
		xcopy /E /I /Y frontend\src\assets\*.* frontend\dist\assets\ >nul 2>&1 \
	)) && \
	wails build -clean && \
	(if exist build\bin\installer-gui.exe ( \
		copy /Y build\bin\installer-gui.exe ..\..\$(BIN_DIR)\installer-gui$(BINARY_EXT) >nul \
	))
else
	@cd $(EXAMPLES_DIR)/gui && \
	if [ ! -d "frontend/dist" ]; then \
		mkdir -p frontend/dist; \
	fi && \
	if [ -f "frontend/src/index.html" ]; then \
		cp frontend/src/index.html frontend/dist/; \
	fi && \
	if [ -d "frontend/src/assets" ]; then \
		mkdir -p frontend/dist/assets && \
		cp -r frontend/src/assets/* frontend/dist/assets/ 2>/dev/null || true; \
	fi && \
	wails build -clean && \
	if [ -f "build/bin/installer-gui.exe" ]; then \
		cp build/bin/installer-gui.exe ../../$(BIN_DIR)/installer-gui$(BINARY_EXT); \
	elif [ -f "build/bin/installer-gui" ]; then \
		cp build/bin/installer-gui ../../$(BIN_DIR)/installer-gui; \
	fi
endif

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	go test -short ./...

.PHONY: test-verbose
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./...

.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	go test -race -short ./...

.PHONY: test-builds
test-builds:
	@echo "Testing all build configurations..."
	@echo "This requires Mage to be installed"
	@which mage > /dev/null 2>&1 || (echo "Mage not found. Install: go install github.com/magefile/mage@latest" && exit 1)
	mage testAllBuilds

.PHONY: test-configs
test-configs:
	@echo "Testing UI configuration system..."
	@echo "1. Testing theme listing..."
	$(BIN_DIR)$(PATH_SEP)installer-cli$(BINARY_EXT) -list-themes || (make build-cli && $(BIN_DIR)$(PATH_SEP)installer-cli$(BINARY_EXT) -list-themes)
	@echo "2. Testing config generation..."
	$(BIN_DIR)$(PATH_SEP)installer-cli$(BINARY_EXT) -generate-config test-config.yaml || (make build-cli && $(BIN_DIR)$(PATH_SEP)installer-cli$(BINARY_EXT) -generate-config test-config.yaml)
	@echo "3. Testing with different themes..."
	for theme in default corporate-blue medical-green tech-dark minimal-light; do \
		echo "Testing theme: $theme"; \
		$(BIN_DIR)$(PATH_SEP)installer-cli$(BINARY_EXT) -theme $theme -help > /dev/null || echo "Theme $theme test failed"; \
	done
	@echo "4. Testing example configs..."
	for config in examples/configs/*.yaml; do \
		echo "Testing config: $config"; \
		$(BIN_DIR)$(PATH_SEP)installer-cli$(BINARY_EXT) -config $config -help > /dev/null || echo "Config $config test failed"; \
	done
	@$(RM) test-config.yaml
	@echo "UI configuration tests completed!"

.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	@$(MKDIR) coverage
	go test -coverprofile=coverage/coverage.out ./...
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
	@$(RM) -rf $(EXAMPLES_DIR)/gui/build
	@find . -name "*.test" -delete 2>/dev/null || true
	@find . -name "*.out" -delete 2>/dev/null || true

# Installation
.PHONY: install
install: build-cli
	@echo "Installing to GOPATH/bin..."
	go install $(MODULE)/examples/basic

# Wails check
.PHONY: wails-check
wails-check:
	@echo "Checking Wails installation..."
ifeq ($(OS),Windows_NT)
	@where wails >nul 2>&1 || (echo Wails not found. Install: go install github.com/wailsapp/wails/v2/cmd/wails@latest && exit 1)
else
	@which wails > /dev/null 2>&1 || (echo "Wails not found. Install: go install github.com/wailsapp/wails/v2/cmd/wails@latest" && exit 1)
endif
	@echo "Wails is installed"

# Development helpers
.PHONY: dev
dev:
	@echo "Starting development mode..."
	cd $(EXAMPLES_DIR)/gui && wails dev

.PHONY: run-cli
run-cli: build-cli
	@echo "Running CLI installer..."
	$(BIN_DIR)$(PATH_SEP)installer-cli$(BINARY_EXT) --help

.PHONY: run-console
run-console: build-console
	@echo "Running console GUI..."
	$(BIN_DIR)$(PATH_SEP)installer-console$(BINARY_EXT)

# Version info
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Module: $(MODULE)"

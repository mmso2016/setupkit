# SetupKit Makefile - Simplified Framework Version
# UI is provided by framework, not built by user

# Variables
BINARY_NAME := setupkit-example
MODULE := github.com/mmso2016/setupkit
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GOFLAGS := -v

# Directories
BIN_DIR := bin
EXAMPLE_DIR := ./examples/minimal

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
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Clean, test and build"
	@echo "  build        - Build the example installer"
	@echo "  test         - Run all tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  test-race    - Run tests with race detector"
	@echo "  bench        - Run benchmarks"
	@echo "  coverage     - Generate test coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy go.mod"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run linters (requires golangci-lint)"
	@echo "  run          - Run the example installer"
	@echo "  help         - Show this help"

# Build target
.PHONY: build
build:
	@echo "Building example installer..."
	@$(MKDIR) $(BIN_DIR)
	go build -tags desktop,production $(GOFLAGS) -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)" -o $(BIN_DIR)$(PATH_SEP)$(BINARY_NAME)$(BINARY_EXT) $(EXAMPLE_DIR)

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	go test -short ./pkg/... ./internal/...

.PHONY: test-verbose
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./pkg/... ./internal/...

.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	go test -race -short ./pkg/... ./internal/...

.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./pkg/... ./internal/...

.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	@$(MKDIR) coverage
	go test -coverprofile=coverage/coverage.out ./pkg/... ./internal/...
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

# Run example
.PHONY: run
run: build
	@echo "Running example installer..."
	$(BIN_DIR)$(PATH_SEP)$(BINARY_NAME)$(BINARY_EXT)

# Version info
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Module: $(MODULE)"
	@echo ""
	@echo "SetupKit - Modern Installer Framework"
	@echo "Create professional installers with minimal Go code!"

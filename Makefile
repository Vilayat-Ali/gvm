# GVM - Go Version Manager Makefile
# Copyright © 2025 Syed Vilayat Ali Rizvi

.PHONY: all build install uninstall clean test fmt lint vet vendor help

# Variables
APP_NAME := gvm
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "v1.0.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO ?= go
GOFMT ?= gofmt
GOLINT ?= golint
GOOS ?= $(shell go env GOOS 2>/dev/null || echo "linux")
GOARCH ?= $(shell go env GOARCH 2>/dev/null || echo "amd64")
CGO_ENABLED ?= 0

# Build flags - Update these if your version variables are in main package, not cmd
LDFLAGS := -s -w
ifneq ($(VERSION),)
	LDFLAGS += -X github.com/vilayat-ali/gvm/cmd.version=$(VERSION)
endif
ifneq ($(GIT_COMMIT),)
	LDFLAGS += -X github.com/vilayat-ali/gvm/cmd.commit=$(GIT_COMMIT)
endif
ifneq ($(BUILD_TIME),)
	LDFLAGS += -X github.com/vilayat-ali/gvm/cmd.buildTime=$(BUILD_TIME)
endif

# Directories
BIN_DIR := bin
DIST_DIR := dist

# Ensure directories exist
$(shell mkdir -p $(BIN_DIR) $(DIST_DIR) $(DIST_DIR)/releases)

# Default target
all: build

# Build the application for current platform - Build from root directory
build:
	@echo "Building $(APP_NAME) for $(GOOS)/$(GOARCH)..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(BIN_DIR)/$(APP_NAME) \
		.
	@chmod +x $(BIN_DIR)/$(APP_NAME)
	@echo "Build complete: $(BIN_DIR)/$(APP_NAME)"

# Quick build and run
run: build
	@echo "Running $(APP_NAME)..."
	./$(BIN_DIR)/$(APP_NAME)

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Build for Linux - Build from root directory
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-linux-amd64 \
		.
	@chmod +x $(DIST_DIR)/$(APP_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-linux-arm64 \
		.
	@chmod +x $(DIST_DIR)/$(APP_NAME)-linux-arm64
	@echo "Linux builds complete in $(DIST_DIR)/"

# Build for macOS - Build from root directory
build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 \
		.
	@chmod +x $(DIST_DIR)/$(APP_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 \
		.
	@chmod +x $(DIST_DIR)/$(APP_NAME)-darwin-arm64
	@echo "macOS builds complete in $(DIST_DIR)/"

# Build for Windows - Build from root directory
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe \
		.
	@echo "Windows build complete in $(DIST_DIR)/"

# Install to GOPATH/bin or system bin directory
install: build
	@echo "Installing $(APP_NAME)..."
	@if [ -d "$(shell go env GOPATH 2>/dev/null)/bin" ]; then \
		INSTALL_DIR="$(shell go env GOPATH)/bin"; \
		cp $(BIN_DIR)/$(APP_NAME) $$INSTALL_DIR/$(APP_NAME); \
		chmod +x $$INSTALL_DIR/$(APP_NAME); \
		echo "Installed to $$INSTALL_DIR/$(APP_NAME)"; \
	elif [ -d "/usr/local/bin" ]; then \
		sudo cp $(BIN_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME); \
		sudo chmod +x /usr/local/bin/$(APP_NAME); \
		echo "Installed to /usr/local/bin/$(APP_NAME)"; \
	elif [ -d "$HOME/.local/bin" ]; then \
		cp $(BIN_DIR)/$(APP_NAME) $$HOME/.local/bin/$(APP_NAME); \
		chmod +x $$HOME/.local/bin/$(APP_NAME); \
		echo "Installed to $$HOME/.local/bin/$(APP_NAME)"; \
	else \
		echo "Please add execute permissions manually:"; \
		echo "  chmod +x $(BIN_DIR)/$(APP_NAME)"; \
		echo "  sudo mv $(BIN_DIR)/$(APP_NAME) /usr/local/bin/"; \
	fi

# Uninstall from system
uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	@if [ -f "$(shell go env GOPATH 2>/dev/null)/bin/$(APP_NAME)" ]; then \
		rm -f "$(shell go env GOPATH)/bin/$(APP_NAME)"; \
		echo "Removed from $(shell go env GOPATH)/bin/$(APP_NAME)"; \
	elif [ -f "/usr/local/bin/$(APP_NAME)" ]; then \
		sudo rm -f /usr/local/bin/$(APP_NAME); \
		echo "Removed from /usr/local/bin/$(APP_NAME)"; \
	elif [ -f "$$HOME/.local/bin/$(APP_NAME)" ]; then \
		rm -f "$$HOME/.local/bin/$(APP_NAME)"; \
		echo "Removed from $$HOME/.local/bin/$(APP_NAME)"; \
	else \
		echo "$(APP_NAME) not found in common locations"; \
	fi

# Create release archives
release: build-all
	@echo "Creating release archives..."
	@mkdir -p $(DIST_DIR)/releases
	@cd $(DIST_DIR) && \
		tar -czf releases/$(APP_NAME)-linux-amd64-$(VERSION).tar.gz $(APP_NAME)-linux-amd64 && \
		tar -czf releases/$(APP_NAME)-linux-arm64-$(VERSION).tar.gz $(APP_NAME)-linux-arm64 && \
		zip -q releases/$(APP_NAME)-darwin-amd64-$(VERSION).zip $(APP_NAME)-darwin-amd64 && \
		zip -q releases/$(APP_NAME)-darwin-arm64-$(VERSION).zip $(APP_NAME)-darwin-arm64 && \
		zip -q releases/$(APP_NAME)-windows-amd64-$(VERSION).zip $(APP_NAME)-windows-amd64.exe
	@echo "Release archives created in $(DIST_DIR)/releases/"
	@echo "SHA256 checksums:"
	@cd $(DIST_DIR)/releases && sha256sum * > SHA256SUMS.txt
	@cat $(DIST_DIR)/releases/SHA256SUMS.txt

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format Go code
fmt:
	@echo "Formatting Go code..."
	$(GO) fmt ./...

# Lint Go code
lint:
	@echo "Linting Go code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		if golangci-lint --version 2>&1 | grep -q "version 1.59"; then \
			echo "Warning: golangci-lint version may have issues with Go 1.25"; \
			echo "Skipping golangci-lint, running go vet instead..."; \
			$(GO) vet ./...; \
		else \
			golangci-lint run || (echo "Linting had issues. Running go vet as fallback..."; $(GO) vet ./...); \
		fi; \
	else \
		echo "golangci-lint not found, running go vet..."; \
		$(GO) vet ./...; \
	fi

# Vet Go code
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run all code quality checks
check: fmt vet test

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out coverage.html
	$(GO) clean

# Generate Cobra CLI documentation
docs: build
	@echo "Generating documentation..."
	@mkdir -p docs
	./$(BIN_DIR)/$(APP_NAME) --help > docs/usage.txt
	@for cmd in download list use; do \
		./$(BIN_DIR)/$(APP_NAME) $$cmd --help 2>/dev/null > docs/$$cmd.txt || true; \
	done
	@echo "Documentation generated in docs/"

# Update dependencies
deps:
	@echo "Updating dependencies..."
	$(GO) mod tidy
	$(GO) mod download

# Development mode (watch and rebuild)
dev:
	@echo "Starting development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Installing air for live reload..."; \
		$(GO) install github.com/cosmtrek/air@latest; \
		$$(go env GOPATH)/bin/air; \
	fi

# Show help
help:
	@echo "GVM - Go Version Manager Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build for current platform"
	@echo "  run            - Build and run"
	@echo "  build-all      - Build for all platforms (Linux, macOS, Windows)"
	@echo "  build-linux    - Build for Linux (amd64, arm64)"
	@echo "  build-darwin   - Build for macOS (amd64, arm64)"
	@echo "  build-windows  - Build for Windows (amd64)"
	@echo "  install        - Install to system"
	@echo "  uninstall      - Uninstall from system"
	@echo "  release        - Create release archives with checksums"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Lint Go code"
	@echo "  vet            - Vet Go code"
	@echo "  check          - Run all code quality checks"
	@echo "  clean          - Clean build artifacts"
	@echo "  docs           - Generate CLI documentation"
	@echo "  deps           - Update dependencies"
	@echo "  dev            - Start development mode with live reload"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS      - Target OS (default: current)"
	@echo "  GOARCH    - Target architecture (default: current)"
	@echo "  CGO_ENABLED - Enable CGO (default: 0)"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build for current platform"
	@echo "  make run                      # Build and run"
	@echo "  make install                  # Install to system"
	@echo "  make release                  # Create release packages"

# Default target
.DEFAULT_GOAL := help
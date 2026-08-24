# GVM - Go Version Manager
# Copyright © 2026 Syed Vilayat Ali Rizvi

.PHONY: all build run build-all build-linux build-darwin install uninstall \
        release test test-race test-coverage fmt fmt-check lint vet check \
        clean clean-setup deps verify-go tools docs help

APP_NAME   := gvm
MODULE     := github.com/vilayat-ali/gvm
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# The Go toolchain version is declared once, in go.mod. Change it there and
# the Makefile, CI and `make verify-go` all follow automatically.
GO_VERSION := $(shell awk '/^go[[:space:]]/ {print $$2; exit}' go.mod)

GO          ?= go
GOOS        ?= $(shell $(GO) env GOOS)
GOARCH      ?= $(shell $(GO) env GOARCH)
CGO_ENABLED ?= 0

BIN_DIR      := bin
DIST_DIR     := dist
RELEASE_DIR  := $(DIST_DIR)/releases
PLATFORMS    := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

LDFLAGS := -s -w \
	-X '$(MODULE)/internal.AppVersion=$(VERSION)' \
	-X '$(MODULE)/internal.GitCommit=$(GIT_COMMIT)' \
	-X '$(MODULE)/internal.BuildTime=$(BUILD_TIME)'

GOBUILD := CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags "$(LDFLAGS)"

all: build

$(BIN_DIR) $(DIST_DIR) $(RELEASE_DIR):
	@mkdir -p $@

verify-go:
	@have=$$($(GO) env GOVERSION | sed 's/^go//'); \
	want=$(GO_VERSION); \
	printf 'go.mod requires go %s, toolchain provides go %s\n' "$$want" "$$have"; \
	lowest=$$(printf '%s\n%s\n' "$$want" "$$have" | sort -V | head -n 1); \
	if [ "$$lowest" != "$$want" ] && [ "$$want" != "$$have" ]; then \
		echo "error: your Go toolchain is older than go.mod requires"; exit 1; \
	fi

build: | $(BIN_DIR)
	@echo "Building $(APP_NAME) $(VERSION) for $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -o $(BIN_DIR)/$(APP_NAME) .
	@echo "Built $(BIN_DIR)/$(APP_NAME)"

run: build
	@./$(BIN_DIR)/$(APP_NAME)

build-all: | $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch $(GOBUILD) -o $(DIST_DIR)/$(APP_NAME)-$$os-$$arch . || exit 1; \
	done
	@echo "Binaries in $(DIST_DIR)/"

build-linux:
	@$(MAKE) --no-print-directory build GOOS=linux GOARCH=amd64
	@$(MAKE) --no-print-directory build GOOS=linux GOARCH=arm64

build-darwin:
	@$(MAKE) --no-print-directory build GOOS=darwin GOARCH=amd64
	@$(MAKE) --no-print-directory build GOOS=darwin GOARCH=arm64

# Installs into the user's own bin directory. gvm never needs root.
install: build
	@dir="$${PREFIX:-$$HOME/.local/bin}"; \
	mkdir -p "$$dir"; \
	install -m 0755 $(BIN_DIR)/$(APP_NAME) "$$dir/$(APP_NAME)"; \
	echo "Installed $$dir/$(APP_NAME)"; \
	case ":$$PATH:" in *":$$dir:"*) ;; *) echo "Note: $$dir is not on your PATH";; esac

uninstall:
	@dir="$${PREFIX:-$$HOME/.local/bin}"; \
	if [ -f "$$dir/$(APP_NAME)" ]; then \
		rm -f "$$dir/$(APP_NAME)"; echo "Removed $$dir/$(APP_NAME)"; \
	else \
		echo "$(APP_NAME) not found in $$dir"; \
	fi

# Archive names match exactly what scripts/install.sh downloads.
release: build-all | $(RELEASE_DIR)
	@rm -f $(RELEASE_DIR)/*
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		tar -czf $(RELEASE_DIR)/$(APP_NAME)-$$os-$$arch-$(VERSION).tar.gz \
			-C $(DIST_DIR) $(APP_NAME)-$$os-$$arch \
			--transform 's|$(APP_NAME)-.*|$(APP_NAME)|' || exit 1; \
	done
	@cd $(RELEASE_DIR) && sha256sum *.tar.gz > SHA256SUMS.txt
	@echo "Release artifacts in $(RELEASE_DIR)/"
	@cat $(RELEASE_DIR)/SHA256SUMS.txt

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -n 1
	$(GO) tool cover -html=coverage.out -o coverage.html

fmt:
	$(GO) fmt ./...

fmt-check:
	@files=$$(gofmt -l . 2>/dev/null); \
	if [ -n "$$files" ]; then echo "Not gofmt'd:"; echo "$$files"; exit 1; fi; \
	echo "All files are formatted"

vet:
	$(GO) vet ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; running go vet instead"; \
		$(GO) vet ./...; \
	fi

check: verify-go fmt-check vet test

deps:
	$(GO) mod tidy
	$(GO) mod verify

docs: build
	@mkdir -p docs
	@./$(BIN_DIR)/$(APP_NAME) --help > docs/usage.txt
	@for sub in configure list download use remove doctor env; do \
		./$(BIN_DIR)/$(APP_NAME) $$sub --help > docs/$$sub.txt 2>/dev/null || true; \
	done
	@echo "Documentation in docs/"

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR) docs coverage.out coverage.html
	$(GO) clean

# Removes gvm's own data. Only ever touches gvm-owned directories.
clean-setup:
	@root="$${GVM_ROOT:-$${XDG_DATA_HOME:-$$HOME/.local/share}/gvm}"; \
	config="$${XDG_CONFIG_HOME:-$$HOME/.config}/gvm"; \
	echo "This will delete:"; echo "  $$root"; echo "  $$config"; \
	printf 'Continue? [y/N] '; read -r reply; \
	case "$$reply" in [yY]*) rm -rf "$$root" "$$config"; echo "Removed";; *) echo "Aborted";; esac

help:
	@echo "gvm $(VERSION)  (go $(GO_VERSION), $(GOOS)/$(GOARCH))"
	@echo ""
	@echo "  build           Build for the current platform"
	@echo "  build-all       Build for $(PLATFORMS)"
	@echo "  install         Install to \$$PREFIX (default ~/.local/bin)"
	@echo "  uninstall       Remove the installed binary"
	@echo "  release         Build all platforms and write SHA256SUMS.txt"
	@echo "  test            Run tests"
	@echo "  test-race       Run tests with the race detector"
	@echo "  test-coverage   Run tests and write coverage.html"
	@echo "  fmt / fmt-check Format code / fail if unformatted"
	@echo "  vet / lint      Static analysis"
	@echo "  check           verify-go + fmt-check + vet + test"
	@echo "  verify-go       Check the toolchain matches go.mod"
	@echo "  deps            go mod tidy && go mod verify"
	@echo "  clean           Remove build artifacts"
	@echo "  clean-setup     Remove gvm's installed toolchains and config"
	@echo ""
	@echo "The required Go version lives in go.mod (currently $(GO_VERSION))."

.DEFAULT_GOAL := help

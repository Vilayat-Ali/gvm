.PHONY: build clean test run install lint fmt help

BINARY_NAME=./build/gvm
GO=go
GOFLAGS=-v

help:
	@echo "Available targets:"
	@echo "  build   - Build the project"
	@echo "  clean   - Remove build artifacts"
	@echo "  test    - Run tests"
	@echo "  run     - Build and run the project"
	@echo "  install - Install the binary"
	@echo "  lint    - Run linter"
	@echo "  fmt     - Format code"

build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/gvm-cli/main.go

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

test:
	$(GO) test $(GOFLAGS) ./...

run: build
	./$(BINARY_NAME)

install:
	$(GO) install $(GOFLAGS)

lint:
	golangci-lint run ./...

fmt:
	$(GO) fmt ./...
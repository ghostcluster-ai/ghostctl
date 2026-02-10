.PHONY: build install clean test lint help

# Variables
BINARY_NAME=ghostctl
BINARY_PATH=./bin/$(BINARY_NAME)
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Build variables
VERSION?=$(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Linker flags for binary
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

help: ## Display this help screen
	@echo "Ghostctl Makefile"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the ghostctl binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY_PATH) .
	@echo "Binary built: $(BINARY_PATH)"

build-linux: ## Build for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY_PATH)-linux-amd64 .
	@echo "Binary built: $(BINARY_PATH)-linux-amd64"

build-darwin: ## Build for macOS
	@echo "Building $(BINARY_NAME) for macOS..."
	@mkdir -p bin
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY_PATH)-darwin-amd64 .
	@echo "Binary built: $(BINARY_PATH)-darwin-amd64"

build-windows: ## Build for Windows
	@echo "Building $(BINARY_NAME) for Windows..."
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY_PATH).exe .
	@echo "Binary built: $(BINARY_PATH).exe"

install: build ## Install ghostctl to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_PATH) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

install-dev: build ## Install ghostctl for local development
	@echo "Installing $(BINARY_NAME) for development..."
	@mkdir -p $(HOME)/.local/bin
	@cp $(BINARY_PATH) $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Installed to $(HOME)/.local/bin/$(BINARY_NAME)"
	@echo "Add to PATH: export PATH=$(HOME)/.local/bin:\$$PATH"

uninstall: ## Uninstall ghostctl
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@rm -f $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Uninstalled"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@$(GO) clean
	@echo "Cleaned"

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests completed"

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...
	@echo "Linting completed"

fmt: ## Format code
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "Formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GO) vet ./...
	@echo "Vet completed"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "Dependencies updated"

version: ## Display version information
	@echo "ghostctl version $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build time: $(BUILD_TIME)"

all: clean fmt vet lint test build ## Run all checks and build
	@echo "All tasks completed"

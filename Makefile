# Makefile for unified-ads-mcp

# Variables
BINARY_NAME=unified-ads-mcp
FACEBOOK_BINARY=facebook-mcp
IMPROVED_BINARY=unified-ads-mcp-improved
CODEGEN_DIR=internal/facebook/codegen
API_SPECS_DIR=internal/facebook/api_specs/specs
GENERATED_DIR=internal/facebook/generated

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test run codegen fmt vet deps help pre-commit pre-commit-install

# Default target
all: build

# Build the main server
build:
	@echo "Building unified-ads-mcp server..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/server

# Build the Facebook-specific server
build-facebook:
	@echo "Building Facebook MCP server..."
	$(GOBUILD) $(LDFLAGS) -o $(FACEBOOK_BINARY) ./cmd/facebook-mcp

# Build the improved server with context support
build-improved:
	@echo "Building improved MCP server with context support..."
	$(GOBUILD) $(LDFLAGS) -o $(IMPROVED_BINARY) ./cmd/server-improved

# Build all binaries
build-all: build build-facebook build-improved

# Run the main server
run: build
	@echo "Starting unified-ads-mcp server..."
	@if [ -z "$(FACEBOOK_ACCESS_TOKEN)" ]; then \
		echo "Warning: FACEBOOK_ACCESS_TOKEN not set"; \
	fi
	./$(BINARY_NAME)

# Run the Facebook-specific server
run-facebook: build-facebook
	@echo "Starting Facebook MCP server..."
	@if [ -z "$(FACEBOOK_ACCESS_TOKEN)" ]; then \
		echo "Error: FACEBOOK_ACCESS_TOKEN environment variable must be set"; \
		exit 1; \
	fi
	./$(FACEBOOK_BINARY)

# Run the improved server with context support
run-improved: build-improved
	@echo "Starting improved MCP server..."
	@if [ -z "$(FACEBOOK_ACCESS_TOKEN)" ]; then \
		echo "Error: FACEBOOK_ACCESS_TOKEN environment variable must be set"; \
		exit 1; \
	fi
	@echo "Enabled categories: $(ENABLED_CATEGORIES)"
	./$(IMPROVED_BINARY)

# Run code generation
codegen:
	@echo "Running code generation for Facebook API..."
	@cd $(CODEGEN_DIR) && $(GOCMD) run main.go ../api_specs/specs
	@echo "Code generation completed!"
	@echo "Running formatters on generated code..."
	@find $(GENERATED_DIR) -name "*.go" -exec gofmt -w {} \;
	@if command -v goimports > /dev/null; then \
		find $(GENERATED_DIR) -name "*.go" -exec goimports -w {} \; ; \
	fi
	@echo "Formatting completed!"


# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run linting (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME) $(FACEBOOK_BINARY)
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Development mode - rebuild and run on file changes (requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		make run; \
	fi

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed!"


# Check code quality (format, vet, lint, test)
check: fmt vet lint test
	@echo "Code quality checks passed!"


# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the main server"
	@echo "  make build-facebook - Build the Facebook-specific server"
	@echo "  make build-improved - Build the improved server with context support"
	@echo "  make build-all      - Build all binaries"
	@echo "  make run            - Build and run the main server"
	@echo "  make run-facebook   - Build and run the Facebook server"
	@echo "  make run-improved   - Build and run the improved server (ENABLED_CATEGORIES=core_ads,reporting)"
	@echo "  make codegen        - Run code generation with formatting"
	@echo "  make regenerate     - Clean and regenerate code"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run linter (requires golangci-lint)"
	@echo "  make deps           - Download dependencies"
	@echo "  make deps-update    - Update dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make clean-generated- Clean generated files"
	@echo "  make dev            - Run in development mode with hot reload"
	@echo "  make install-tools  - Install development tools"
	@echo "  make check          - Run all code quality checks"
	@echo "  make help           - Show this help message"

# Set default goal
.DEFAULT_GOAL := help
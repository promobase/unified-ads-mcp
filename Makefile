# Makefile for unified-ads-mcp

# Variables
BINARY_NAME=unified-ads-mcp
FACEBOOK_BINARY=facebook-mcp
GOCMD=go
GENERATED_DIR=internal/facebook/generated

.PHONY: build run codegen test fmt clean

# Default target
all: build

# Build the unified server
build: codegen
	@echo "Building unified-ads-mcp server..."
	$(GOCMD) build -ldflags "-s -w" -o $(BINARY_NAME) ./cmd/server
	$(GOCMD) fmt ./...

# Build the Facebook-only server
build-facebook:
	@echo "Building Facebook MCP server..."
	$(GOCMD) build -ldflags "-s -w" -o $(FACEBOOK_BINARY) ./cmd/facebook-mcp

# Run the server (default: core_ads)
run: build
	./$(BINARY_NAME)

# Run the server with all categories
run-all: build
	./$(BINARY_NAME) --categories=all

# Run the server with specific categories
run-reporting: build
	./$(BINARY_NAME) --categories=core_ads,reporting

# Show help
help: build
	./$(BINARY_NAME) --help

# Run the Facebook server
run-facebook: build-facebook
	./$(FACEBOOK_BINARY)

# Run code generation
codegen:
	@echo "Running code generation for Facebook API..."
	@cd internal/facebook/codegen && $(GOCMD) run main.go ../api_specs/specs

# Run tests
test:
	$(GOCMD) test -v ./...

# Format code
fmt:
	$(GOCMD) fmt ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(FACEBOOK_BINARY)

# Install dependencies
deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy
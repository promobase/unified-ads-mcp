# Makefile for unified-ads-mcp

# Variables
BINARY_NAME=unified-ads-mcp
GOCMD=go
GENERATED_DIR=internal/facebook/generated

.PHONY: build run codegen test fmt clean

# Default target
all: build

# Build the server
build:
	@echo "Building unified-ads-mcp server..."
	$(GOCMD) build -ldflags "-s -w" -o $(BINARY_NAME) ./cmd/server

# Run the server
run: build
	./$(BINARY_NAME)

# Run code generation
codegen:
	@echo "Running code generation for Facebook API..."
	@cd internal/facebook/codegen && $(GOCMD) run main.go ../api_specs/specs
	@echo "Running formatters..."
	@find $(GENERATED_DIR) -name "*.go" -exec gofmt -w {} \;
	@if command -v goimports > /dev/null; then \
		find $(GENERATED_DIR) -name "*.go" -exec goimports -w {} \; ; \
	fi

# Run tests
test:
	$(GOCMD) test -v ./...

# Format code
fmt:
	$(GOCMD) fmt ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

# Install dependencies
deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy
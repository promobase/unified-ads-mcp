BINARY_NAME=unified-ads-mcp
FACEBOOK_BINARY=facebook-mcp
GOCMD=go
GENERATED_DIR=internal/facebook/generated

.PHONY: build run codegen test fmt clean deps check

all: build

build:
	@echo "Checking all packages for compilation errors..."
	@$(GOCMD) build ./... || (echo "ERROR: Build failed for packages" && exit 1)
	@echo "Package compilation successful"
	@echo "Checking test compilation..."
	@$(GOCMD) test -run=xxxxx ./... > /dev/null 2>&1 || (echo "ERROR: Test compilation failed" && exit 1)
	@echo "Test compilation successful"
	@echo "Building unified-ads-mcp server..."
	@$(GOCMD) build -ldflags "-s -w" -o $(BINARY_NAME) ./cmd/server || (echo "ERROR: Failed to build server" && exit 1)
	@echo "Built $(BINARY_NAME)"
	@echo "Building facebook-mcp..."
	@$(GOCMD) build -ldflags "-s -w" -o $(FACEBOOK_BINARY) ./cmd/facebook-mcp || (echo "ERROR: Failed to build facebook-mcp" && exit 1)
	@echo "Built $(FACEBOOK_BINARY)"
	@$(GOCMD) fmt ./...

run: build
	./$(BINARY_NAME) -t http

help: build
	./$(BINARY_NAME) --help

run-facebook: build-facebook
	./$(FACEBOOK_BINARY)

codegen:
	@echo "Running code generation for Facebook API..."
	@cd internal/facebook/codegen && $(GOCMD) run . -specs ../api_specs/specs

test:
	$(GOCMD) test -v ./...

fmt:
	$(GOCMD) fmt ./...

clean:
	rm -f $(BINARY_NAME) $(FACEBOOK_BINARY)
	rm -rf dist/

deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy


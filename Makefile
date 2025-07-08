BINARY_NAME=unified-ads-mcp
FACEBOOK_BINARY=facebook-mcp
GOCMD=go
GENERATED_DIR=internal/facebook/generated

.PHONY: build run codegen test fmt clean

all: build

build:
	@echo "Building unified-ads-mcp server..."
	$(GOCMD) build -ldflags "-s -w" -o $(BINARY_NAME) ./cmd/server
	$(GOCMD) fmt ./...

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

deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy
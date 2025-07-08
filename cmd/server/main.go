package main

import (
	"context"
	"flag"
	"log"
	"os"

	"unified-ads-mcp/internal/facebook/generated"
	"unified-ads-mcp/internal/tools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewFacebookMCPServer() *server.MCPServer {
	// Create hooks for debugging (optional)
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		log.Printf("[DEBUG] Before %s: %v", method, id)
	})

	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		log.Printf("[ERROR] %s failed: %v", method, err)
	})

	// Create the MCP server
	mcpServer := server.NewMCPServer(
		"facebook-business-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithHooks(hooks),
	)

	// Register all Facebook Business API tools
	if err := generated.RegisterAllTools(mcpServer); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Register custom health and diagnostic tools
	if err := tools.RegisterHealthTools(mcpServer); err != nil {
		log.Fatalf("Failed to register health tools: %v", err)
	}

	return mcpServer
}

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or http)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or http)")
	flag.Parse()

	// Check for Facebook access token
	if os.Getenv("FACEBOOK_ACCESS_TOKEN") == "" {
		log.Fatalln("FACEBOOK_ACCESS_TOKEN environment variable must be set")
	}

	// Create and start the server
	mcpServer := NewFacebookMCPServer()

	if transport == "http" {
		httpServer := server.NewStreamableHTTPServer(mcpServer)
		log.Printf("HTTP server listening on :8080/mcp")
		if err := httpServer.Start(":8080"); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		log.Printf("Starting stdio server...")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

package main

import (
	"context"
	"flag"
	"log"
	"os"

	"unified-ads-mcp/internal/facebook/generated"
	"unified-ads-mcp/internal/facebook/tools"
	"unified-ads-mcp/internal/utils"

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

	// register low level tools
	if err := generated.RegisterAllTools(mcpServer); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// ---- High level tools ----
	if err := tools.RegisterAccountTools(mcpServer); err != nil {
		log.Fatalf("Failed to register account tools: %v", err)
	}
	tools.RegisterBatchTools(mcpServer)

	return mcpServer
}

func main() {
	var transport string

	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or http)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or http)")
	flag.Parse()

	utils.LoadFacebookConfig()

	// Check for Facebook access token
	if os.Getenv("FACEBOOK_ACCESS_TOKEN") == "" {
		log.Fatalln("FACEBOOK_ACCESS_TOKEN environment variable must be set")
	}

	// Create and start the server
	mcpServer := NewFacebookMCPServer()

	if transport == "http" {
		httpServer := server.NewStreamableHTTPServer(mcpServer)
		log.Printf("Starting Facebook Business MCP Server (HTTP mode)")
		log.Printf("Listening on :8080/mcp")
		log.Printf("Registered %d tools for Ad management", 162)
		if err := httpServer.Start(":8080"); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		// In stdio mode, be quiet unless there's an error
		// MCP clients expect clean JSON communication
		log.Printf("Starting Facebook Business MCP Server (stdio mode)")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

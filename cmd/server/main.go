package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"unified-ads-mcp/internal/facebook/generated"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewFacebookMCPServer() *server.MCPServer {
	if os.Getenv("FACEBOOK_ACCESS_TOKEN") == "" {
		log.Fatal("FACEBOOK_ACCESS_TOKEN environment variable must be set")
	}
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

	// Add a health check tool
	mcpServer.AddTool(
		mcp.NewTool("health_check",
			mcp.WithDescription("Check if the Facebook Business MCP server is running"),
		),
		handleHealthCheck,
	)

	// Add a tool to check Facebook access token
	mcpServer.AddTool(
		mcp.NewTool("check_access_token",
			mcp.WithDescription("Check if Facebook access token is configured"),
		),
		handleCheckAccessToken,
	)

	return mcpServer
}

func handleHealthCheck(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("Facebook Business MCP server is running"), nil
}

func handleCheckAccessToken(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	token := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	if token == "" {
		return mcp.NewToolResultText("WARNING: FACEBOOK_ACCESS_TOKEN environment variable is not set"), nil
	}

	// Mask the token for security
	maskedToken := token[:10] + "..." + token[len(token)-4:]
	return mcp.NewToolResultText(fmt.Sprintf("Facebook access token is configured: %s", maskedToken)), nil
}

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or http)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or http)")
	flag.Parse()

	// Check for Facebook access token
	if os.Getenv("FACEBOOK_ACCESS_TOKEN") == "" {
		log.Println("WARNING: FACEBOOK_ACCESS_TOKEN environment variable is not set")
		log.Println("The server will start, but API calls will fail without a valid token")
		log.Println("Get your token from: https://developers.facebook.com/tools/explorer/")
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

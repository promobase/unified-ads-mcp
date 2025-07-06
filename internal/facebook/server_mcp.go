package facebook

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"unified-ads-mcp/internal/facebook/generated/tools"
)

// CreateMCPServer creates a new MCP server with all Facebook Business API tools registered
func CreateMCPServer() (*server.MCPServer, error) {
	// Get access token from environment variable
	accessToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("FACEBOOK_ACCESS_TOKEN environment variable not set")
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"Facebook Business API MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Register all Facebook Business API tools
	if err := tools.RegisterTools(s, accessToken); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return s, nil
}

// RunServer starts the MCP server
func RunServer() error {
	s, err := CreateMCPServer()
	if err != nil {
		return err
	}

	// Start the server using stdio transport
	return server.ServeStdio(s)
}

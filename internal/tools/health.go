package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterHealthTools registers health and diagnostic tools
func RegisterHealthTools(mcpServer *server.MCPServer) error {
	// Add a tool to check Facebook access token
	mcpServer.AddTool(
		mcp.NewTool("check_access_token",
			mcp.WithDescription("Check if Facebook access token is configured"),
		),
		handleCheckAccessToken,
	)

	return nil
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

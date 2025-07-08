package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterHealthTools(mcpServer *server.MCPServer) error {
	mcpServer.AddTool(
		mcp.NewTool("get_default_ad_account",
			mcp.WithDescription("Read the default ad account."),
		),
		handleGetDefaultAdAccount,
	)

	return nil
}

func handleGetDefaultAdAccount(
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

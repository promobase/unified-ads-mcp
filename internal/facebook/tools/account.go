package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAccountTools registers account-related tools
func RegisterAccountTools(mcpServer *server.MCPServer) error {
	// Add tool to get current user's ad accounts
	mcpServer.AddTool(
		mcp.NewTool("get_my_adaccounts",
			mcp.WithDescription("Get all ad accounts accessible to the current authenticated user"),
		),
		handleGetMyAdAccounts,
	)

	return nil
}

func handleGetMyAdAccounts(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	accessToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	if accessToken == "" {
		return mcp.NewToolResultErrorf("FACEBOOK_ACCESS_TOKEN environment variable is not set"), nil
	}

	// Build the API URL
	apiURL := "https://graph.facebook.com/v23.0/me/adaccounts"
	params := url.Values{}
	params.Add("access_token", accessToken)
	params.Add("fields", "id,account_id,name,account_status,currency,timezone_id,timezone_name,timezone_offset_hours_utc")

	fullURL := apiURL + "?" + params.Encode()

	// Make the HTTP request
	resp, err := http.Get(fullURL)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to make API request: %v", err), nil
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to read response body: %v", err), nil
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return mcp.NewToolResultErrorf("API request failed with status %d: %s", resp.StatusCode, string(body)), nil
	}

	// Return the JSON directly
	return mcp.NewToolResultText(string(body)), nil
}

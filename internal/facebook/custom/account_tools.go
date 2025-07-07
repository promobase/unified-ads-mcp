package custom

import (
	"context"
	"encoding/json"
	"fmt"

	"unified-ads-mcp/internal/facebook/utils"
	"unified-ads-mcp/internal/shared"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	// FacebookGraphAPIBaseURL is the base URL for Facebook Graph API
	FacebookGraphAPIBaseURL = "https://graph.facebook.com/v23.0"
)

// GetDefaultAdAccountTool returns the MCP tool definition for get_default_ad_account
func GetDefaultAdAccountTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_default_ad_account",
		Description: "Get the default advertising account for the authenticated user. Returns the first accessible ad account with its details including account ID, name, currency, and status.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"fields": map[string]interface{}{
					"type":        "array",
					"description": "Fields to retrieve for the ad account (e.g., ['id', 'name', 'currency', 'account_status'])",
					"items": map[string]interface{}{
						"type": "string",
					},
					"default": []string{"id", "name", "currency", "account_status", "balance", "spend_cap"},
				},
			},
		},
	}
}

// HandleGetDefaultAdAccount handles the get_default_ad_account tool request
func HandleGetDefaultAdAccount(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get access token from context
	accessToken, ok := shared.FacebookAccessTokenFromContext(ctx)
	if !ok || accessToken == "" {
		return mcp.NewToolResultError("Facebook access token not found in context"), nil
	}

	// Build arguments map
	args := make(map[string]interface{})

	// Set default fields
	args["fields"] = "id,name,currency,account_status,balance,spend_cap"
	args["limit"] = 1 // Get only the first account as default

	// Parse optional fields parameter
	utils.ParseFieldsArray(request, args)

	// Build URL
	baseURL := FacebookGraphAPIBaseURL + "/me/adaccounts"

	// Build URL parameters
	urlParams := utils.BuildURLParams(accessToken, args)

	// Execute API request
	response, err := utils.ExecuteAPIRequest("GET", baseURL, urlParams)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get default ad account: %v", err)), nil
	}

	// Parse response based on type
	var result struct {
		Data []map[string]interface{} `json:"data"`
	}

	// Handle different response types
	switch resp := response.(type) {
	case map[string]interface{}:
		// Response is already a map, marshal and unmarshal to get into our struct
		jsonData, _ := json.Marshal(resp)
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to parse response: %v", err)), nil
		}
	case string:
		// Response is a string (non-JSON response)
		return mcp.NewToolResultText(resp), nil
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unexpected response type: %T", resp)), nil
	}

	// Check if user has any ad accounts
	if len(result.Data) == 0 {
		return mcp.NewToolResultText("No ad accounts found for the authenticated user."), nil
	}

	// Return the first ad account as default
	defaultAccount := result.Data[0]

	// Format the response as JSON
	jsonBytes, err := json.MarshalIndent(defaultAccount, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

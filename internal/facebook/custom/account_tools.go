package custom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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
		return nil, fmt.Errorf("no Facebook access token found in context")
	}

	// Parse parameters
	fields := []string{"id", "name", "currency", "account_status", "balance", "spend_cap"}
	
	// Get fields parameter if provided (passed as JSON array string)
	if fieldsStr := request.GetString("fields", ""); fieldsStr != "" {
		var fieldsArray []string
		if err := json.Unmarshal([]byte(fieldsStr), &fieldsArray); err == nil {
			fields = fieldsArray
		}
	}

	// Build URL to get user's ad accounts
	baseURL := fmt.Sprintf("%s/me/adaccounts", FacebookGraphAPIBaseURL)
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Set query parameters
	q := u.Query()
	q.Set("access_token", accessToken)
	q.Set("fields", joinFields(fields))
	q.Set("limit", "1") // Get only the first account as default
	u.RawQuery = q.Encode()

	// Make HTTP request
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		Data []map[string]interface{} `json:"data"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for API errors
	if result.Error != nil {
		return nil, fmt.Errorf("Facebook API error: %s (code: %d, type: %s)", 
			result.Error.Message, result.Error.Code, result.Error.Type)
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

// joinFields joins field names into a comma-separated string
func joinFields(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	result := fields[0]
	for i := 1; i < len(fields); i++ {
		result += "," + fields[i]
	}
	return result
}
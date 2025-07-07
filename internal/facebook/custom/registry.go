package custom

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetCustomTools returns all custom (non-generated) MCP tools
func GetCustomTools() []mcp.Tool {
	return []mcp.Tool{
		GetDefaultAdAccountTool(),
		// Add more custom tools here as needed
	}
}

// GetCustomHandlers returns handlers for all custom tools
func GetCustomHandlers() map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error){
		"get_default_ad_account": HandleGetDefaultAdAccount,
		// Add more custom handlers here as needed
	}
}
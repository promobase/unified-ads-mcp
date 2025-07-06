// Package facebook provides improved API functions for context-aware MCP tools
package facebook

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"unified-ads-mcp/internal/facebook/generated/tools"
)

// GetFilteredMCPTools returns filtered MCP tools based on enabled object types
func GetFilteredMCPTools(enabledObjects map[string]bool) []mcp.Tool {
	return tools.GetFilteredToolsWithoutAuth(enabledObjects)
}

// GetContextAwareHandlers returns context-aware handlers that don't require access token in params
func GetContextAwareHandlers(accessToken string) map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.GetContextAwareHandlers()
}

// RegisterFilteredMCPTools registers filtered tools with context-aware handlers
func RegisterFilteredMCPTools(s *server.MCPServer, enabledObjects map[string]bool) error {
	return tools.RegisterFilteredToolsWithContextAuth(s, enabledObjects)
}

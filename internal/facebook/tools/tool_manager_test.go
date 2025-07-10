package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestToolManager(t *testing.T) {
	// Create a test server
	s := server.NewMCPServer(
		"test-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register the tool manager
	err := RegisterToolManagerTool(s)
	if err != nil {
		t.Fatalf("Failed to register tool manager: %v", err)
	}

	ctx := context.Background()

	t.Run("List available scopes", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tool_manager",
				Arguments: map[string]interface{}{
					"action": "list",
				},
			},
		}

		result, err := handleToolManager(ctx, req)
		if err != nil {
			t.Fatalf("Failed to list scopes: %v", err)
		}

		if len(result.Content) == 0 {
			t.Fatal("Expected content in result")
		}

		// Check that result contains scopes information
		content := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(content, "custom_scopes") {
			t.Error("Expected custom_scopes in result")
		}
		if !strings.Contains(content, "codegen_scopes") {
			t.Error("Expected codegen_scopes in result")
		}
		if !strings.Contains(content, "campaign") {
			t.Error("Expected campaign in codegen scopes")
		}
		if !strings.Contains(content, "essentials") {
			t.Error("Expected essentials in custom scopes")
		}
	})

	t.Run("Get loaded scopes (initially empty)", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tool_manager",
				Arguments: map[string]interface{}{
					"action": "get",
				},
			},
		}

		result, err := handleToolManager(ctx, req)
		if err != nil {
			t.Fatalf("Failed to get scopes: %v", err)
		}

		content := result.Content[0].(mcp.TextContent).Text
		var data map[string]interface{}
		json.Unmarshal([]byte(content), &data)

		if data["total_loaded"].(float64) != 0 {
			t.Error("Expected no loaded scopes initially")
		}
	})

	t.Run("Add scopes", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tool_manager",
				Arguments: map[string]interface{}{
					"action": "add",
					"scopes": []string{"campaign", "adset"},
				},
			},
		}

		result, err := handleToolManager(ctx, req)
		if err != nil {
			t.Fatalf("Failed to add scopes: %v", err)
		}

		content := result.Content[0].(mcp.TextContent).Text
		var data map[string]interface{}
		json.Unmarshal([]byte(content), &data)

		added := data["added"].([]interface{})
		if len(added) != 2 {
			t.Errorf("Expected 2 scopes added, got %d", len(added))
		}
	})

	t.Run("Remove scopes", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tool_manager",
				Arguments: map[string]interface{}{
					"action": "remove",
					"scopes": []string{"campaign"},
				},
			},
		}

		result, err := handleToolManager(ctx, req)
		if err != nil {
			t.Fatalf("Failed to remove scopes: %v", err)
		}

		content := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(content, "removed") {
			t.Error("Expected removed in result")
		}
	})

	t.Run("Invalid action", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tool_manager",
				Arguments: map[string]interface{}{
					"action": "invalid",
				},
			},
		}

		result, err := handleToolManager(ctx, req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for invalid action")
		}
	})

	t.Run("Add custom scope", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tool_manager",
				Arguments: map[string]interface{}{
					"action": "add",
					"scopes": []string{"essentials"},
				},
			},
		}

		result, err := handleToolManager(ctx, req)
		if err != nil {
			t.Fatalf("Failed to add custom scope: %v", err)
		}

		content := result.Content[0].(mcp.TextContent).Text
		var data map[string]interface{}
		json.Unmarshal([]byte(content), &data)

		added := data["added"].([]interface{})
		if len(added) != 1 || added[0] != "essentials" {
			t.Errorf("Expected essentials scope to be added, got %v", added)
		}
	})

	t.Run("List shows custom scopes", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "tool_manager",
				Arguments: map[string]interface{}{
					"action": "list",
				},
			},
		}

		result, err := handleToolManager(ctx, req)
		if err != nil {
			t.Fatalf("Failed to list scopes: %v", err)
		}

		content := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(content, "custom_scopes") {
			t.Error("Expected custom_scopes in result")
		}
		if !strings.Contains(content, "essentials") {
			t.Error("Expected essentials in custom scopes")
		}
		if !strings.Contains(content, "campaign_management") {
			t.Error("Expected campaign_management in custom scopes")
		}
	})
}

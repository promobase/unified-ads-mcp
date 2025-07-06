package facebook

import (
	"testing"
	
	"unified-ads-mcp/internal/facebook/generated/tools"
)

func TestGetAllTools(t *testing.T) {
	// Test that we can get all tools
	accessToken := "test-token"
	allTools := tools.GetAllTools(accessToken)
	
	if len(allTools) == 0 {
		t.Error("Expected to get some tools, but got none")
	}
	
	// Check that tools have names
	for _, tool := range allTools {
		if tool.Name == "" {
			t.Error("Found tool with empty name")
		}
	}
	
	t.Logf("Found %d tools", len(allTools))
}
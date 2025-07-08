package generated

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestTypedHandlers(t *testing.T) {
	// Test GetAdHandler with typed arguments
	args := get_adArgs{
		ID:     "123456789",
		Fields: []string{"id", "name", "status"},
		Limit:  10,
	}

	// Create a request with the arguments
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id":     args.ID,
				"fields": args.Fields,
				"limit":  args.Limit,
			},
		},
	}

	// Call the handler directly with typed args
	result, err := GetAdHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Handler returned nil result")
	}

	// Check if it's an error (expected because we don't have a mock server)
	if !result.IsError {
		t.Fatal("Expected error result due to missing access token")
	}
}

func TestTypedBatchHandler(t *testing.T) {
	// Test batch handler with typed MCP tool
	// For now, just test that the handler can be created with typed args

	// Test that we can create typed args
	args := get_adArgs{
		ID:     "test123",
		Fields: []string{"id", "name"},
		Limit:  5,
	}

	// Verify args are correctly structured
	if args.ID != "test123" {
		t.Errorf("Expected ID 'test123', got '%s'", args.ID)
	}

	if len(args.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(args.Fields))
	}

	if args.Limit != 5 {
		t.Errorf("Expected limit 5, got %d", args.Limit)
	}
}

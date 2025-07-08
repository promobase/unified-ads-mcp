package generated

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestAdAccountGETActivitiesHandler(t *testing.T) {
	// Skip if no access token
	if os.Getenv("FACEBOOK_ACCESS_TOKEN") == "" {
		t.Skip("FACEBOOK_ACCESS_TOKEN not set")
	}

	// Create a mock request
	params := map[string]interface{}{
		"id":     "act_123456789", // Replace with a real ad account ID
		"limit":  10,
		"fields": []string{"event_time", "event_type", "extra_data"},
	}

	paramsJSON, _ := json.Marshal(params)

	// Create request
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: json.RawMessage(paramsJSON),
		},
	}

	// Call the handler
	result, err := AdAccount_GET_activitiesHandler(context.Background(), request)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}

	// Check result
	if result == nil {
		t.Error("Handler returned nil result")
	} else {
		// The result contains the JSON response
		t.Logf("Result received successfully")
	}
}

func TestToolRegistration(t *testing.T) {
	// This just tests that we can compile and the registration doesn't panic
	// In a real test, you'd create a mock server
	t.Log("Tool registration test - compilation check only")
}

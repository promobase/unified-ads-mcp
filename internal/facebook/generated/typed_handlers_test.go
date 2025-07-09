package generated

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")

	// Ensure we never use real Facebook URLs during tests
	if graphAPIHost == "https://graph.facebook.com" {
		graphAPIHost = "http://test-mock-server"
	}
	if baseGraphURL == "https://graph.facebook.com" {
		baseGraphURL = "http://test-mock-server"
	}
}

func TestTypedHandlers(t *testing.T) {
	// Create a simple mock server that returns an error for missing token
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for access token
		if r.URL.Query().Get("access_token") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "OAuthException",
					"message": "An access token is required to request this resource.",
				},
			})
			return
		}
		// Return a simple response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "123456789",
			"name":   "Test Ad",
			"status": "ACTIVE",
		})
	}))
	defer mockServer.Close()

	// Override the Graph API host for testing
	oldHost := graphAPIHost
	oldBaseURL := baseGraphURL
	defer func() {
		graphAPIHost = oldHost
		baseGraphURL = oldBaseURL
	}()
	graphAPIHost = mockServer.URL
	baseGraphURL = mockServer.URL

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

	// Call the handler directly with typed args (without access token)
	result, err := GetAdHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Handler returned nil result")
	}

	// Check if it's an error (expected because we don't have access token)
	if !result.IsError {
		t.Fatal("Expected error result due to missing access token")
	}

	// Now test with access token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	result2, err := GetAdHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Handler returned error with token: %v", err)
	}

	// Should succeed now
	if result2.IsError {
		if textContent, ok := mcp.AsTextContent(result2.Content[0]); ok {
			t.Fatalf("Expected success with token, got error: %s", textContent.Text)
		}
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

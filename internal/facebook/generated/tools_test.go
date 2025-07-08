package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// mockGraphAPIServer creates a test server that mocks Facebook Graph API responses
func mockGraphAPIServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request for debugging
		t.Logf("Mock API request: %s %s", r.Method, r.URL.Path)

		// Extract access token to verify it's being sent
		accessToken := r.URL.Query().Get("access_token")
		if accessToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "OAuthException",
					"message": "An access token is required to request this resource.",
				},
			})
			return
		}

		// Mock different endpoints
		switch {
		case strings.Contains(r.URL.Path, "/activities"):
			// Mock activities endpoint
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"event_time": "2024-01-01T12:00:00+0000",
						"event_type": "update_status",
						"extra_data": map[string]interface{}{
							"old_status": "ACTIVE",
							"new_status": "PAUSED",
						},
					},
					{
						"event_time": "2024-01-01T11:00:00+0000",
						"event_type": "create_campaign",
						"extra_data": map[string]interface{}{
							"campaign_id": "123456789",
						},
					},
				},
				"paging": map[string]interface{}{
					"cursors": map[string]string{
						"before": "BEFORE_CURSOR",
						"after":  "AFTER_CURSOR",
					},
				},
			})

		case strings.Contains(r.URL.Path, "/insights"):
			// Mock insights endpoint
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"date_start":   "2024-01-01",
						"date_stop":    "2024-01-01",
						"impressions":  "1000",
						"clicks":       "50",
						"spend":        "25.50",
						"reach":        "800",
						"frequency":    "1.25",
						"account_id":   "act_123456789",
						"account_name": "Test Account",
					},
				},
				"paging": map[string]interface{}{
					"cursors": map[string]string{
						"before": "BEFORE_CURSOR",
						"after":  "AFTER_CURSOR",
					},
				},
			})

		case r.Method == "POST" && strings.Contains(r.URL.Path, "/adlabels"):
			// Mock creating adlabels
			body, _ := io.ReadAll(r.Body)
			var params map[string]interface{}
			json.Unmarshal(body, &params)
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"adlabels": params["adlabels"],
			})

		case r.Method == "DELETE":
			// Mock delete operations
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
			})

		default:
			// Default response for unhandled endpoints
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "123456789",
				"name": "Test Object",
			})
		}
	}))
}

func TestListAdAccountActivitiesHandler_Success(t *testing.T) {
	// Create mock server
	mockServer := mockGraphAPIServer(t)
	defer mockServer.Close()

	// Override the Graph API host for testing
	oldHost := graphAPIHost
	defer func() { graphAPIHost = oldHost }()
	graphAPIHost = mockServer.URL

	// Set a test access token
	oldToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token_123")
	defer func() {
		if oldToken != "" {
			os.Setenv("FACEBOOK_ACCESS_TOKEN", oldToken)
		} else {
			os.Unsetenv("FACEBOOK_ACCESS_TOKEN")
		}
	}()

	// Create test request
	params := map[string]interface{}{
		"id":     "act_123456789",
		"limit":  10,
		"fields": []string{"event_time", "event_type", "extra_data"},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: params,
		},
	}

	// Call the handler
	result, err := ListAdAccountActivitiesHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Handler returned nil result")
	}
	
	// Debug: Check if it's an error
	if result.IsError {
		textContent, _ := mcp.AsTextContent(result.Content[0])
		t.Fatalf("Handler returned error result: %s", textContent.Text)
	}

	// Parse the result to verify it's valid JSON
	var responseData map[string]interface{}
	// Extract text content from the first content item
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in result")
	}
	resultText := textContent.Text
	if err := json.Unmarshal([]byte(resultText), &responseData); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	// Verify the response structure
	if data, ok := responseData["data"].([]interface{}); !ok {
		t.Error("Response missing 'data' field")
	} else if len(data) != 2 {
		t.Errorf("Expected 2 activities, got %d", len(data))
	}

	t.Logf("Successfully received activities: %s", resultText)
}

func TestListAdAccountActivitiesHandler_NoAccessToken(t *testing.T) {
	// Ensure no access token is set
	oldToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	os.Unsetenv("FACEBOOK_ACCESS_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("FACEBOOK_ACCESS_TOKEN", oldToken)
		}
	}()

	// Create test request
	params := map[string]interface{}{
		"id": "act_123456789",
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: params,
		},
	}

	// Call the handler
	result, err := ListAdAccountActivitiesHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify error result
	if result == nil {
		t.Fatal("Handler returned nil result")
	}

	// Check that it's an error result
	if !result.IsError {
		t.Error("Expected error result for missing access token")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in error result")
	}
	if !strings.Contains(textContent.Text, "FACEBOOK_ACCESS_TOKEN") {
		t.Errorf("Error message should mention missing access token, got: %s", textContent.Text)
	}
}

func TestGetAdSetInsightsHandler_Success(t *testing.T) {
	// Create mock server
	mockServer := mockGraphAPIServer(t)
	defer mockServer.Close()

	// Override the Graph API host for testing
	oldHost := graphAPIHost
	defer func() { graphAPIHost = oldHost }()
	graphAPIHost = mockServer.URL

	// Set a test access token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token_123")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	// Create test request for insights
	params := map[string]interface{}{
		"id":          "123456789",
		"date_preset": "yesterday",
		"fields":      []string{"impressions", "clicks", "spend", "reach", "frequency"},
		"level":       "adset",
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: params,
		},
	}

	// Call the handler
	result, err := GetAdSetInsightsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Handler returned nil result")
	}

	// Parse and verify the insights data
	var responseData map[string]interface{}
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in result")
	}
	resultText := textContent.Text
	if err := json.Unmarshal([]byte(resultText), &responseData); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	// Verify insights data
	if data, ok := responseData["data"].([]interface{}); !ok {
		t.Error("Response missing 'data' field")
	} else if len(data) > 0 {
		insight := data[0].(map[string]interface{})
		// Verify expected fields
		expectedFields := []string{"impressions", "clicks", "spend", "reach", "frequency"}
		for _, field := range expectedFields {
			if _, exists := insight[field]; !exists {
				t.Errorf("Insight missing expected field: %s", field)
			}
		}
	}

	t.Logf("Successfully received insights: %s", resultText)
}

func TestCreateAdSetAdlabelHandler_Success(t *testing.T) {
	// Create mock server
	mockServer := mockGraphAPIServer(t)
	defer mockServer.Close()

	// Override the Graph API host for testing
	oldHost := graphAPIHost
	defer func() { graphAPIHost = oldHost }()
	graphAPIHost = mockServer.URL

	// Set a test access token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token_123")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	// Create test request for creating adlabels
	params := map[string]interface{}{
		"id": "123456789",
		"adlabels": []map[string]interface{}{
			{
				"name": "Test Label",
				"id":   "label_123",
			},
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: params,
		},
	}

	// Call the handler
	result, err := CreateAdSetAdlabelHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Handler returned nil result")
	}

	// Parse and verify the response
	var responseData map[string]interface{}
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in result")
	}
	resultText := textContent.Text
	if err := json.Unmarshal([]byte(resultText), &responseData); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	// Verify success
	if success, ok := responseData["success"].(bool); !ok || !success {
		t.Error("Expected success=true in response")
	}

	t.Logf("Successfully created adlabels: %s", resultText)
}

func TestAPIErrorHandling(t *testing.T) {
	// Create a mock server that returns API errors
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "GraphMethodException",
				"message": "Invalid parameter",
				"code":    100,
			},
		})
	}))
	defer errorServer.Close()

	// Override the Graph API host
	oldHost := graphAPIHost
	defer func() { graphAPIHost = oldHost }()
	graphAPIHost = errorServer.URL

	// Set a test access token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token_123")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	// Create test request
	params := map[string]interface{}{
		"id": "invalid_id",
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: params,
		},
	}

	// Call the handler
	result, err := GetAdSetHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify error result
	if result == nil {
		t.Fatal("Handler returned nil result")
	}

	// Check that it's an error result
	if !result.IsError {
		t.Error("Expected error result for API error")
	}

	// Verify error message contains API error details
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in error result")
	}
	errorMsg := textContent.Text
	if !strings.Contains(errorMsg, "GraphMethodException") || !strings.Contains(errorMsg, "Invalid parameter") {
		t.Errorf("Error message should contain API error details, got: %s", errorMsg)
	}
}

// TestToolResultFormat verifies that successful and error results follow the correct MCP format
func TestToolResultFormat(t *testing.T) {
	// Test data
	successResponse := []byte(`{"data": [{"id": "123", "name": "Test"}]}`)
	errorMessage := "Test error message"

	// Test successful result format
	t.Run("Success Result Format", func(t *testing.T) {
		// Simulate what handlers do for success
		result := mcp.NewToolResultText(string(successResponse))
		
		if result.IsError {
			t.Error("Success result should not be marked as error")
		}
		
		if len(result.Content) != 1 {
			t.Errorf("Expected 1 content item, got %d", len(result.Content))
		}
		
		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content in result")
		}
		
		if textContent.Type != "text" {
			t.Errorf("Expected content type 'text', got %s", textContent.Type)
		}
		
		if textContent.Text != string(successResponse) {
			t.Error("Content text doesn't match expected response")
		}
	})

	// Test error result format
	t.Run("Error Result Format", func(t *testing.T) {
		// Simulate what handlers do for errors
		result := mcp.NewToolResultErrorf("API request failed: %v", fmt.Errorf("%s", errorMessage))
		
		if !result.IsError {
			t.Error("Error result should be marked as error")
		}
		
		if len(result.Content) != 1 {
			t.Errorf("Expected 1 content item, got %d", len(result.Content))
		}
		
		textContent, ok := mcp.AsTextContent(result.Content[0])
		if !ok {
			t.Error("Expected text content in error result")
		}
		
		expectedError := fmt.Sprintf("API request failed: %s", errorMessage)
		if textContent.Text != expectedError {
			t.Errorf("Error message mismatch: got %s, want %s", textContent.Text, expectedError)
		}
	})
}

// Benchmark test to ensure performance
func BenchmarkListAdAccountActivitiesHandler(b *testing.B) {
	// Create mock server
	mockServer := mockGraphAPIServer(&testing.T{})
	defer mockServer.Close()

	// Override the Graph API host
	oldHost := graphAPIHost
	defer func() { graphAPIHost = oldHost }()
	graphAPIHost = mockServer.URL

	// Set access token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token_123")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	// Prepare request
	params := map[string]interface{}{
		"id":     "act_123456789",
		"limit":  10,
		"fields": []string{"event_time", "event_type", "extra_data"},
	}
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: params,
		},
	}

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ListAdAccountActivitiesHandler(context.Background(), request)
	}
}
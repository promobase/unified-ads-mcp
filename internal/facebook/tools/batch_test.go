package tools

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
}

// mockBatchAPIServer creates a test server that handles batch requests
func mockBatchAPIServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Check access token
		if r.FormValue("access_token") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Missing access token",
				},
			})
			return
		}

		// Get batch parameter
		batchJSON := r.FormValue("batch")
		if batchJSON == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Missing batch parameter",
				},
			})
			return
		}

		// Parse batch requests
		var batchRequests []BatchRequest
		if err := json.Unmarshal([]byte(batchJSON), &batchRequests); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Invalid batch JSON",
				},
			})
			return
		}

		// Create mock responses
		responses := make([]BatchResponse, len(batchRequests))
		for i := range batchRequests {
			// Mock successful response for each request
			mockData := map[string]interface{}{
				"id":   "mock_id_" + string(rune(i)),
				"name": "Mock Object " + string(rune(i)),
			}
			mockBody, _ := json.Marshal(mockData)
			responses[i] = BatchResponse{
				Code: 200,
				Body: mockBody,
			}
		}

		// Return batch responses
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	}))
}

// TestBatchGetObjectsHandler tests the batch get objects handler
func TestBatchGetObjectsHandler(t *testing.T) {
	// Create mock server
	server := mockBatchAPIServer(t)
	defer server.Close()

	// Override the graph API host for testing
	originalHost := graphAPIHost
	SetGraphAPIHost(server.URL)
	defer SetGraphAPIHost(originalHost)

	// Set test access token
	SetAccessToken("test_token")

	// Create test request
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"object_ids": []string{"act_123", "campaign_456"},
				"fields":     []string{"id", "name"},
			},
		},
	}

	// Execute handler
	result, err := BatchGetObjectsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if result.IsError {
		t.Errorf("Expected success, got error: %v", result.Content)
	}

	// Parse result
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content")
	}

	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &resultData); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify structure
	if _, ok := resultData["objects"]; !ok {
		t.Error("Expected 'objects' field in result")
	}

	if _, ok := resultData["total_objects"]; !ok {
		t.Error("Expected 'total_objects' field in result")
	}
}

// TestBatchUpdateCampaignsHandler tests the batch update campaigns handler
func TestBatchUpdateCampaignsHandler(t *testing.T) {
	// Create mock server
	server := mockBatchAPIServer(t)
	defer server.Close()

	// Override the graph API host for testing
	originalHost := graphAPIHost
	SetGraphAPIHost(server.URL)
	defer SetGraphAPIHost(originalHost)

	// Set test access token
	SetAccessToken("test_token")

	// Create test request
	updates := map[string]map[string]interface{}{
		"campaign_123": {
			"name":   "Updated Campaign 1",
			"status": "ACTIVE",
		},
		"campaign_456": {
			"name":   "Updated Campaign 2",
			"status": "PAUSED",
		},
	}
	updatesJSON, _ := json.Marshal(updates)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"updates": string(updatesJSON),
			},
		},
	}

	// Execute handler
	result, err := BatchUpdateCampaignsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if result.IsError {
		t.Errorf("Expected success, got error: %v", result.Content)
	}

	// Parse result
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content")
	}

	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &resultData); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify structure
	if _, ok := resultData["campaigns"]; !ok {
		t.Error("Expected 'campaigns' field in result")
	}

	if _, ok := resultData["total_campaigns"]; !ok {
		t.Error("Expected 'total_campaigns' field in result")
	}
}

// TestBatchGetInsightsHandler tests the batch get insights handler
func TestBatchGetInsightsHandler(t *testing.T) {
	// Create mock server
	server := mockBatchAPIServer(t)
	defer server.Close()

	// Override the graph API host for testing
	originalHost := graphAPIHost
	SetGraphAPIHost(server.URL)
	defer SetGraphAPIHost(originalHost)

	// Set test access token
	SetAccessToken("test_token")

	// Create test request
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"object_ids":  []string{"campaign_123", "adset_456"},
				"date_preset": "yesterday",
				"fields":      []string{"impressions", "clicks", "spend"},
				"level":       "campaign",
			},
		},
	}

	// Execute handler
	result, err := BatchGetInsightsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result
	if result.IsError {
		t.Errorf("Expected success, got error: %v", result.Content)
	}

	// Parse result
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content")
	}

	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &resultData); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify structure
	if _, ok := resultData["insights"]; !ok {
		t.Error("Expected 'insights' field in result")
	}

	if _, ok := resultData["total_objects"]; !ok {
		t.Error("Expected 'total_objects' field in result")
	}

	if _, ok := resultData["date_preset"]; !ok {
		t.Error("Expected 'date_preset' field in result")
	}
}

// TestBatchToolsValidation tests input validation
func TestBatchToolsValidation(t *testing.T) {
	tests := []struct {
		name        string
		handler     func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		request     mcp.CallToolRequest
		expectError bool
	}{
		{
			name:    "BatchGetObjects_EmptyObjectIDs",
			handler: BatchGetObjectsHandler,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"object_ids": []string{},
					},
				},
			},
			expectError: true,
		},
		{
			name:    "BatchGetObjects_TooManyObjectIDs",
			handler: BatchGetObjectsHandler,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"object_ids": func() []string {
							ids := make([]string, 51)
							for i := range ids {
								ids[i] = "id_" + string(rune(i))
							}
							return ids
						}(),
					},
				},
			},
			expectError: true,
		},
		{
			name:    "BatchUpdateCampaigns_EmptyUpdates",
			handler: BatchUpdateCampaignsHandler,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"updates": "",
					},
				},
			},
			expectError: true,
		},
		{
			name:    "BatchGetInsights_EmptyObjectIDs",
			handler: BatchGetInsightsHandler,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"object_ids": []string{},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.handler(context.Background(), tt.request)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if tt.expectError && !result.IsError {
				t.Error("Expected error but got success")
			} else if !tt.expectError && result.IsError {
				t.Errorf("Expected success but got error: %v", result.Content)
			}
		})
	}
}

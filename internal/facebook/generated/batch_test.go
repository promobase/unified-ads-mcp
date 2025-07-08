package generated

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

// TestBatchRequest tests the basic batch request functionality
func TestBatchRequest(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a POST request
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/x-www-form-urlencoded") {
			t.Errorf("Expected form content type, got %s", contentType)
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Check access token
		accessToken := r.FormValue("access_token")
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

		// Get batch parameter
		batchJSON := r.FormValue("batch")
		if batchJSON == "" {
			t.Error("Missing batch parameter")
		}

		// Parse batch requests
		var requests []BatchRequest
		if err := json.Unmarshal([]byte(batchJSON), &requests); err != nil {
			t.Errorf("Failed to parse batch requests: %v", err)
		}

		// Create responses matching requests
		responses := make([]BatchResponse, len(requests))
		for i, req := range requests {
			switch req.Method {
			case "GET":
				if strings.Contains(req.RelativeURL, "act_123") {
					responses[i] = BatchResponse{
						Code: 200,
						Body: json.RawMessage(`{"id":"act_123","name":"Test Account"}`),
					}
				} else if strings.Contains(req.RelativeURL, "campaign_456") {
					responses[i] = BatchResponse{
						Code: 200,
						Body: json.RawMessage(`{"id":"campaign_456","name":"Test Campaign","status":"ACTIVE"}`),
					}
				} else {
					responses[i] = BatchResponse{
						Code: 404,
						Body: json.RawMessage(`{"error":{"message":"Object not found","type":"GraphMethodException","code":100}}`),
					}
				}
			case "POST":
				responses[i] = BatchResponse{
					Code: 200,
					Body: json.RawMessage(`{"success":true}`),
				}
			case "DELETE":
				responses[i] = BatchResponse{
					Code: 200,
					Body: json.RawMessage(`{"success":true}`),
				}
			}
		}

		// Return batch responses
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	}))
	defer mockServer.Close()

	// Override API host
	oldHost := graphAPIHost
	graphAPIHost = mockServer.URL
	defer func() { graphAPIHost = oldHost }()

	// Set test token
	t.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")

	// Test cases
	t.Run("Simple Batch Request", func(t *testing.T) {
		requests := []BatchRequest{
			{
				Method:      "GET",
				RelativeURL: "act_123?fields=id,name",
			},
			{
				Method:      "GET",
				RelativeURL: "campaign_456?fields=id,name,status",
			},
		}

		responses, err := MakeBatchRequest(requests)
		if err != nil {
			t.Fatalf("Batch request failed: %v", err)
		}

		if len(responses) != 2 {
			t.Fatalf("Expected 2 responses, got %d", len(responses))
		}

		// Verify first response
		if responses[0].Code != 200 {
			t.Errorf("Expected code 200, got %d", responses[0].Code)
		}

		var account map[string]interface{}
		if err := json.Unmarshal(responses[0].Body, &account); err != nil {
			t.Errorf("Failed to parse account response: %v", err)
		}
		if account["id"] != "act_123" {
			t.Errorf("Expected account id act_123, got %v", account["id"])
		}

		// Verify second response
		if responses[1].Code != 200 {
			t.Errorf("Expected code 200, got %d", responses[1].Code)
		}

		var campaign map[string]interface{}
		if err := json.Unmarshal(responses[1].Body, &campaign); err != nil {
			t.Errorf("Failed to parse campaign response: %v", err)
		}
		if campaign["id"] != "campaign_456" {
			t.Errorf("Expected campaign id campaign_456, got %v", campaign["id"])
		}
	})

	t.Run("Mixed Methods Batch", func(t *testing.T) {
		requests := []BatchRequest{
			{
				Method:      "GET",
				RelativeURL: "act_123",
			},
			{
				Method:      "POST",
				RelativeURL: "campaign_456",
				Body: map[string]interface{}{
					"status": "PAUSED",
				},
			},
			{
				Method:      "DELETE",
				RelativeURL: "adset_789",
			},
		}

		responses, err := MakeBatchRequest(requests)
		if err != nil {
			t.Fatalf("Batch request failed: %v", err)
		}

		if len(responses) != 3 {
			t.Fatalf("Expected 3 responses, got %d", len(responses))
		}

		// All should be successful
		for i, resp := range responses {
			if resp.Code != 200 {
				t.Errorf("Request %d failed with code %d", i, resp.Code)
			}
		}
	})

	t.Run("Error Handling", func(t *testing.T) {
		requests := []BatchRequest{
			{
				Method:      "GET",
				RelativeURL: "invalid_object",
			},
		}

		responses, err := MakeBatchRequest(requests)
		if err != nil {
			t.Fatalf("Batch request failed: %v", err)
		}

		if len(responses) != 1 {
			t.Fatalf("Expected 1 response, got %d", len(responses))
		}

		if responses[0].Code != 404 {
			t.Errorf("Expected error code 404, got %d", responses[0].Code)
		}

		var errorResp map[string]interface{}
		if err := json.Unmarshal(responses[0].Body, &errorResp); err != nil {
			t.Errorf("Failed to parse error response: %v", err)
		}
		if _, ok := errorResp["error"]; !ok {
			t.Error("Expected error field in response")
		}
	})

	t.Run("Batch Request Limit", func(t *testing.T) {
		// Test that we enforce the 50 request limit
		requests := make([]BatchRequest, 51)
		for i := range requests {
			requests[i] = BatchRequest{
				Method:      "GET",
				RelativeURL: "test",
			}
		}

		_, err := MakeBatchRequest(requests)
		if err == nil {
			t.Error("Expected error for exceeding batch limit")
		}
		if !strings.Contains(err.Error(), "50") {
			t.Errorf("Error should mention 50 request limit: %v", err)
		}
	})

	t.Run("Empty Batch Request", func(t *testing.T) {
		_, err := MakeBatchRequest([]BatchRequest{})
		if err == nil {
			t.Error("Expected error for empty batch")
		}
	})
}

// TestBatchRequestBuilder tests the batch request builder
func TestBatchRequestBuilder(t *testing.T) {
	t.Run("Builder Methods", func(t *testing.T) {
		builder := NewBatchRequestBuilder()

		// Add various requests
		builder.
			AddGET("act_123", "", map[string]interface{}{"fields": "id,name"}, "get_account").
			AddGET("campaign_456", "insights", map[string]interface{}{"date_preset": "yesterday"}, "get_insights").
			AddPOST("campaign_789", "", map[string]interface{}{"status": "PAUSED"}, "update_campaign").
			AddDELETE("adset_111", "", "delete_adset")

		requests := builder.GetRequests()
		if len(requests) != 4 {
			t.Fatalf("Expected 4 requests, got %d", len(requests))
		}

		// Verify GET request
		if requests[0].Method != "GET" {
			t.Errorf("Expected GET, got %s", requests[0].Method)
		}
		if !strings.Contains(requests[0].RelativeURL, "act_123") {
			t.Errorf("Expected act_123 in URL, got %s", requests[0].RelativeURL)
		}
		if !strings.Contains(requests[0].RelativeURL, "fields=id%2Cname") {
			t.Errorf("Expected fields parameter in URL, got %s", requests[0].RelativeURL)
		}

		// Verify GET with endpoint
		if requests[1].Method != "GET" {
			t.Errorf("Expected GET, got %s", requests[1].Method)
		}
		if !strings.Contains(requests[1].RelativeURL, "campaign_456/insights") {
			t.Errorf("Expected campaign_456/insights in URL, got %s", requests[1].RelativeURL)
		}

		// Verify POST request
		if requests[2].Method != "POST" {
			t.Errorf("Expected POST, got %s", requests[2].Method)
		}
		if requests[2].Body["status"] != "PAUSED" {
			t.Errorf("Expected status=PAUSED in body, got %v", requests[2].Body)
		}

		// Verify DELETE request
		if requests[3].Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", requests[3].Method)
		}
		if requests[3].RelativeURL != "adset_111" {
			t.Errorf("Expected adset_111 URL, got %s", requests[3].RelativeURL)
		}
	})

	t.Run("Query Parameter Encoding", func(t *testing.T) {
		builder := NewBatchRequestBuilder()
		builder.AddGET("test", "", map[string]interface{}{
			"fields": "id,name,status",
			"limit":  10,
			"after":  "cursor_123",
		}, "test")

		requests := builder.GetRequests()
		if len(requests) != 1 {
			t.Fatalf("Expected 1 request, got %d", len(requests))
		}

		// Parse the URL to check parameters
		parsedURL, err := url.Parse("http://example.com/" + requests[0].RelativeURL)
		if err != nil {
			t.Fatalf("Failed to parse URL: %v", err)
		}

		query := parsedURL.Query()
		if query.Get("fields") != "id,name,status" {
			t.Errorf("Expected fields parameter, got %s", query.Get("fields"))
		}
		if query.Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", query.Get("limit"))
		}
		if query.Get("after") != "cursor_123" {
			t.Errorf("Expected after=cursor_123, got %s", query.Get("after"))
		}
	})
}

// TestBatchRequestIntegration tests batch requests with the mock Graph API
func TestBatchRequestIntegration(t *testing.T) {
	// Create mock server for integration test
	mockServer := mockGraphAPIServer(t)
	defer mockServer.Close()

	// Override API host
	oldHost := graphAPIHost
	graphAPIHost = mockServer.URL
	defer func() { graphAPIHost = oldHost }()

	// Set test token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	t.Run("Batch Get Multiple Objects", func(t *testing.T) {
		builder := NewBatchRequestBuilder()
		builder.
			AddGET("act_123456789", "activities", map[string]interface{}{"limit": 5}, "activities").
			AddGET("123456789", "insights", map[string]interface{}{"date_preset": "yesterday"}, "insights")

		responses, err := builder.Execute()
		if err != nil {
			t.Fatalf("Batch execution failed: %v", err)
		}

		if len(responses) != 2 {
			t.Fatalf("Expected 2 responses, got %d", len(responses))
		}

		// Both should succeed
		for i, resp := range responses {
			if resp.Code != 200 {
				t.Errorf("Request %d failed with code %d", i, resp.Code)
			}
			if resp.Body == nil {
				t.Errorf("Request %d has no body", i)
			}
		}
	})
}
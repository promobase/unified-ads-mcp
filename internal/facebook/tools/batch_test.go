package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"unified-ads-mcp/internal/facebook/generated"
)

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
		var requests []generated.BatchRequest
		if err := json.Unmarshal([]byte(batchJSON), &requests); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Invalid batch JSON",
				},
			})
			return
		}

		// Create mock responses
		responses := make([]generated.BatchResponse, len(requests))
		for i, req := range requests {
			switch {
			case strings.Contains(req.RelativeURL, "insights"):
				responses[i] = generated.BatchResponse{
					Code: 200,
					Body: json.RawMessage(`{"data":[{"impressions":"1000","clicks":"50","spend":"25.50"}]}`),
				}
			case req.Method == "POST" && strings.Contains(req.RelativeURL, "adsets"):
				responses[i] = generated.BatchResponse{
					Code: 200,
					Body: json.RawMessage(`{"id":"adset_` + string(rune(i+'1')) + `","name":"Test AdSet ` + string(rune(i+'1')) + `"}`),
				}
			case req.Method == "POST":
				responses[i] = generated.BatchResponse{
					Code: 200,
					Body: json.RawMessage(`{"success":true}`),
				}
			default:
				responses[i] = generated.BatchResponse{
					Code: 200,
					Body: json.RawMessage(`{"id":"` + strings.Split(req.RelativeURL, "?")[0] + `","name":"Test Object"}`),
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	}))
}

func TestExecuteBatchRequestsHandler(t *testing.T) {
	mockServer := mockBatchAPIServer(t)
	defer mockServer.Close()

	// Override API host
	oldHost := generated.GetGraphAPIHost()
	generated.SetGraphAPIHost(mockServer.URL)
	defer func() {
		generated.SetGraphAPIHost(oldHost)
	}()

	// Set test token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	// Prepare batch requests
	batchRequests := []generated.BatchRequest{
		{
			Method:      "GET",
			RelativeURL: "act_123?fields=id,name",
		},
		{
			Method:      "POST",
			RelativeURL: "campaign_456",
			Body: map[string]interface{}{
				"status": "PAUSED",
			},
		},
	}

	requestsJSON, _ := json.Marshal(batchRequests)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"requests": string(requestsJSON),
			},
		},
	}

	// Call handler
	result, err := ExecuteBatchRequestsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		textContent, _ := mcp.AsTextContent(result.Content[0])
		t.Fatalf("Handler returned error result: %s", textContent.Text)
	}

	// Verify response
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in result")
	}

	var responses []map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &responses); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(responses))
	}

	// Verify both responses succeeded
	for i, resp := range responses {
		if code, ok := resp["code"].(float64); !ok || code != 200 {
			t.Errorf("Response %d failed with code %v", i, resp["code"])
		}
	}
}

func TestBatchGetObjectsHandler(t *testing.T) {
	mockServer := mockBatchAPIServer(t)
	defer mockServer.Close()

	// Override API host
	oldHost := generated.GetGraphAPIHost()
	generated.SetGraphAPIHost(mockServer.URL)
	defer func() {
		generated.SetGraphAPIHost(oldHost)
	}()

	// Set test token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"object_ids": []string{"act_123", "campaign_456", "adset_789"},
				"fields":     []string{"id", "name", "status"},
			},
		},
	}

	// Call handler
	result, err := BatchGetObjectsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		textContent, _ := mcp.AsTextContent(result.Content[0])
		t.Fatalf("Handler returned error result: %s", textContent.Text)
	}

	// Verify response
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in result")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify we got data for all objects
	expectedObjects := []string{"act_123", "campaign_456", "adset_789"}
	for _, objID := range expectedObjects {
		if _, ok := response[objID]; !ok {
			t.Errorf("Missing response for object %s", objID)
		}
	}
}

func TestBatchUpdateCampaignsHandler(t *testing.T) {
	mockServer := mockBatchAPIServer(t)
	defer mockServer.Close()

	// Override API host
	oldHost := generated.GetGraphAPIHost()
	generated.SetGraphAPIHost(mockServer.URL)
	defer func() {
		generated.SetGraphAPIHost(oldHost)
	}()

	// Set test token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	updates := map[string]map[string]interface{}{
		"campaign_123": {
			"status": "PAUSED",
			"name":   "Updated Campaign 1",
		},
		"campaign_456": {
			"status": "ACTIVE",
			"budget": 5000,
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

	// Call handler
	result, err := BatchUpdateCampaignsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		textContent, _ := mcp.AsTextContent(result.Content[0])
		t.Fatalf("Handler returned error result: %s", textContent.Text)
	}

	// Verify response
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in result")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify both campaigns were updated
	for campaignID, result := range response {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Errorf("Invalid result format for campaign %s", campaignID)
			continue
		}

		if success, ok := resultMap["success"].(bool); !ok || !success {
			t.Errorf("Campaign %s update failed", campaignID)
		}
	}
}

func TestBatchGetInsightsHandler(t *testing.T) {
	mockServer := mockBatchAPIServer(t)
	defer mockServer.Close()

	// Override API host
	oldHost := generated.GetGraphAPIHost()
	generated.SetGraphAPIHost(mockServer.URL)
	defer func() {
		generated.SetGraphAPIHost(oldHost)
	}()

	// Set test token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"object_ids":  []string{"act_123", "campaign_456"},
				"level":       "campaign",
				"date_preset": "yesterday",
				"fields":      []string{"impressions", "clicks", "spend"},
			},
		},
	}

	// Call handler
	result, err := BatchGetInsightsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		textContent, _ := mcp.AsTextContent(result.Content[0])
		t.Fatalf("Handler returned error result: %s", textContent.Text)
	}

	// Verify response
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in result")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify we got insights for both objects
	for _, objID := range []string{"act_123", "campaign_456"} {
		if _, ok := response[objID]; !ok {
			t.Errorf("Missing insights for object %s", objID)
		}
	}
}

func TestBatchCreateAdsetsHandler(t *testing.T) {
	mockServer := mockBatchAPIServer(t)
	defer mockServer.Close()

	// Override API host
	oldHost := generated.GetGraphAPIHost()
	generated.SetGraphAPIHost(mockServer.URL)
	defer func() {
		generated.SetGraphAPIHost(oldHost)
	}()

	// Set test token
	os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

	adsets := []map[string]interface{}{
		{
			"name":              "Test AdSet 1",
			"daily_budget":      1000,
			"targeting":         map[string]interface{}{"geo_locations": map[string]interface{}{"countries": []string{"US"}}},
			"billing_event":     "IMPRESSIONS",
			"optimization_goal": "REACH",
		},
		{
			"name":              "Test AdSet 2",
			"daily_budget":      2000,
			"targeting":         map[string]interface{}{"geo_locations": map[string]interface{}{"countries": []string{"CA"}}},
			"billing_event":     "IMPRESSIONS",
			"optimization_goal": "REACH",
		},
	}

	adsetsJSON, _ := json.Marshal(adsets)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"campaign_id": "123456789",
				"adsets":      string(adsetsJSON),
			},
		},
	}

	// Call handler
	result, err := BatchCreateAdsetsHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		textContent, _ := mcp.AsTextContent(result.Content[0])
		t.Fatalf("Handler returned error result: %s", textContent.Text)
	}

	// Verify response
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("Expected text content in result")
	}

	var response []map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 ad sets created, got %d", len(response))
	}

	// Verify all creations succeeded
	for i, result := range response {
		if success, ok := result["success"].(bool); !ok || !success {
			t.Errorf("AdSet %d creation failed", i)
		}
	}
}

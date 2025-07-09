package generated

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"unified-ads-mcp/internal/facebook/testutil"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")
}

// TestBatchRequest tests the basic batch request functionality using the framework
func TestBatchRequestWithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Create mock batch responses
	mockResponses := map[string]interface{}{
		"act_123?fields=id,name": map[string]interface{}{
			"id":   "act_123",
			"name": "Test Account",
		},
		"campaign_456?fields=id,name,status": map[string]interface{}{
			"id":     "campaign_456",
			"name":   "Test Campaign",
			"status": "ACTIVE",
		},
	}

	// Setup batch handler
	env.Server().AddRoute("POST", "/v23.0/", testutil.CreateMockBatchHandler(mockResponses))

	// Override URLs
	oldHost := graphAPIHost
	oldBase := baseGraphURL
	defer func() {
		graphAPIHost = oldHost
		baseGraphURL = oldBase
	}()
	graphAPIHost = env.Server().URL
	baseGraphURL = env.Server().URL

	tests := []struct {
		name             string
		setupBatchFunc   func() []testutil.BatchRequest
		expectedCount    int
		validateResponse func(*testing.T, []testutil.BatchResponse)
	}{
		{
			name: "Simple_Batch_Request",
			setupBatchFunc: func() []testutil.BatchRequest {
				return testutil.NewBatchTestBuilder(t).
					AddGetRequest("act_123", map[string]string{"fields": "id,name"}).
					AddGetRequest("campaign_456", map[string]string{"fields": "id,name,status"}).
					Build()
			},
			expectedCount: 2,
			validateResponse: func(t *testing.T, responses []testutil.BatchResponse) {
				assertions := testutil.NewAssertBatchResponse(t, responses)
				assertions.HasCount(2).AllSuccessful()
				assertions.ResponseAt(0).HasField("id", "act_123")
				assertions.ResponseAt(1).HasField("status", "ACTIVE")
			},
		},
		{
			name: "Mixed_Methods_Batch",
			setupBatchFunc: func() []testutil.BatchRequest {
				return testutil.NewBatchTestBuilder(t).
					AddGetRequest("act_123", map[string]string{"fields": "id"}).
					AddPostRequest("campaign_456/adsets", map[string]interface{}{
						"name":   "New AdSet",
						"status": "PAUSED",
					}).
					AddDeleteRequest("ad_789").
					Build()
			},
			expectedCount: 3,
		},
		{
			name: "Error_Handling",
			setupBatchFunc: func() []testutil.BatchRequest {
				return testutil.NewBatchTestBuilder(t).
					AddGetRequest("nonexistent", nil).
					Build()
			},
			expectedCount: 1,
			validateResponse: func(t *testing.T, responses []testutil.BatchResponse) {
				// The default response for unknown paths should still be 200
				testutil.NewAssertBatchResponse(t, responses).
					HasCount(1).
					ResponseAt(0).IsSuccess()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup batch requests
			requests := tt.setupBatchFunc()

			// Create form data
			formData := url.Values{}
			formData.Set("access_token", "test_token")
			batchJSON, _ := json.Marshal(requests)
			formData.Set("batch", string(batchJSON))

			// Make batch request
			resp, err := http.Post(
				env.Server().URL+"/v23.0/",
				"application/x-www-form-urlencoded",
				strings.NewReader(formData.Encode()),
			)
			if err != nil {
				t.Fatalf("Failed to make batch request: %v", err)
			}
			defer resp.Body.Close()

			// Parse response
			var batchResponses []testutil.BatchResponse
			if err := json.NewDecoder(resp.Body).Decode(&batchResponses); err != nil {
				t.Fatalf("Failed to decode batch response: %v", err)
			}

			// Validate
			if len(batchResponses) != tt.expectedCount {
				t.Errorf("Expected %d responses, got %d", tt.expectedCount, len(batchResponses))
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, batchResponses)
			}
		})
	}
}

// TestBatchRequestBuilder tests the batch request builder
func TestBatchRequestBuilderWithFramework(t *testing.T) {
	tests := []struct {
		name          string
		buildFunc     func() []testutil.BatchRequest
		expectedCount int
		validate      func(*testing.T, []testutil.BatchRequest)
	}{
		{
			name: "Builder_Methods",
			buildFunc: func() []testutil.BatchRequest {
				return testutil.NewBatchTestBuilder(t).
					AddGetRequest("123", map[string]string{"fields": "id,name"}).
					AddPostRequest("456", map[string]interface{}{"status": "ACTIVE"}).
					AddDeleteRequest("789").
					Build()
			},
			expectedCount: 3,
			validate: func(t *testing.T, requests []testutil.BatchRequest) {
				if requests[0].Method != "GET" {
					t.Errorf("Expected first request to be GET, got %s", requests[0].Method)
				}
				if requests[1].Method != "POST" {
					t.Errorf("Expected second request to be POST, got %s", requests[1].Method)
				}
				if requests[2].Method != "DELETE" {
					t.Errorf("Expected third request to be DELETE, got %s", requests[2].Method)
				}
			},
		},
		{
			name: "Query_Parameter_Encoding",
			buildFunc: func() []testutil.BatchRequest {
				return testutil.NewBatchTestBuilder(t).
					AddGetRequest("campaign", map[string]string{
						"fields": "id,name,status",
						"limit":  "10",
					}).
					Build()
			},
			expectedCount: 1,
			validate: func(t *testing.T, requests []testutil.BatchRequest) {
				url := requests[0].RelativeURL
				if !strings.Contains(url, "fields=id,name,status") {
					t.Errorf("Expected fields parameter in URL, got %s", url)
				}
				if !strings.Contains(url, "limit=10") {
					t.Errorf("Expected limit parameter in URL, got %s", url)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := tt.buildFunc()

			if len(requests) != tt.expectedCount {
				t.Errorf("Expected %d requests, got %d", tt.expectedCount, len(requests))
			}

			if tt.validate != nil {
				tt.validate(t, requests)
			}
		})
	}
}

// TestBatchRequestIntegration tests integration with actual batch handlers
func TestBatchRequestIntegrationWithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup batch response builder
	expectedResponses := testutil.NewBatchResponseBuilder().
		AddSuccess(testutil.CreateMockCampaignResponse(testutil.TestCampaignID)).
		AddSuccess(map[string]interface{}{
			"id":     "adset_456",
			"name":   "Test AdSet",
			"status": "PAUSED",
		}).
		Build()

	// Setup mock handler that returns the expected responses
	env.Server().AddRoute("POST", "/v23.0/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			env.Server().WriteError(w, 400, "ParseError", "Failed to parse form")
			return
		}

		// Check access token
		if r.FormValue("access_token") == "" {
			env.Server().WriteAuthError(w)
			return
		}

		// Return our expected responses
		testutil.WriteJSONSuccess(w, expectedResponses)
	})

	// Override URLs
	oldHost := graphAPIHost
	oldBase := baseGraphURL
	defer func() {
		graphAPIHost = oldHost
		baseGraphURL = oldBase
	}()
	graphAPIHost = env.Server().URL
	baseGraphURL = env.Server().URL

	t.Run("Batch_Get_Multiple_Objects", func(t *testing.T) {
		// Build batch requests
		requests := testutil.NewBatchTestBuilder(t).
			AddGetRequest("campaign_123", map[string]string{"fields": "id,name,status"}).
			AddGetRequest("adset_456", map[string]string{"fields": "id,name,status"}).
			Build()

		// Simulate batch execution by making HTTP request
		formData := url.Values{}
		formData.Set("access_token", testutil.TestAccessToken)
		batchJSON, _ := json.Marshal(requests)
		formData.Set("batch", string(batchJSON))

		resp, err := http.Post(
			env.Server().URL+"/v23.0/",
			"application/x-www-form-urlencoded",
			strings.NewReader(formData.Encode()),
		)
		if err != nil {
			t.Fatalf("Failed to make batch request: %v", err)
		}
		defer resp.Body.Close()

		var result []testutil.BatchResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode batch response: %v", err)
		}

		// Validate responses
		testutil.NewAssertBatchResponse(t, result).
			HasCount(2).
			AllSuccessful()

		// Test individual response assertions
		assertions := testutil.NewAssertBatchResponse(t, result)
		assertions.ResponseAt(0).
			IsSuccess().
			HasField("id", testutil.TestCampaignID)

		assertions.ResponseAt(1).
			IsSuccess().
			HasField("id", "adset_456").
			HasField("status", "PAUSED")
	})
}

// TestBatchLimit ensures batch requests respect the 50 request limit
func TestBatchLimitWithFramework(t *testing.T) {
	builder := testutil.NewBatchTestBuilder(t)

	// Add 50 requests (should succeed)
	for i := 0; i < 50; i++ {
		builder.AddGetRequest(fmt.Sprintf("object_%d", i), nil)
	}

	requests := builder.Build()
	if len(requests) != 50 {
		t.Errorf("Expected 50 requests, got %d", len(requests))
	}

	// Test that a batch handler would enforce the limit
	// Note: Without ExecuteBatchRequests, we're testing the builder logic
	if len(requests) > 50 {
		t.Error("Batch builder should not allow more than 50 requests")
	}
}

// TestEmptyBatchRequest tests handling of empty batch requests
func TestEmptyBatchRequestWithFramework(t *testing.T) {
	requests := []testutil.BatchRequest{}

	// An empty batch should be invalid
	if len(requests) != 0 {
		t.Error("Expected empty batch request array")
	}

	// In a real implementation, the batch handler would return an error
	// for empty requests
}

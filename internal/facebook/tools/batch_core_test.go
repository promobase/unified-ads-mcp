package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"unified-ads-mcp/internal/facebook/testutil"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")
}

func TestFacebookBatchHandler_BasicOperations(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Override URLs
	oldHost := graphAPIHost
	defer func() {
		graphAPIHost = oldHost
	}()
	graphAPIHost = env.Server().URL

	// Set up mock responses
	env.Server().AddRoute("GET", "/v23.0/"+testutil.TestCampaignID+"/", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateMockCampaignResponse(testutil.TestCampaignID))
	})

	env.Server().AddRoute("GET", "/v23.0/"+testutil.TestAdsetID+"/", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateMockAdSetResponse(testutil.TestAdsetID))
	})

	// Set up batch endpoint
	env.Server().AddRoute("POST", "/v23.0/", func(w http.ResponseWriter, r *http.Request) {
		// Create mock responses as JSON strings
		campaignResp, _ := json.Marshal(testutil.CreateMockCampaignResponse(testutil.TestCampaignID))
		adsetResp, _ := json.Marshal(testutil.CreateMockAdSetResponse(testutil.TestAdsetID))

		// Mock batch response
		response := `[
			{"code": 200, "body": ` + string(campaignResp) + `},
			{"code": 200, "body": ` + string(adsetResp) + `}
		]`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(response))
	})

	// Test batch request with multiple operations
	args := BatchRequestArgs{
		Operations: []BatchOperationArgs{
			{
				Method:      "GET",
				RelativeURL: testutil.TestCampaignID + "?fields=id,name,status",
				Name:        "get_campaign",
			},
			{
				Method:      "GET",
				RelativeURL: testutil.TestAdsetID + "?fields=id,name,status",
				Name:        "get_adset",
			},
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"operations": []map[string]interface{}{
					{
						"method":       "GET",
						"relative_url": testutil.TestCampaignID + "?fields=id,name,status",
						"name":         "get_campaign",
					},
					{
						"method":       "GET",
						"relative_url": testutil.TestAdsetID + "?fields=id,name,status",
						"name":         "get_adset",
					},
				},
			},
		},
	}

	// Execute
	result, err := FacebookBatchHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("FacebookBatchHandler failed: %v", err)
	}

	// Assert
	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	// Verify batch result structure
	if totalOps, ok := data["total_operations"]; !ok || totalOps != float64(2) {
		t.Errorf("Expected total_operations=2, got %v", totalOps)
	}

	if successOps, ok := data["successful_operations"]; !ok || successOps != float64(2) {
		t.Errorf("Expected successful_operations=2, got %v", successOps)
	}

	if failedOps, ok := data["failed_operations"]; !ok || failedOps != float64(0) {
		t.Errorf("Expected failed_operations=0, got %v", failedOps)
	}

	// Verify results array
	results, ok := data["results"].([]interface{})
	if !ok || len(results) != 2 {
		t.Fatalf("Expected 2 results, got %T %v", data["results"], data["results"])
	}

	// Check first result
	result1 := results[0].(map[string]interface{})
	if result1["success"] != true {
		t.Errorf("Expected first operation to succeed")
	}
	if result1["name"] != "get_campaign" {
		t.Errorf("Expected first operation name to be 'get_campaign', got %v", result1["name"])
	}

	// Check second result
	result2 := results[1].(map[string]interface{})
	if result2["success"] != true {
		t.Errorf("Expected second operation to succeed")
	}
	if result2["name"] != "get_adset" {
		t.Errorf("Expected second operation name to be 'get_adset', got %v", result2["name"])
	}
}

func TestFacebookBatchHandler_ErrorHandling(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Override URLs
	oldHost := graphAPIHost
	defer func() {
		graphAPIHost = oldHost
	}()
	graphAPIHost = env.Server().URL

	// Set up batch endpoint with mixed success/failure
	env.Server().AddRoute("POST", "/v23.0/", func(w http.ResponseWriter, r *http.Request) {
		// Create mock response as JSON string
		campaignResp, _ := json.Marshal(testutil.CreateMockCampaignResponse(testutil.TestCampaignID))

		// Mock batch response with one success and one error
		response := `[
			{"code": 200, "body": ` + string(campaignResp) + `},
			{"code": 400, "body": "{\"error\": {\"message\": \"Invalid object ID\", \"type\": \"OAuthException\"}}"}
		]`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(response))
	})

	// Test batch request with mixed results
	args := BatchRequestArgs{
		Operations: []BatchOperationArgs{
			{
				Method:      "GET",
				RelativeURL: testutil.TestCampaignID + "?fields=id,name",
				Name:        "valid_campaign",
			},
			{
				Method:      "GET",
				RelativeURL: "invalid_id?fields=id,name",
				Name:        "invalid_campaign",
			},
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"operations": []map[string]interface{}{
					{
						"method":       "GET",
						"relative_url": testutil.TestCampaignID + "?fields=id,name",
						"name":         "valid_campaign",
					},
					{
						"method":       "GET",
						"relative_url": "invalid_id?fields=id,name",
						"name":         "invalid_campaign",
					},
				},
			},
		},
	}

	// Execute
	result, err := FacebookBatchHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("FacebookBatchHandler failed: %v", err)
	}

	// Assert
	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	// Verify mixed results
	if successOps, ok := data["successful_operations"]; !ok || successOps != float64(1) {
		t.Errorf("Expected successful_operations=1, got %v", successOps)
	}

	if failedOps, ok := data["failed_operations"]; !ok || failedOps != float64(1) {
		t.Errorf("Expected failed_operations=1, got %v", failedOps)
	}

	results, ok := data["results"].([]interface{})
	if !ok || len(results) != 2 {
		t.Fatalf("Expected 2 results, got %T %v", data["results"], data["results"])
	}

	// Check successful operation
	result1 := results[0].(map[string]interface{})
	if result1["success"] != true {
		t.Errorf("Expected first operation to succeed")
	}

	// Check failed operation
	result2 := results[1].(map[string]interface{})
	if result2["success"] != false {
		t.Errorf("Expected second operation to fail")
	}
	if result2["error"] == nil {
		t.Errorf("Expected error message for failed operation")
	}
}

func TestFacebookBatchHandler_Validation(t *testing.T) {
	tests := []struct {
		name        string
		operations  []BatchOperationArgs
		expectError string
	}{
		{
			name:        "Empty operations",
			operations:  []BatchOperationArgs{},
			expectError: "At least one operation is required",
		},
		{
			name: "Too many operations",
			operations: func() []BatchOperationArgs {
				ops := make([]BatchOperationArgs, 51)
				for i := range ops {
					ops[i] = BatchOperationArgs{
						Method:      "GET",
						RelativeURL: "123456789",
					}
				}
				return ops
			}(),
			expectError: "Maximum 50 operations allowed per batch",
		},
		{
			name: "Valid single operation",
			operations: []BatchOperationArgs{
				{
					Method:      "GET",
					RelativeURL: testutil.TestCampaignID,
				},
			},
			expectError: "", // No error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := BatchRequestArgs{
				Operations: tt.operations,
			}

			request := mcp.CallToolRequest{}

			result, err := FacebookBatchHandler(context.Background(), request, args)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if tt.expectError != "" {
				testutil.AssertResult(t, result).
					IsError().
					HasErrorContaining(tt.expectError)
			} else {
				// For valid cases, we expect the handler to proceed
				// (though it may fail due to missing mock setup)
				content := testutil.AssertResult(t, result).GetContent()
				if strings.Contains(content, "At least one operation is required") ||
					strings.Contains(content, "Maximum 50 operations allowed per batch") {
					t.Errorf("Unexpected validation error: %v", content)
				}
			}
		})
	}
}

func TestFacebookBatchHandler_PostOperations(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Override URLs
	oldHost := graphAPIHost
	defer func() {
		graphAPIHost = oldHost
	}()
	graphAPIHost = env.Server().URL

	// Set up batch endpoint for POST operations
	env.Server().AddRoute("POST", "/v23.0/", func(w http.ResponseWriter, r *http.Request) {
		// Mock batch response for POST operations
		response := `[
			{"code": 200, "body": "{\"id\": \"12345\", \"success\": true}"},
			{"code": 200, "body": "{\"id\": \"67890\", \"success\": true}"}
		]`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(response))
	})

	// Test batch with POST operations
	args := BatchRequestArgs{
		Operations: []BatchOperationArgs{
			{
				Method:      "POST",
				RelativeURL: testutil.TestCampaignID,
				Body: map[string]interface{}{
					"name":   "Updated Campaign Name",
					"status": "ACTIVE",
				},
				Name: "update_campaign",
			},
			{
				Method:      "POST",
				RelativeURL: testutil.TestAdsetID,
				Body: map[string]interface{}{
					"name":         "Updated AdSet Name",
					"daily_budget": 1000,
				},
				Name: "update_adset",
			},
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"operations": []map[string]interface{}{
					{
						"method":       "POST",
						"relative_url": testutil.TestCampaignID,
						"body": map[string]interface{}{
							"name":   "Updated Campaign Name",
							"status": "ACTIVE",
						},
						"name": "update_campaign",
					},
					{
						"method":       "POST",
						"relative_url": testutil.TestAdsetID,
						"body": map[string]interface{}{
							"name":         "Updated AdSet Name",
							"daily_budget": 1000,
						},
						"name": "update_adset",
					},
				},
			},
		},
	}

	// Execute
	result, err := FacebookBatchHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("FacebookBatchHandler failed: %v", err)
	}

	// Assert
	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	// Verify all operations succeeded
	if successOps, ok := data["successful_operations"]; !ok || successOps != float64(2) {
		t.Errorf("Expected successful_operations=2, got %v", successOps)
	}

	if failedOps, ok := data["failed_operations"]; !ok || failedOps != float64(0) {
		t.Errorf("Expected failed_operations=0, got %v", failedOps)
	}

	results, ok := data["results"].([]interface{})
	if !ok || len(results) != 2 {
		t.Fatalf("Expected 2 results, got %T %v", data["results"], data["results"])
	}

	// Verify both POST operations succeeded
	for i, result := range results {
		resultMap := result.(map[string]interface{})
		if resultMap["success"] != true {
			t.Errorf("Expected operation %d to succeed", i)
		}
		if resultMap["code"] != float64(200) {
			t.Errorf("Expected operation %d to have code 200, got %v", i, resultMap["code"])
		}
	}
}

package generated

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"unified-ads-mcp/internal/facebook/testutil"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")
}

// Campaign Integration Tests
func TestGetCampaignIntegration_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("GET", "/v23.0/"+testutil.TestCampaignID+"/", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateMockCampaignResponse(testutil.TestCampaignID))
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

	args := get_campaignArgs{
		ID: testutil.TestCampaignID,
		Fields: []string{
			"id", "name", "status", "objective", "daily_budget",
			"lifetime_budget", "bid_strategy", "promoted_object", "adlabels",
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_campaign",
			Arguments: map[string]interface{}{
				"id":     args.ID,
				"fields": args.Fields,
			},
		},
	}

	// Execute
	result, err := GetCampaignHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("GetCampaignHandler failed: %v", err)
	}

	// Assert and validate
	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	// Validate response structure
	if data["id"] != testutil.TestCampaignID {
		t.Errorf("Expected campaign ID %s, got %v", testutil.TestCampaignID, data["id"])
	}

	if data["name"] != "Test Campaign" {
		t.Errorf("Expected name 'Test Campaign', got %v", data["name"])
	}

	if data["status"] != "PAUSED" {
		t.Errorf("Expected status 'PAUSED', got %v", data["status"])
	}

	// Validate complex types
	if promotedObject, ok := data["promoted_object"].(map[string]interface{}); ok {
		if promotedObject["page_id"] != testutil.TestPageID {
			t.Errorf("Expected promoted_object.page_id '%s', got %v", testutil.TestPageID, promotedObject["page_id"])
		}
	} else {
		t.Error("Expected promoted_object to be an object")
	}

	// Validate array types
	if adlabels, ok := data["adlabels"].([]interface{}); ok {
		if len(adlabels) != 1 {
			t.Errorf("Expected 1 adlabel, got %d", len(adlabels))
		}
	} else {
		t.Error("Expected adlabels to be an array")
	}
}

func TestUpdateCampaignWithTypedArgsIntegration_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("POST", "/v23.0/"+testutil.TestCampaignID+"/", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateSuccessResponse(testutil.TestCampaignID))
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

	// Test typed arguments with complex types
	args := update_campaignArgs{
		ID:          testutil.TestCampaignID,
		Name:        "Updated Campaign Name",
		Status:      "ACTIVE",
		DailyBudget: 300,
		BidStrategy: "COST_CAP",
		PromotedObject: &AdPromotedObject{
			PageID: "54321",
		},
		Adlabels: []*AdLabel{
			{
				ID:   "new_label_123",
				Name: "New Label",
			},
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "update_campaign",
			Arguments: map[string]interface{}{
				"id":           args.ID,
				"name":         args.Name,
				"status":       args.Status,
				"daily_budget": args.DailyBudget,
				"bid_strategy": args.BidStrategy,
				"promoted_object": map[string]interface{}{
					"page_id": "54321",
				},
				"adlabels": []map[string]interface{}{
					{
						"id":   "new_label_123",
						"name": "New Label",
					},
				},
			},
		},
	}

	// Execute
	result, err := UpdateCampaignHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("UpdateCampaignHandler failed: %v", err)
	}

	// Assert
	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	if data["id"] != testutil.TestCampaignID {
		t.Errorf("Expected campaign ID %s, got %v", testutil.TestCampaignID, data["id"])
	}

	if success, ok := data["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}
}

func TestCreateCampaignAdlabelWithValidationIntegration_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("POST", "/v23.0/"+testutil.TestCampaignID+"/adlabels", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateSuccessResponse(testutil.TestCampaignID))
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

	args := create_campaign_adlabelArgs{
		ID: testutil.TestCampaignID,
		Adlabels: []*AdLabel{
			{
				ID:   "label_001",
				Name: "Performance Label",
			},
			{
				ID:   "label_002",
				Name: "Brand Label",
			},
		},
		ExecutionOptions: []string{"include_recommendations"},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_campaign_adlabel",
			Arguments: map[string]interface{}{
				"id": args.ID,
				"adlabels": []map[string]interface{}{
					{
						"id":   "label_001",
						"name": "Performance Label",
					},
					{
						"id":   "label_002",
						"name": "Brand Label",
					},
				},
				"execution_options": args.ExecutionOptions,
			},
		},
	}

	// Execute
	result, err := CreateCampaignAdlabelHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("CreateCampaignAdlabelHandler failed: %v", err)
	}

	// Assert
	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	if !data["success"].(bool) {
		t.Error("Expected success to be true")
	}
}

// Test validation errors
func TestCampaignHandlerValidationErrors_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Override URLs
	oldHost := graphAPIHost
	oldBase := baseGraphURL
	defer func() {
		graphAPIHost = oldHost
		baseGraphURL = oldBase
	}()
	graphAPIHost = env.Server().URL
	baseGraphURL = env.Server().URL

	// Test missing required ID
	args := get_campaignArgs{
		ID:     "", // Missing required ID
		Fields: []string{"name"},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id":     args.ID,
				"fields": args.Fields,
			},
		},
	}

	// Execute
	result, err := GetCampaignHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Expected validation error in result, not handler error: %v", err)
	}

	// Assert
	testutil.AssertResult(t, result).
		IsError().
		HasErrorContaining("id is required")
}

// Ad Set Integration Tests
func TestGetAdSetWithComplexTypesIntegration_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("GET", "/v23.0/"+testutil.TestAdsetID+"/", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateMockAdSetResponse(testutil.TestAdsetID))
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

	args := get_ad_setArgs{
		ID: testutil.TestAdsetID,
		Fields: []string{
			"id", "name", "status", "targeting", "promoted_object", "adlabels",
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_ad_set",
			Arguments: map[string]interface{}{
				"id":     args.ID,
				"fields": args.Fields,
			},
		},
	}

	// Execute
	result, err := GetAdSetHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("GetAdSetHandler failed: %v", err)
	}

	// Assert
	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	// Validate targeting complex type
	if targeting, ok := data["targeting"].(map[string]interface{}); ok {
		if geoLocations, ok := targeting["geo_locations"].(map[string]interface{}); ok {
			if countries, ok := geoLocations["countries"].([]interface{}); ok {
				if len(countries) != 1 || countries[0] != "US" {
					t.Errorf("Expected targeting countries ['US'], got %v", countries)
				}
			} else {
				t.Error("Expected geo_locations.countries to be an array")
			}
		} else {
			t.Error("Expected targeting.geo_locations to be an object")
		}
	} else {
		t.Error("Expected targeting to be an object")
	}
}

// Comprehensive test covering multiple tools
func TestCampaignLifecycleIntegration_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Override URLs
	oldHost := graphAPIHost
	oldBase := baseGraphURL
	defer func() {
		graphAPIHost = oldHost
		baseGraphURL = oldBase
	}()
	graphAPIHost = env.Server().URL
	baseGraphURL = env.Server().URL

	// Test 1: Get campaign
	t.Run("GetCampaign", func(t *testing.T) {
		env.Server().AddRoute("GET", "/v23.0/"+testutil.TestCampaignID+"/", func(w http.ResponseWriter, r *http.Request) {
			env.Server().WriteSuccess(w, testutil.CreateMockCampaignResponse(testutil.TestCampaignID))
		})

		args := get_campaignArgs{
			ID:     testutil.TestCampaignID,
			Fields: []string{"id", "name", "status"},
		}

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"id":     args.ID,
					"fields": args.Fields,
				},
			},
		}

		result, err := GetCampaignHandler(context.Background(), request, args)
		if err != nil {
			t.Fatalf("Get campaign failed: %v", err)
		}

		testutil.AssertResult(t, result).IsSuccess()
	})

	// Test 2: Update campaign
	t.Run("UpdateCampaign", func(t *testing.T) {
		env.Server().AddRoute("POST", "/v23.0/"+testutil.TestCampaignID+"/", func(w http.ResponseWriter, r *http.Request) {
			env.Server().WriteSuccess(w, testutil.CreateSuccessResponse(testutil.TestCampaignID))
		})

		args := update_campaignArgs{
			ID:     testutil.TestCampaignID,
			Name:   "Updated Name",
			Status: "ACTIVE",
		}

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"id":     args.ID,
					"name":   args.Name,
					"status": args.Status,
				},
			},
		}

		result, err := UpdateCampaignHandler(context.Background(), request, args)
		if err != nil {
			t.Fatalf("Update campaign failed: %v", err)
		}

		testutil.AssertResult(t, result).IsSuccess()
	})

	// Test 3: Get insights
	t.Run("GetCampaignInsights", func(t *testing.T) {
		env.Server().AddRoute("GET", "/v23.0/"+testutil.TestCampaignID+"/insights", func(w http.ResponseWriter, r *http.Request) {
			env.Server().WriteSuccess(w, testutil.CreateMockInsightsResponse())
		})

		args := get_campaign_insightsArgs{
			ID:         testutil.TestCampaignID,
			Fields:     []string{"campaign_id", "impressions", "clicks", "spend"},
			DatePreset: "last_30d",
		}

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"id":          args.ID,
					"fields":      args.Fields,
					"date_preset": args.DatePreset,
				},
			},
		}

		result, err := GetCampaignInsightsHandler(context.Background(), request, args)
		if err != nil {
			t.Fatalf("Get campaign insights failed: %v", err)
		}

		// Validate insights data structure
		data := testutil.AssertResult(t, result).
			IsSuccess().
			ParseJSON()

		if insights, ok := data["data"].([]interface{}); ok {
			if len(insights) > 0 {
				if insight, ok := insights[0].(map[string]interface{}); ok {
					// Verify expected fields exist
					expectedFields := []string{"impressions", "clicks", "spend"}
					for _, field := range expectedFields {
						if _, exists := insight[field]; !exists {
							t.Errorf("Missing expected field: %s", field)
						}
					}
				}
			}
		}
	})
}

// Error handling tests
func TestErrorHandlingIntegration_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup error response
	env.Server().SetDefaultHandler(func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteError(w, 400, "OAuthException", "Invalid parameter")
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

	args := get_campaignArgs{
		ID:     "invalid_id",
		Fields: []string{"id"},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id":     args.ID,
				"fields": args.Fields,
			},
		},
	}

	// Execute
	result, err := GetCampaignHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Expected error in result, not handler error: %v", err)
	}

	// Assert
	testutil.AssertResult(t, result).
		IsError().
		HasErrorContaining("Invalid parameter")
}

// Benchmark tests
func BenchmarkGetCampaignHandler_WithFramework(b *testing.B) {
	env := testutil.NewTestEnvironment(&testing.T{})
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("GET", "/v23.0/"+testutil.TestCampaignID+"/", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateMockCampaignResponse(testutil.TestCampaignID))
	})

	// Override URLs
	graphAPIHost = env.Server().URL
	baseGraphURL = env.Server().URL

	args := get_campaignArgs{
		ID:     testutil.TestCampaignID,
		Fields: []string{"id", "name", "status"},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id":     args.ID,
				"fields": args.Fields,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCampaignHandler(context.Background(), request, args)
		if err != nil {
			b.Fatalf("Handler failed: %v", err)
		}
	}
}

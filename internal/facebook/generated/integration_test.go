package generated

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

// Test values based on Facebook SDK test constants
const (
	TestAccountID         = "act_123"
	TestCampaignID        = "1234321"
	TestAdsetID           = "12345"
	TestAdID              = "125475"
	TestName              = "test_name"
	TestStatus            = "PAUSED"
	TestBidStrategy       = "LOWEST_COST_WITHOUT_CAP"
	TestObjective         = "LINK_CLICKS"
	TestDailyBudget       = "200"
	TestLifetimeBudget    = "10000"
	TestBuyingType        = "AUCTION"
	TestSpecialAdCategory = "EMPLOYMENT"
)

// Test structures
type MockGraphResponse struct {
	ID      string                   `json:"id"`
	Name    string                   `json:"name,omitempty"`
	Status  string                   `json:"status,omitempty"`
	Data    []map[string]interface{} `json:"data,omitempty"`
	Success bool                     `json:"success,omitempty"`
}

// IntegrationTestSuite provides common test infrastructure
type IntegrationTestSuite struct {
	server    *httptest.Server
	serverURL string
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Create mock Facebook Graph API server
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Route requests based on path and method
		suite.handleRequest(w, r)
	}))

	suite.serverURL = suite.server.URL
	// Override the Graph API host for testing to point to our mock server
	SetGraphAPIHost(suite.serverURL)
	baseGraphURL = suite.serverURL

	// Set test access token
	SetAccessToken("test_access_token")
}

func (suite *IntegrationTestSuite) TeardownTest() {
	if suite.server != nil {
		suite.server.Close()
	}
	// Restore original settings
	SetGraphAPIHost("https://graph.facebook.com")
	baseGraphURL = "https://graph.facebook.com"
	SetAccessToken("")
}

func (suite *IntegrationTestSuite) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Debug log for path matching
	// fmt.Printf("Mock server received: %s %s\n", r.Method, r.URL.Path)

	switch {
	// Campaign endpoints
	case r.Method == "GET" && (r.URL.Path == "/v23.0/"+TestCampaignID || r.URL.Path == "/v23.0/"+TestCampaignID+"/"):
		suite.handleGetCampaign(w, r)
	case r.Method == "POST" && (r.URL.Path == "/v23.0/"+TestCampaignID || r.URL.Path == "/v23.0/"+TestCampaignID+"/"):
		suite.handleUpdateCampaign(w, r)
	case r.Method == "DELETE" && (r.URL.Path == "/v23.0/"+TestCampaignID || r.URL.Path == "/v23.0/"+TestCampaignID+"/"):
		suite.handleDeleteCampaign(w, r)
	case r.Method == "GET" && (r.URL.Path == "/v23.0/"+TestCampaignID+"/insights" || r.URL.Path == "/v23.0/"+TestCampaignID+"/insights/"):
		suite.handleGetCampaignInsights(w, r)
	case r.Method == "POST" && (r.URL.Path == "/v23.0/"+TestCampaignID+"/adlabels" || r.URL.Path == "/v23.0/"+TestCampaignID+"/adlabels/"):
		suite.handleCreateCampaignAdlabel(w, r)

	// Ad Set endpoints
	case r.Method == "GET" && (r.URL.Path == "/v23.0/"+TestAdsetID || r.URL.Path == "/v23.0/"+TestAdsetID+"/"):
		suite.handleGetAdSet(w, r)
	case r.Method == "POST" && (r.URL.Path == "/v23.0/"+TestAdsetID || r.URL.Path == "/v23.0/"+TestAdsetID+"/"):
		suite.handleUpdateAdSet(w, r)
	case r.Method == "POST" && (r.URL.Path == "/v23.0/"+TestAdsetID+"/adlabels" || r.URL.Path == "/v23.0/"+TestAdsetID+"/adlabels/"):
		suite.handleCreateAdSetAdlabel(w, r)

	// Ad endpoints
	case r.Method == "GET" && (r.URL.Path == "/v23.0/"+TestAdID || r.URL.Path == "/v23.0/"+TestAdID+"/"):
		suite.handleGetAd(w, r)
	case r.Method == "POST" && (r.URL.Path == "/v23.0/"+TestAdID || r.URL.Path == "/v23.0/"+TestAdID+"/"):
		suite.handleUpdateAd(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Endpoint not found",
				"type":    "GraphMethodException",
			},
		})
	}
}

func (suite *IntegrationTestSuite) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":                  TestCampaignID,
		"name":                TestName,
		"status":              TestStatus,
		"objective":           TestObjective,
		"buying_type":         TestBuyingType,
		"daily_budget":        TestDailyBudget,
		"lifetime_budget":     TestLifetimeBudget,
		"bid_strategy":        TestBidStrategy,
		"special_ad_category": TestSpecialAdCategory,
		"promoted_object": map[string]interface{}{
			"page_id": "13531",
		},
		"adlabels": []map[string]interface{}{
			{
				"id":   "label_123",
				"name": "Test Label",
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleUpdateCampaign(w http.ResponseWriter, r *http.Request) {
	response := MockGraphResponse{
		ID:      TestCampaignID,
		Success: true,
	}
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleDeleteCampaign(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"success": true,
	}
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleGetCampaignInsights(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"campaign_id": TestCampaignID,
				"impressions": "1000",
				"clicks":      "50",
				"spend":       "100.00",
				"date_start":  "2024-01-01",
				"date_stop":   "2024-01-31",
			},
		},
		"paging": map[string]interface{}{
			"cursors": map[string]interface{}{
				"before": "before_cursor",
				"after":  "after_cursor",
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleCreateCampaignAdlabel(w http.ResponseWriter, r *http.Request) {
	response := MockGraphResponse{
		ID:      TestCampaignID,
		Success: true,
	}
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleGetAdSet(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":                TestAdsetID,
		"name":              TestName,
		"status":            TestStatus,
		"campaign_id":       TestCampaignID,
		"daily_budget":      TestDailyBudget,
		"lifetime_budget":   TestLifetimeBudget,
		"bid_strategy":      TestBidStrategy,
		"optimization_goal": "LINK_CLICKS",
		"targeting": map[string]interface{}{
			"geo_locations": map[string]interface{}{
				"countries": []string{"US"},
			},
		},
		"promoted_object": map[string]interface{}{
			"page_id": "13531",
		},
		"adlabels": []map[string]interface{}{
			{
				"id":   "label_456",
				"name": "AdSet Label",
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleUpdateAdSet(w http.ResponseWriter, r *http.Request) {
	response := MockGraphResponse{
		ID:      TestAdsetID,
		Success: true,
	}
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleCreateAdSetAdlabel(w http.ResponseWriter, r *http.Request) {
	response := MockGraphResponse{
		ID:      TestAdsetID,
		Success: true,
	}
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleGetAd(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":          TestAdID,
		"name":        TestName,
		"status":      TestStatus,
		"campaign_id": TestCampaignID,
		"adset_id":    TestAdsetID,
		"creative": map[string]interface{}{
			"id":   "15742462",
			"name": "test creative",
		},
		"adlabels": []map[string]interface{}{
			{
				"id":   "label_789",
				"name": "Ad Label",
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleUpdateAd(w http.ResponseWriter, r *http.Request) {
	response := MockGraphResponse{
		ID:      TestAdID,
		Success: true,
	}
	json.NewEncoder(w).Encode(response)
}

// Campaign Integration Tests
func TestGetCampaignIntegration(t *testing.T) {
	// Arrange
	suite := &IntegrationTestSuite{}
	suite.SetupTest()
	defer suite.TeardownTest()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_campaign",
			Arguments: map[string]interface{}{
				"id": TestCampaignID,
				"fields": []string{
					"id", "name", "status", "objective", "daily_budget",
					"lifetime_budget", "bid_strategy", "promoted_object", "adlabels",
				},
			},
		},
	}

	args := get_campaignArgs{
		ID: TestCampaignID,
		Fields: []string{
			"id", "name", "status", "objective", "daily_budget",
			"lifetime_budget", "bid_strategy", "promoted_object", "adlabels",
		},
	}

	// Act
	result, err := GetCampaignHandler(ctx, request, args)

	// Assert
	if err != nil {
		t.Fatalf("GetCampaignHandler failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Fatalf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Fatal("Expected text content in error result")
		}
	}

	// Parse and validate response content
	var response map[string]interface{}
	if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
		if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
	} else {
		t.Fatal("Expected text content in result")
	}

	// Validate response structure
	if response["id"] != TestCampaignID {
		t.Errorf("Expected campaign ID %s, got %v", TestCampaignID, response["id"])
	}

	if response["name"] != TestName {
		t.Errorf("Expected name %s, got %v", TestName, response["name"])
	}

	if response["status"] != TestStatus {
		t.Errorf("Expected status %s, got %v", TestStatus, response["status"])
	}

	// Validate complex types
	if promotedObject, ok := response["promoted_object"].(map[string]interface{}); ok {
		if promotedObject["page_id"] != "13531" {
			t.Errorf("Expected promoted_object.page_id '13531', got %v", promotedObject["page_id"])
		}
	} else {
		t.Error("Expected promoted_object to be an object")
	}

	// Validate array types
	if adlabels, ok := response["adlabels"].([]interface{}); ok {
		if len(adlabels) != 1 {
			t.Errorf("Expected 1 adlabel, got %d", len(adlabels))
		}
		if adlabel, ok := adlabels[0].(map[string]interface{}); ok {
			if adlabel["name"] != "Test Label" {
				t.Errorf("Expected adlabel name 'Test Label', got %v", adlabel["name"])
			}
		}
	} else {
		t.Error("Expected adlabels to be an array")
	}
}

func TestUpdateCampaignWithTypedArgsIntegration(t *testing.T) {
	// Arrange
	suite := &IntegrationTestSuite{}
	suite.SetupTest()
	defer suite.TeardownTest()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "update_campaign",
			Arguments: map[string]interface{}{
				"id":           TestCampaignID,
				"name":         "Updated Campaign Name",
				"status":       "ACTIVE",
				"daily_budget": 300,
				"bid_strategy": "COST_CAP",
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

	// Test typed arguments with complex types
	args := update_campaignArgs{
		ID:          TestCampaignID,
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

	// Act
	result, err := UpdateCampaignHandler(ctx, request, args)

	// Assert
	if err != nil {
		t.Fatalf("UpdateCampaignHandler failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Fatalf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Fatal("Expected text content in error result")
		}
	}

	var response MockGraphResponse
	if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
		if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
	} else {
		t.Fatal("Expected text content in result")
	}

	if response.ID != TestCampaignID {
		t.Errorf("Expected campaign ID %s, got %s", TestCampaignID, response.ID)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestCreateCampaignAdlabelWithValidationIntegration(t *testing.T) {
	// Arrange
	suite := &IntegrationTestSuite{}
	suite.SetupTest()
	defer suite.TeardownTest()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_campaign_adlabel",
			Arguments: map[string]interface{}{
				"id": TestCampaignID,
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
				"execution_options": []string{"include_recommendations"},
			},
		},
	}

	args := create_campaign_adlabelArgs{
		ID: TestCampaignID,
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

	// Act
	result, err := CreateCampaignAdlabelHandler(ctx, request, args)

	// Assert
	if err != nil {
		t.Fatalf("CreateCampaignAdlabelHandler failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Fatalf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Fatal("Expected text content in error result")
		}
	}

	var response MockGraphResponse
	if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
		if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
	} else {
		t.Fatal("Expected text content in result")
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

// Test validation errors
func TestCampaignHandlerValidationErrors(t *testing.T) {
	// Arrange
	suite := &IntegrationTestSuite{}
	suite.SetupTest()
	defer suite.TeardownTest()

	ctx := context.Background()
	request := mcp.CallToolRequest{}

	// Test missing required ID
	args := get_campaignArgs{
		ID:     "", // Missing required ID
		Fields: []string{"name"},
	}

	// Act
	result, err := GetCampaignHandler(ctx, request, args)

	// Assert
	if err != nil {
		t.Fatalf("Expected validation error in result, not handler error: %v", err)
	}

	if !result.IsError {
		t.Error("Expected validation error for missing ID")
	}

	// Check if result has error content
	if len(result.Content) == 0 {
		t.Error("Expected error content in result")
	} else {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if !strings.Contains(textContent.Text, "id is required") {
				t.Errorf("Expected 'id is required' error, got: %s", textContent.Text)
			}
		} else {
			t.Error("Expected text content in error result")
		}
	}
}

// Ad Set Integration Tests
func TestGetAdSetWithComplexTypesIntegration(t *testing.T) {
	// Arrange
	suite := &IntegrationTestSuite{}
	suite.SetupTest()
	defer suite.TeardownTest()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_ad_set",
			Arguments: map[string]interface{}{
				"id": TestAdsetID,
				"fields": []string{
					"id", "name", "status", "targeting", "promoted_object", "adlabels",
				},
			},
		},
	}

	args := get_ad_setArgs{
		ID: TestAdsetID,
		Fields: []string{
			"id", "name", "status", "targeting", "promoted_object", "adlabels",
		},
	}

	// Act
	result, err := GetAdSetHandler(ctx, request, args)

	// Assert
	if err != nil {
		t.Fatalf("GetAdSetHandler failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Fatalf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Fatal("Expected text content in error result")
		}
	}

	var response map[string]interface{}
	if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
		if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
	} else {
		t.Fatal("Expected text content in result")
	}

	// Validate targeting complex type
	if targeting, ok := response["targeting"].(map[string]interface{}); ok {
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
func TestCampaignLifecycleIntegration(t *testing.T) {
	// Arrange
	suite := &IntegrationTestSuite{}
	suite.SetupTest()
	defer suite.TeardownTest()

	ctx := context.Background()

	// Test 1: Get campaign
	t.Run("GetCampaign", func(t *testing.T) {
		args := get_campaignArgs{
			ID:     TestCampaignID,
			Fields: []string{"id", "name", "status"},
		}

		result, err := GetCampaignHandler(ctx, mcp.CallToolRequest{}, args)

		if err != nil {
			t.Fatalf("Get campaign failed: %v", err)
		}
		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Fatalf("Get campaign failed: %s", textContent.Text)
			} else {
				t.Fatal("Expected text content in error result")
			}
		}
	})

	// Test 2: Update campaign
	t.Run("UpdateCampaign", func(t *testing.T) {
		args := update_campaignArgs{
			ID:     TestCampaignID,
			Name:   "Updated Name",
			Status: "ACTIVE",
		}

		result, err := UpdateCampaignHandler(ctx, mcp.CallToolRequest{}, args)

		if err != nil {
			t.Fatalf("Update campaign failed: %v", err)
		}
		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Fatalf("Update campaign failed: %s", textContent.Text)
			} else {
				t.Fatal("Expected text content in error result")
			}
		}
	})

	// Test 3: Get insights
	t.Run("GetCampaignInsights", func(t *testing.T) {
		args := get_campaign_insightsArgs{
			ID:         TestCampaignID,
			Fields:     []string{"campaign_id", "impressions", "clicks", "spend"},
			DatePreset: "last_30d",
		}

		result, err := GetCampaignInsightsHandler(ctx, mcp.CallToolRequest{}, args)

		if err != nil {
			t.Fatalf("Get campaign insights failed: %v", err)
		}
		if result.IsError {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				t.Fatalf("Get campaign insights failed: %s", textContent.Text)
			} else {
				t.Fatal("Expected text content in error result")
			}
		}

		// Validate insights data structure
		var response map[string]interface{}
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("Failed to parse insights response: %v", err)
			}
		} else {
			t.Fatal("Expected text content in result")
		}

		if data, ok := response["data"].([]interface{}); ok {
			if len(data) > 0 {
				if insight, ok := data[0].(map[string]interface{}); ok {
					if insight["campaign_id"] != TestCampaignID {
						t.Errorf("Expected campaign_id %s in insights, got %v", TestCampaignID, insight["campaign_id"])
					}
				}
			}
		}
	})
}

// Error handling tests
func TestErrorHandlingIntegration(t *testing.T) {
	// Arrange
	suite := &IntegrationTestSuite{}
	suite.SetupTest()
	defer suite.TeardownTest()

	// Override server to return error
	suite.server.Close()
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message":    "Invalid parameter",
				"type":       "OAuthException",
				"code":       100,
				"fbtrace_id": "trace123",
			},
		})
	}))

	ctx := context.Background()
	args := get_campaignArgs{
		ID:     "invalid_id",
		Fields: []string{"id"},
	}

	// Act
	result, err := GetCampaignHandler(ctx, mcp.CallToolRequest{}, args)

	// Assert
	if err != nil {
		t.Fatalf("Expected error in result, not handler error: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error result for invalid request")
	}

	// Should contain error information
	if len(result.Content) == 0 {
		t.Error("Expected error content in result")
	} else {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			if textContent.Text == "" {
				t.Error("Expected non-empty error message")
			}
		} else {
			t.Error("Expected text content in error result")
		}
	}
}

// Benchmark tests
func BenchmarkGetCampaignHandler(b *testing.B) {
	suite := &IntegrationTestSuite{}
	suite.SetupTest()
	defer suite.TeardownTest()

	ctx := context.Background()
	args := get_campaignArgs{
		ID:     TestCampaignID,
		Fields: []string{"id", "name", "status"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCampaignHandler(ctx, mcp.CallToolRequest{}, args)
		if err != nil {
			b.Fatalf("Handler failed: %v", err)
		}
	}
}

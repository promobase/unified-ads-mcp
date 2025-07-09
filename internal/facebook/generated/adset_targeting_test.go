package generated

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"unified-ads-mcp/internal/facebook/testutil"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")
}

// TestCreateAdSetWithTargeting tests the create_ad_account_adset handler with proper targeting
func TestCreateAdSetWithTargeting(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("POST", "/v23.0/act_123456789/adsets", func(w http.ResponseWriter, r *http.Request) {
		// Check access token in URL
		if r.URL.Query().Get("access_token") == "" {
			env.Server().WriteAuthError(w)
			return
		}

		// Parse JSON body
		var requestData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			env.Server().WriteError(w, 400, "ParseError", "Failed to parse JSON body")
			return
		}

		// Validate required fields
		if requestData["name"] == nil || requestData["name"] == "" {
			env.Server().WriteError(w, 400, "ValidationError", "name is required")
			return
		}

		if requestData["campaign_id"] == nil || requestData["campaign_id"] == "" {
			env.Server().WriteError(w, 400, "ValidationError", "campaign_id is required")
			return
		}

		if requestData["targeting"] == nil {
			env.Server().WriteError(w, 400, "ValidationError", "targeting is required")
			return
		}

		// Validate targeting is an object
		if _, ok := requestData["targeting"].(map[string]interface{}); !ok {
			env.Server().WriteError(w, 400, "ValidationError", "targeting must be an object")
			return
		}

		// Return success response
		env.Server().WriteSuccess(w, map[string]interface{}{
			"id": "120229334567890123",
		})
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

	tests := []struct {
		name      string
		args      interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid targeting as struct",
			args: ad_account_create_adsetArgs{
				ID:               "act_123456789",
				Name:             "Test Ad Set",
				CampaignId:       "120229333949190588",
				Status:           "PAUSED",
				DailyBudget:      1000,
				BillingEvent:     "LINK_CLICKS",
				OptimizationGoal: "LINK_CLICKS",
				Targeting: &Targeting{
					AgeMin: 18,
					AgeMax: 65,
					GeoLocations: &TargetingGeoLocation{
						Countries:     []string{"US"},
						LocationTypes: []string{"home", "recent"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "Valid targeting with cities",
			args: ad_account_create_adsetArgs{
				ID:               "act_123456789",
				Name:             "Test Ad Set with Cities",
				CampaignId:       "120229333949190588",
				Status:           "PAUSED",
				DailyBudget:      2000,
				BillingEvent:     "IMPRESSIONS",
				OptimizationGoal: "REACH",
				Targeting: &Targeting{
					AgeMin: 25,
					AgeMax: 54,
					GeoLocations: &TargetingGeoLocation{
						Cities: []*TargetingGeoLocationCity{
							{
								Key:  "777934",
								Name: "San Francisco",
							},
							{
								Key:  "2420379",
								Name: "New York",
							},
						},
						LocationTypes: []string{"home"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "Missing required targeting",
			args: ad_account_create_adsetArgs{
				ID:               "act_123456789",
				Name:             "Test Ad Set",
				CampaignId:       "120229333949190588",
				Status:           "PAUSED",
				DailyBudget:      1000,
				BillingEvent:     "LINK_CLICKS",
				OptimizationGoal: "LINK_CLICKS",
				// No targeting specified
			},
			wantError: true,
			errorMsg:  "targeting is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args.(ad_account_create_adsetArgs)

			// Create request
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "create_ad_account_adset",
					Arguments: map[string]interface{}{
						"id":                args.ID,
						"name":              args.Name,
						"campaign_id":       args.CampaignId,
						"status":            args.Status,
						"daily_budget":      args.DailyBudget,
						"billing_event":     args.BillingEvent,
						"optimization_goal": args.OptimizationGoal,
						"targeting":         args.Targeting,
					},
				},
			}

			// Execute handler
			result, err := AdAccountCreateAdsetHandler(context.Background(), request, args)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			// Check result
			if tt.wantError {
				testutil.AssertResult(t, result).
					IsError().
					HasErrorContaining(tt.errorMsg)
			} else {
				data := testutil.AssertResult(t, result).
					IsSuccess().
					ParseJSON()

				if data["id"] == nil || data["id"] == "" {
					t.Error("Expected ad set ID in response")
				}
			}
		})
	}
}

// TestTargetingJSONHandling tests how targeting JSON is handled
func TestTargetingJSONHandling(t *testing.T) {
	tests := []struct {
		name      string
		targeting *Targeting
		expected  map[string]interface{}
	}{
		{
			name: "Basic geo targeting",
			targeting: &Targeting{
				AgeMin: 18,
				AgeMax: 65,
				GeoLocations: &TargetingGeoLocation{
					Countries:     []string{"US", "CA"},
					LocationTypes: []string{"home", "recent"},
				},
			},
			expected: map[string]interface{}{
				"age_min": uint(18),
				"age_max": uint(65),
				"geo_locations": map[string]interface{}{
					"countries":      []string{"US", "CA"},
					"location_types": []string{"home", "recent"},
				},
			},
		},
		{
			name: "Targeting with interests",
			targeting: &Targeting{
				AgeMin:  21,
				AgeMax:  35,
				Genders: []uint{1, 2}, // 1=male, 2=female
				Interests: []*IDName{
					{
						ID:   "6003139266461",
						Name: "Movies",
					},
					{
						ID:   "6003248604572",
						Name: "Music",
					},
				},
				GeoLocations: &TargetingGeoLocation{
					Countries: []string{"US"},
				},
			},
			expected: map[string]interface{}{
				"age_min": uint(21),
				"age_max": uint(35),
				"genders": []uint{1, 2},
				"interests": []map[string]interface{}{
					{"id": "6003139266461", "name": "Movies"},
					{"id": "6003248604572", "name": "Music"},
				},
				"geo_locations": map[string]interface{}{
					"countries": []string{"US"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal targeting to JSON
			jsonBytes, err := json.Marshal(tt.targeting)
			if err != nil {
				t.Fatalf("Failed to marshal targeting: %v", err)
			}

			// Unmarshal back to verify
			var result map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &result); err != nil {
				t.Fatalf("Failed to unmarshal targeting JSON: %v", err)
			}

			// Compare key fields
			if tt.expected["age_min"] != nil {
				ageMin, ok := result["age_min"].(float64)
				if !ok || uint(ageMin) != tt.expected["age_min"].(uint) {
					t.Errorf("age_min mismatch: got %v, want %v", result["age_min"], tt.expected["age_min"])
				}
			}

			if tt.expected["age_max"] != nil {
				ageMax, ok := result["age_max"].(float64)
				if !ok || uint(ageMax) != tt.expected["age_max"].(uint) {
					t.Errorf("age_max mismatch: got %v, want %v", result["age_max"], tt.expected["age_max"])
				}
			}

			// Verify geo_locations exists
			if geoLoc, ok := result["geo_locations"].(map[string]interface{}); ok {
				if countries, ok := geoLoc["countries"].([]interface{}); ok {
					if len(countries) == 0 {
						t.Error("Expected countries in geo_locations")
					}
				}
			} else {
				t.Error("Expected geo_locations in result")
			}
		})
	}
}

// TestCreateAdSetIntegrationExample shows how to properly create an ad set
func TestCreateAdSetIntegrationExample(t *testing.T) {
	t.Skip("This is an example test showing proper usage")

	// Example of how to properly create an ad set with targeting
	args := ad_account_create_adsetArgs{
		ID:               "act_648073588254125",
		Name:             "Test Ad Set",
		CampaignId:       "120229333949190588",
		Status:           "PAUSED",
		DailyBudget:      1000,
		BillingEvent:     "LINK_CLICKS",
		OptimizationGoal: "LINK_CLICKS",
		Targeting: &Targeting{
			AgeMin: 18,
			AgeMax: 65,
			GeoLocations: &TargetingGeoLocation{
				Countries:     []string{"US"},
				LocationTypes: []string{"home", "recent"},
			},
		},
	}

	// The MCP framework expects the targeting as a struct, not a JSON string
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_ad_account_adset",
			Arguments: map[string]interface{}{
				"id":                args.ID,
				"name":              args.Name,
				"campaign_id":       args.CampaignId,
				"status":            args.Status,
				"daily_budget":      args.DailyBudget,
				"billing_event":     args.BillingEvent,
				"optimization_goal": args.OptimizationGoal,
				"targeting": map[string]interface{}{
					"age_min": 18,
					"age_max": 65,
					"geo_locations": map[string]interface{}{
						"countries":      []string{"US"},
						"location_types": []string{"home", "recent"},
					},
				},
			},
		},
	}

	// This is how the handler would be called
	_ = request
}

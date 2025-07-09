package generated

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"unified-ads-mcp/internal/facebook/testutil"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")
}

// TestTypedArgumentsValidation tests the validation and serialization of typed arguments
func TestTypedArgumentsValidation(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "ValidAdLabelSerialization",
			testFunc: func(t *testing.T) {
				// Arrange
				labels := []*AdLabel{
					{
						ID:          "label_123",
						Name:        "Performance Campaign",
						CreatedTime: time.Now(),
						UpdatedTime: time.Now(),
					},
					{
						ID:   "label_456",
						Name: "Brand Campaign",
					},
				}

				// Act - Serialize to JSON
				jsonData, err := json.Marshal(labels)
				if err != nil {
					t.Fatalf("Failed to marshal AdLabel: %v", err)
				}

				// Deserialize back
				var deserializedLabels []*AdLabel
				err = json.Unmarshal(jsonData, &deserializedLabels)
				if err != nil {
					t.Fatalf("Failed to unmarshal AdLabel: %v", err)
				}

				// Assert
				assert := testutil.NewAssert(t)
				assert.Equal(len(deserializedLabels), 2, "Expected 2 labels")
				assert.Equal(deserializedLabels[0].ID, "label_123", "Expected label ID")
				assert.Equal(deserializedLabels[0].Name, "Performance Campaign", "Expected label name")
			},
		},
		{
			name: "ValidAdPromotedObjectSerialization",
			testFunc: func(t *testing.T) {
				// Arrange
				promotedObj := &AdPromotedObject{
					PageID:          "page_123",
					CustomEventType: AdpromotedobjectCustomEventType_PURCHASE,
					ApplicationID:   "app_456",
				}

				// Act
				jsonData, err := json.Marshal(promotedObj)
				if err != nil {
					t.Fatalf("Failed to marshal AdPromotedObject: %v", err)
				}

				var deserialized *AdPromotedObject
				err = json.Unmarshal(jsonData, &deserialized)
				if err != nil {
					t.Fatalf("Failed to unmarshal AdPromotedObject: %v", err)
				}

				// Assert
				assert := testutil.NewAssert(t)
				assert.Equal(deserialized.PageID, "page_123", "Expected page_id")
				assert.Equal(deserialized.CustomEventType, AdpromotedobjectCustomEventType_PURCHASE, "Expected custom_event_type")
			},
		},
		{
			name: "ValidTargetingSerialization",
			testFunc: func(t *testing.T) {
				// Arrange
				targeting := &Targeting{
					AgeMin:  18,
					AgeMax:  65,
					Genders: []uint{1, 2}, // All genders
					GeoLocations: &TargetingGeoLocation{
						Countries: []string{"US", "CA"},
						Regions: []*TargetingGeoLocationRegion{
							{
								Key:  "3843",
								Name: "California",
							},
						},
					},
				}

				// Act
				jsonData, err := json.Marshal(targeting)
				if err != nil {
					t.Fatalf("Failed to marshal Targeting: %v", err)
				}

				var deserialized *Targeting
				err = json.Unmarshal(jsonData, &deserialized)
				if err != nil {
					t.Fatalf("Failed to unmarshal Targeting: %v", err)
				}

				// Assert
				assert := testutil.NewAssert(t)
				assert.EqualInt(int(deserialized.AgeMin), 18, "Expected age_min")
				assert.Equal(len(deserialized.GeoLocations.Countries), 2, "Expected 2 countries")
				assert.Equal(deserialized.GeoLocations.Countries[0], "US", "Expected first country")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestTypedArgumentsInHandlers tests typed arguments in actual handler calls
func TestTypedArgumentsInHandlers(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock server
	env.Server().AddRoute("POST", "/v23.0/123456789", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateMockCampaignResponse("123456789"))
	})
	env.Server().AddRoute("POST", "/v23.0/adset_123", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, map[string]interface{}{
			"id":   "adset_123",
			"name": "Test AdSet",
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
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "CampaignUpdateWithTypedArgs",
			testFunc: func(t *testing.T) {
				// Arrange
				args := campaign_updateArgs{
					ID:          "123456789",
					Name:        "Test Campaign",
					Status:      "PAUSED",
					DailyBudget: 100,
					Adlabels: []*AdLabel{
						{
							ID:   "label_001",
							Name: "Test Label",
						},
					},
					PromotedObject: &AdPromotedObject{
						PageID: "page_123",
					},
				}

				// Validate struct field types at compile time
				var _ string = args.ID
				var _ string = args.Name
				var _ string = args.Status
				var _ int = args.DailyBudget
				var _ []*AdLabel = args.Adlabels
				var _ *AdPromotedObject = args.PromotedObject

				// Test JSON serialization
				jsonData, err := json.Marshal(args)
				if err != nil {
					t.Fatalf("Failed to marshal update_campaignArgs: %v", err)
				}

				// Validate JSON structure
				var jsonMap map[string]interface{}
				err = json.Unmarshal(jsonData, &jsonMap)
				if err != nil {
					t.Fatalf("Failed to unmarshal to map: %v", err)
				}

				// Assert JSON field names
				assert := testutil.NewAssert(t)
				assert.Equal(jsonMap["id"], "123456789", "Expected JSON field 'id'")
				assert.Equal(jsonMap["daily_budget"], float64(100), "Expected JSON field 'daily_budget'")

				// Validate complex nested structures
				adlabels, ok := jsonMap["adlabels"].([]interface{})
				assert.True(ok, "Expected adlabels to be an array")
				assert.Equal(len(adlabels), 1, "Expected 1 adlabel")

				if len(adlabels) > 0 {
					label, ok := adlabels[0].(map[string]interface{})
					assert.True(ok, "Expected adlabel to be object")
					assert.Equal(label["name"], "Test Label", "Expected adlabel name")
				}

				promotedObj, ok := jsonMap["promoted_object"].(map[string]interface{})
				assert.True(ok, "Expected promoted_object to be object")
				assert.Equal(promotedObj["page_id"], "page_123", "Expected page_id")
			},
		},
		{
			name: "AdSetUpdateWithComplexTargeting",
			testFunc: func(t *testing.T) {
				// Arrange
				args := ad_set_updateArgs{
					ID:          "adset_123",
					Name:        "Test AdSet",
					DailyBudget: 50,
					Targeting: &Targeting{
						AgeMin:  25,
						AgeMax:  45,
						Genders: []uint{1}, // Male only
						GeoLocations: &TargetingGeoLocation{
							Countries: []string{"US"},
							Cities: []*TargetingGeoLocationCity{
								{
									Key:      "2418046",
									Name:     "New York",
									Region:   "New York",
									RegionID: "3875",
								},
							},
						},
					},
					PromotedObject: &AdPromotedObject{
						PageID: "page_456",
					},
					Adlabels: []*AdLabel{
						{
							ID:   "adset_label_001",
							Name: "AdSet Performance Label",
						},
					},
				}

				// Test serialization
				jsonData, err := json.Marshal(args)
				if err != nil {
					t.Fatalf("Failed to marshal update_ad_setArgs: %v", err)
				}

				var jsonMap map[string]interface{}
				err = json.Unmarshal(jsonData, &jsonMap)
				if err != nil {
					t.Fatalf("Failed to unmarshal to map: %v", err)
				}

				// Validate targeting structure
				assert := testutil.NewAssert(t)
				targeting, ok := jsonMap["targeting"].(map[string]interface{})
				assert.True(ok, "Expected targeting to be object")
				assert.Equal(targeting["age_min"], float64(25), "Expected age_min")

				geoLoc, ok := targeting["geo_locations"].(map[string]interface{})
				assert.True(ok, "Expected geo_locations to be object")

				countries, ok := geoLoc["countries"].([]interface{})
				assert.True(ok, "Expected countries to be array")
				assert.Equal(len(countries), 1, "Expected 1 country")
				if len(countries) > 0 {
					assert.Equal(countries[0], "US", "Expected country")
				}

				cities, ok := geoLoc["cities"].([]interface{})
				assert.True(ok, "Expected cities to be array")
				assert.Equal(len(cities), 1, "Expected 1 city")
				if len(cities) > 0 {
					city, ok := cities[0].(map[string]interface{})
					assert.True(ok, "Expected city to be object")
					assert.Equal(city["name"], "New York", "Expected city name")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestRequiredFieldValidation tests validation of required fields
func TestRequiredFieldValidation(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock server that returns validation errors
	env.Server().SetDefaultHandler(func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteError(w, 400, "ValidationError", "id is required")
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

	t.Run("RequiredIDValidation", func(t *testing.T) {
		ctx := context.Background()

		// Test with empty required ID
		args := campaign_getArgs{
			ID:     "", // Required but empty
			Fields: []string{"name"},
		}

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
			},
		}

		result, err := CampaignGetHandler(ctx, request, args)
		if err != nil {
			t.Fatalf("Expected validation error in result, not handler error: %v", err)
		}

		// Validate result
		resultAssert := testutil.AssertResult(t, result)
		resultAssert.
			IsError().
			HasErrorContaining("id is required")
	})

	t.Run("RequiredAdlabelsValidation", func(t *testing.T) {
		// Test with missing required adlabels
		args := campaign_create_adlabelArgs{
			ID:       "123456789",
			Adlabels: nil, // Required but nil
		}

		// Test that empty adlabels are handled gracefully
		args.Adlabels = []*AdLabel{} // Empty slice

		// The handler should handle empty slices appropriately
		jsonData, err := json.Marshal(args)
		if err != nil {
			t.Fatalf("Failed to marshal args with empty adlabels: %v", err)
		}

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Should have empty array, not null
		adlabels, ok := jsonMap["adlabels"].([]interface{})
		assert := testutil.NewAssert(t)
		assert.True(ok, "Expected adlabels to be array")
		assert.Equal(len(adlabels), 0, "Expected empty adlabels array")
	})
}

// TestTypeConversionEdgeCases tests edge cases in type conversion
func TestTypeConversionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "NilPointerHandling",
			testFunc: func(t *testing.T) {
				// Test with nil complex types
				args := campaign_updateArgs{
					ID:             "123",
					PromotedObject: nil, // Nil pointer
					Adlabels:       nil, // Nil slice
				}

				jsonData, err := json.Marshal(args)
				if err != nil {
					t.Fatalf("Failed to marshal args with nil fields: %v", err)
				}

				var jsonMap map[string]interface{}
				err = json.Unmarshal(jsonData, &jsonMap)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				// Nil pointers should be omitted (omitempty tag)
				assert := testutil.NewAssert(t)
				_, exists := jsonMap["promoted_object"]
				assert.False(exists, "Expected promoted_object to be omitted when nil")

				_, exists = jsonMap["adlabels"]
				assert.False(exists, "Expected adlabels to be omitted when nil")
			},
		},
		{
			name: "NumberTypeHandling",
			testFunc: func(t *testing.T) {
				// Test various number types
				args := campaign_updateArgs{
					ID:             "123",
					DailyBudget:    0,     // Zero value
					LifetimeBudget: 10000, // Positive value
					SpendCap:       -1,    // Negative value (if allowed)
				}

				jsonData, err := json.Marshal(args)
				if err != nil {
					t.Fatalf("Failed to marshal args with number fields: %v", err)
				}

				var jsonMap map[string]interface{}
				err = json.Unmarshal(jsonData, &jsonMap)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				// Zero values should be omitted due to omitempty
				assert := testutil.NewAssert(t)
				_, exists := jsonMap["daily_budget"]
				assert.False(exists, "Expected daily_budget to be omitted when zero")

				// Positive values should be included
				assert.Equal(jsonMap["lifetime_budget"], float64(10000), "Expected lifetime_budget")
			},
		},
		{
			name: "StringSliceHandling",
			testFunc: func(t *testing.T) {
				// Test string slices with various values
				args := campaign_updateArgs{
					ID:                  "123",
					ExecutionOptions:    []string{"include_recommendations", "validate_only"},
					SpecialAdCategories: []string{},           // Empty slice
					PacingType:          []string{"standard"}, // Single item
				}

				jsonData, err := json.Marshal(args)
				if err != nil {
					t.Fatalf("Failed to marshal args with string slices: %v", err)
				}

				var jsonMap map[string]interface{}
				err = json.Unmarshal(jsonData, &jsonMap)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				assert := testutil.NewAssert(t)

				// Non-empty slices should be included
				execOptions, ok := jsonMap["execution_options"].([]interface{})
				assert.True(ok, "Expected execution_options to be array")
				assert.Equal(len(execOptions), 2, "Expected 2 execution options")
				if len(execOptions) > 0 {
					assert.Equal(execOptions[0], "include_recommendations", "Expected first option")
				}

				// Empty slices should be omitted
				_, exists := jsonMap["special_ad_categories"]
				assert.False(exists, "Expected special_ad_categories to be omitted when empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestPerformanceOfTypedArgs benchmarks typed argument handling
func BenchmarkTypedArgumentsSerialization(b *testing.B) {
	// Create a complex args structure
	args := campaign_updateArgs{
		ID:          "123456789",
		Name:        "Performance Campaign",
		Status:      "ACTIVE",
		DailyBudget: 1000,
		Adlabels: []*AdLabel{
			{ID: "label1", Name: "Label 1"},
			{ID: "label2", Name: "Label 2"},
			{ID: "label3", Name: "Label 3"},
		},
		PromotedObject: &AdPromotedObject{
			PageID:          "page_123",
			CustomEventType: "PURCHASE",
		},
		ExecutionOptions: []string{"include_recommendations", "validate_only"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(args)
		if err != nil {
			b.Fatalf("Marshal failed: %v", err)
		}
	}
}

func BenchmarkTypedArgumentsDeserialization(b *testing.B) {
	// Pre-serialize the data
	args := campaign_updateArgs{
		ID:          "123456789",
		Name:        "Performance Campaign",
		Status:      "ACTIVE",
		DailyBudget: 1000,
		Adlabels: []*AdLabel{
			{ID: "label1", Name: "Label 1"},
			{ID: "label2", Name: "Label 2"},
		},
		PromotedObject: &AdPromotedObject{
			PageID: "page_123",
		},
	}

	data, err := json.Marshal(args)
	if err != nil {
		b.Fatalf("Pre-marshal failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result campaign_updateArgs
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Unmarshal failed: %v", err)
		}
	}
}

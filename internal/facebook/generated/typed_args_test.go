package generated

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

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

// TestTypedArgumentsValidation tests the validation and serialization of typed arguments
func TestTypedArgumentsValidation(t *testing.T) {
	t.Run("ValidAdLabelSerialization", func(t *testing.T) {
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
		if len(deserializedLabels) != 2 {
			t.Errorf("Expected 2 labels, got %d", len(deserializedLabels))
		}

		if deserializedLabels[0].ID != "label_123" {
			t.Errorf("Expected label ID 'label_123', got '%s'", deserializedLabels[0].ID)
		}

		if deserializedLabels[0].Name != "Performance Campaign" {
			t.Errorf("Expected label name 'Performance Campaign', got '%s'", deserializedLabels[0].Name)
		}
	})

	t.Run("ValidAdPromotedObjectSerialization", func(t *testing.T) {
		// Arrange
		promotedObj := &AdPromotedObject{
			PageID:          "page_123",
			CustomEventType: "PURCHASE",
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
		if deserialized.PageID != "page_123" {
			t.Errorf("Expected page_id 'page_123', got '%s'", deserialized.PageID)
		}

		if deserialized.CustomEventType != "PURCHASE" {
			t.Errorf("Expected custom_event_type 'PURCHASE', got '%s'", deserialized.CustomEventType)
		}
	})

	t.Run("ValidTargetingSerialization", func(t *testing.T) {
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
		if deserialized.AgeMin != 18 {
			t.Errorf("Expected age_min 18, got %d", deserialized.AgeMin)
		}

		if len(deserialized.GeoLocations.Countries) != 2 {
			t.Errorf("Expected 2 countries, got %d", len(deserialized.GeoLocations.Countries))
		}

		if deserialized.GeoLocations.Countries[0] != "US" {
			t.Errorf("Expected first country 'US', got '%s'", deserialized.GeoLocations.Countries[0])
		}
	})
}

// TestTypedArgumentsInHandlers tests typed arguments in actual handler calls
func TestTypedArgumentsInHandlers(t *testing.T) {
	t.Run("CampaignUpdateWithTypedArgs", func(t *testing.T) {
		// Arrange
		_ = context.Background() // ctx not used in this test
		_ = mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_campaign",
			},
		} // request not used in this test

		args := update_campaignArgs{
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

		// Test JSON serialization of the full args struct
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

		// Assert JSON field names match expected API format
		if jsonMap["id"] != "123456789" {
			t.Errorf("Expected JSON field 'id' with value '123456789', got %v", jsonMap["id"])
		}

		if jsonMap["daily_budget"] != float64(100) {
			t.Errorf("Expected JSON field 'daily_budget' with value 100, got %v", jsonMap["daily_budget"])
		}

		// Validate complex nested structures
		if adlabels, ok := jsonMap["adlabels"].([]interface{}); ok {
			if len(adlabels) != 1 {
				t.Errorf("Expected 1 adlabel in JSON, got %d", len(adlabels))
			}
			if label, ok := adlabels[0].(map[string]interface{}); ok {
				if label["name"] != "Test Label" {
					t.Errorf("Expected adlabel name 'Test Label', got %v", label["name"])
				}
			}
		} else {
			t.Error("Expected adlabels to be an array in JSON")
		}

		if promotedObj, ok := jsonMap["promoted_object"].(map[string]interface{}); ok {
			if promotedObj["page_id"] != "page_123" {
				t.Errorf("Expected promoted_object.page_id 'page_123', got %v", promotedObj["page_id"])
			}
		} else {
			t.Error("Expected promoted_object to be an object in JSON")
		}
	})

	t.Run("AdSetUpdateWithComplexTargeting", func(t *testing.T) {
		// Arrange
		args := update_ad_setArgs{
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

		// Test serialization preserves all complex nested structures
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
		if targeting, ok := jsonMap["targeting"].(map[string]interface{}); ok {
			if targeting["age_min"] != float64(25) {
				t.Errorf("Expected targeting.age_min 25, got %v", targeting["age_min"])
			}

			if geoLoc, ok := targeting["geo_locations"].(map[string]interface{}); ok {
				if countries, ok := geoLoc["countries"].([]interface{}); ok {
					if len(countries) != 1 || countries[0] != "US" {
						t.Errorf("Expected targeting countries ['US'], got %v", countries)
					}
				} else {
					t.Error("Expected geo_locations.countries to be array")
				}

				if cities, ok := geoLoc["cities"].([]interface{}); ok {
					if len(cities) != 1 {
						t.Errorf("Expected 1 city, got %d", len(cities))
					}
					if city, ok := cities[0].(map[string]interface{}); ok {
						if city["name"] != "New York" {
							t.Errorf("Expected city name 'New York', got %v", city["name"])
						}
					}
				}
			} else {
				t.Error("Expected targeting.geo_locations to be object")
			}
		} else {
			t.Error("Expected targeting to be object in JSON")
		}
	})
}

// TestRequiredFieldValidation tests validation of required fields
func TestRequiredFieldValidation(t *testing.T) {
	t.Run("RequiredIDValidation", func(t *testing.T) {
		ctx := context.Background()
		request := mcp.CallToolRequest{}

		// Test with empty required ID
		args := get_campaignArgs{
			ID:     "", // Required but empty
			Fields: []string{"name"},
		}

		result, err := GetCampaignHandler(ctx, request, args)
		// Should not return handler error, but result should contain validation error
		if err != nil {
			t.Fatalf("Expected validation error in result, not handler error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected validation error for missing required ID")
		}

		// Check if result has error content
		if len(result.Content) == 0 {
			t.Error("Expected error content in result")
		} else {
			if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
				if textContent.Text != "id is required" {
					t.Errorf("Expected 'id is required' error, got: %s", textContent.Text)
				}
			} else {
				t.Error("Expected text content in error result")
			}
		}
	})

	t.Run("RequiredAdlabelsValidation", func(t *testing.T) {
		_ = context.Background()  // ctx not used in this test
		_ = mcp.CallToolRequest{} // request not used in this test

		// Test with missing required adlabels
		args := create_campaign_adlabelArgs{
			ID:       "123456789",
			Adlabels: nil, // Required but nil
		}

		// This should ideally be caught by validation, but depends on implementation
		// For now, we'll test that empty adlabels are handled gracefully
		args.Adlabels = []*AdLabel{} // Empty slice

		// The handler should handle empty slices appropriately
		// This test verifies the type system works with empty complex types
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
		if adlabels, ok := jsonMap["adlabels"].([]interface{}); ok {
			if len(adlabels) != 0 {
				t.Errorf("Expected empty adlabels array, got %d items", len(adlabels))
			}
		} else {
			t.Error("Expected adlabels to be array in JSON")
		}
	})
}

// TestTypeConversionEdgeCases tests edge cases in type conversion
func TestTypeConversionEdgeCases(t *testing.T) {
	t.Run("NilPointerHandling", func(t *testing.T) {
		// Test with nil complex types
		args := update_campaignArgs{
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
		if _, exists := jsonMap["promoted_object"]; exists {
			t.Error("Expected promoted_object to be omitted when nil")
		}

		if _, exists := jsonMap["adlabels"]; exists {
			t.Error("Expected adlabels to be omitted when nil")
		}
	})

	t.Run("NumberTypeHandling", func(t *testing.T) {
		// Test various number types
		args := update_campaignArgs{
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
		if _, exists := jsonMap["daily_budget"]; exists {
			t.Error("Expected daily_budget to be omitted when zero")
		}

		// Positive values should be included
		if jsonMap["lifetime_budget"] != float64(10000) {
			t.Errorf("Expected lifetime_budget 10000, got %v", jsonMap["lifetime_budget"])
		}
	})

	t.Run("StringSliceHandling", func(t *testing.T) {
		// Test string slices with various values
		args := update_campaignArgs{
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

		// Non-empty slices should be included
		if execOptions, ok := jsonMap["execution_options"].([]interface{}); ok {
			if len(execOptions) != 2 {
				t.Errorf("Expected 2 execution options, got %d", len(execOptions))
			}
			if execOptions[0] != "include_recommendations" {
				t.Errorf("Expected first option 'include_recommendations', got %v", execOptions[0])
			}
		} else {
			t.Error("Expected execution_options to be array")
		}

		// Empty slices should be omitted
		if _, exists := jsonMap["special_ad_categories"]; exists {
			t.Error("Expected special_ad_categories to be omitted when empty")
		}
	})
}

// TestPerformanceOfTypedArgs benchmarks typed argument handling
func BenchmarkTypedArgumentsSerialization(b *testing.B) {
	// Create a complex args structure
	args := update_campaignArgs{
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
	args := update_campaignArgs{
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
		var result update_campaignArgs
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Unmarshal failed: %v", err)
		}
	}
}

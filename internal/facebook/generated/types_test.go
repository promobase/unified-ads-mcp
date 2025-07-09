package generated

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"unified-ads-mcp/internal/facebook/testutil"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")
}

// TestAdMarshaling tests marshaling and unmarshaling of Ad objects
func TestAdMarshaling(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "Complete_Ad_Object",
			testFunc: func(t *testing.T) {
				// Arrange
				ad := &Ad{
					ID:              "123456789",
					Name:            "Test Ad",
					AccountID:       "act_987654321",
					Status:          "ACTIVE",
					CreatedTime:     time.Now(),
					BidAmount:       100,
					EffectiveStatus: "ACTIVE",
					Adlabels: []*AdLabel{
						{
							ID:   "label1",
							Name: "Summer Campaign",
						},
					},
				}

				// Act - Marshal to JSON
				data, err := json.Marshal(ad)
				testutil.NewAssert(t).NoError(err, "Failed to marshal Ad")

				// Unmarshal back
				var ad2 Ad
				err = json.Unmarshal(data, &ad2)
				testutil.NewAssert(t).NoError(err, "Failed to unmarshal Ad")

				// Assert
				assert := testutil.NewAssert(t)
				assert.Equal(ad2.ID, ad.ID, "Ad ID")
				assert.Equal(ad2.Name, ad.Name, "Ad Name")
				assert.Equal(ad2.AccountID, ad.AccountID, "Account ID")
				assert.Equal(ad2.Status, ad.Status, "Status")
				assert.Equal(ad2.BidAmount, ad.BidAmount, "Bid Amount")
				assert.Equal(len(ad2.Adlabels), 1, "Adlabels count")
				if len(ad2.Adlabels) > 0 {
					assert.Equal(ad2.Adlabels[0].ID, "label1", "Adlabel ID")
					assert.Equal(ad2.Adlabels[0].Name, "Summer Campaign", "Adlabel Name")
				}
			},
		},
		{
			name: "Minimal_Ad_Object",
			testFunc: func(t *testing.T) {
				// Test with minimal required fields
				ad := &Ad{
					ID: "min123",
				}

				// Should marshal without errors
				data, err := json.Marshal(ad)
				testutil.NewAssert(t).NoError(err, "Failed to marshal minimal Ad")

				// Unmarshal and verify
				var ad2 Ad
				err = json.Unmarshal(data, &ad2)
				testutil.NewAssert(t).NoError(err, "Failed to unmarshal minimal Ad")
				testutil.NewAssert(t).Equal(ad2.ID, "min123", "Minimal Ad ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestEnumTypes tests enum type handling
func TestEnumTypes(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "AdStatus_Enum",
			testFunc: func(t *testing.T) {
				var status AdStatus = "ACTIVE"

				// Should be able to use as string
				assert := testutil.NewAssert(t)
				assert.Equal(string(status), "ACTIVE", "Enum to string conversion")

				// Test JSON marshaling
				data, err := json.Marshal(status)
				assert.NoError(err, "Failed to marshal enum")
				assert.Equal(string(data), `"ACTIVE"`, "JSON representation")
			},
		},
		{
			name: "CampaignBidStrategy_Enum",
			testFunc: func(t *testing.T) {
				var strategy CampaignBidStrategy = "LOWEST_COST"

				data, err := json.Marshal(strategy)
				assert := testutil.NewAssert(t)
				assert.NoError(err, "Failed to marshal CampaignBidStrategy")

				var unmarshaled CampaignBidStrategy
				err = json.Unmarshal(data, &unmarshaled)
				assert.NoError(err, "Failed to unmarshal CampaignBidStrategy")
				assert.Equal(unmarshaled, strategy, "CampaignBidStrategy round trip")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestEmptyStruct tests handling of empty structs
func TestEmptyStruct(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Test empty struct like ProductItemInvalidationError
	item := &ProductItem{
		ID: "prod123",
		InvalidationErrors: []*ProductItemInvalidationError{
			{}, // Empty struct should work
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(item)
	assert := testutil.NewAssert(t)
	assert.NoError(err, "Failed to marshal ProductItem")
	assert.True(len(data) > 0, "JSON should not be empty")

	// Unmarshal back and verify structure
	var item2 ProductItem
	err = json.Unmarshal(data, &item2)
	assert.NoError(err, "Failed to unmarshal ProductItem")
	assert.Equal(item2.ID, "prod123", "ProductItem ID")
	assert.Equal(len(item2.InvalidationErrors), 1, "InvalidationErrors count")
}

// TestComplexMapTypes tests complex map type handling
func TestComplexMapTypes(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "CanvasTemplate_Complex_Maps",
			testFunc: func(t *testing.T) {
				template := &CanvasTemplate{
					ID:   "test123",
					Name: "Test Template",
					// Complex map types should be handled
					Objectives: []map[string]map[string]interface{}{
						{
							"objective1": {
								"type":  "CONVERSIONS",
								"value": 100,
							},
						},
					},
				}

				// Should compile and marshal correctly
				data, err := json.Marshal(template)
				assert := testutil.NewAssert(t)
				assert.NoError(err, "Failed to marshal complex type")

				// Verify JSON structure
				var jsonMap map[string]interface{}
				err = json.Unmarshal(data, &jsonMap)
				assert.NoError(err, "Failed to unmarshal to map")

				objectives, ok := jsonMap["objectives"].([]interface{})
				assert.True(ok, "objectives should be array")
				assert.Equal(len(objectives), 1, "Should have one objective")
			},
		},
		{
			name: "Nested_Map_Types",
			testFunc: func(t *testing.T) {
				// Test deeply nested map structures
				type TestStruct struct {
					Data map[string]map[string][]interface{} `json:"data"`
				}

				test := &TestStruct{
					Data: map[string]map[string][]interface{}{
						"level1": {
							"level2": []interface{}{"item1", "item2", 42},
						},
					},
				}

				data, err := json.Marshal(test)
				assert := testutil.NewAssert(t)
				assert.NoError(err, "Failed to marshal nested maps")

				var test2 TestStruct
				err = json.Unmarshal(data, &test2)
				assert.NoError(err, "Failed to unmarshal nested maps")
				assert.Equal(len(test2.Data["level1"]["level2"]), 3, "Nested array length")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestFieldsWithNumbers tests fields that start with numbers
func TestFieldsWithNumbers(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "AdsActionStats_Numeric_Fields",
			testFunc: func(t *testing.T) {
				// Test fields that start with numbers
				stats := &AdsActionStats{
					X1dClick:  "100",
					X28dClick: "500",
					X1dView:   "1000",
					X7dView:   "2000",
				}

				// Marshal to JSON - should use original field names
				data, err := json.Marshal(stats)
				assert := testutil.NewAssert(t)
				assert.NoError(err, "Failed to marshal AdsActionStats")

				// Check that JSON has original field names
				var m map[string]interface{}
				err = json.Unmarshal(data, &m)
				assert.NoError(err, "Failed to unmarshal to map")

				// Verify field name transformations
				_, has1dClick := m["1d_click"]
				_, has28dClick := m["28d_click"]
				_, has1dView := m["1d_view"]
				_, has7dView := m["7d_view"]

				assert.True(has1dClick, "JSON should have '1d_click' field")
				assert.True(has28dClick, "JSON should have '28d_click' field")
				assert.True(has1dView, "JSON should have '1d_view' field")
				assert.True(has7dView, "JSON should have '7d_view' field")

				// Verify values
				assert.Equal(m["1d_click"], "100", "1d_click value")
				assert.Equal(m["28d_click"], "500", "28d_click value")
			},
		},
		{
			name: "Round_Trip_Numeric_Fields",
			testFunc: func(t *testing.T) {
				// Test round trip marshaling/unmarshaling
				original := &AdsActionStats{
					X1dClick: "999",
					X7dClick: "777",
				}

				data, err := json.Marshal(original)
				assert := testutil.NewAssert(t)
				assert.NoError(err, "Failed to marshal")

				var result AdsActionStats
				err = json.Unmarshal(data, &result)
				assert.NoError(err, "Failed to unmarshal")

				assert.Equal(result.X1dClick, original.X1dClick, "1d_click round trip")
				assert.Equal(result.X7dClick, original.X7dClick, "7d_click round trip")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestZeroValues tests handling of zero/nil values
func TestZeroValues(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "Omitempty_Tags",
			testFunc: func(t *testing.T) {
				// Test that zero values are omitted with omitempty
				campaign := &Campaign{
					ID:          "camp123",
					Name:        "", // Empty string with omitempty
					DailyBudget: "", // Empty string with omitempty
				}

				data, err := json.Marshal(campaign)
				assert := testutil.NewAssert(t)
				assert.NoError(err, "Failed to marshal")

				var m map[string]interface{}
				err = json.Unmarshal(data, &m)
				assert.NoError(err, "Failed to unmarshal to map")

				// ID should be present
				_, hasID := m["id"]
				assert.True(hasID, "ID should be present")

				// Empty/zero values should be omitted
				_, hasName := m["name"]
				_, hasBudget := m["daily_budget"]
				assert.False(hasName, "Empty name should be omitted")
				assert.False(hasBudget, "Zero budget should be omitted")
			},
		},
		{
			name: "Nil_Slices_And_Maps",
			testFunc: func(t *testing.T) {
				type TestType struct {
					ID    string                 `json:"id"`
					Slice []string               `json:"slice,omitempty"`
					Map   map[string]interface{} `json:"map,omitempty"`
					Ptr   *AdLabel               `json:"ptr,omitempty"`
				}

				test := &TestType{
					ID:    "test123",
					Slice: nil,
					Map:   nil,
					Ptr:   nil,
				}

				data, err := json.Marshal(test)
				assert := testutil.NewAssert(t)
				assert.NoError(err, "Failed to marshal")

				var m map[string]interface{}
				err = json.Unmarshal(data, &m)
				assert.NoError(err, "Failed to unmarshal to map")

				// Only ID should be present
				assert.Equal(len(m), 1, "Only ID should be in JSON")
				assert.Equal(m["id"], "test123", "ID value")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

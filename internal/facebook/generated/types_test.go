package generated

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAdMarshaling(t *testing.T) {
	// Test creating and marshaling an Ad object
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

	// Marshal to JSON
	data, err := json.Marshal(ad)
	if err != nil {
		t.Fatalf("Failed to marshal Ad: %v", err)
	}

	// Unmarshal back
	var ad2 Ad
	if err := json.Unmarshal(data, &ad2); err != nil {
		t.Fatalf("Failed to unmarshal Ad: %v", err)
	}

	// Verify fields
	if ad2.ID != ad.ID {
		t.Errorf("ID mismatch: got %s, want %s", ad2.ID, ad.ID)
	}
	if ad2.Name != ad.Name {
		t.Errorf("Name mismatch: got %s, want %s", ad2.Name, ad.Name)
	}
}

func TestEnumTypes(t *testing.T) {
	// Test enum types
	var status AdStatus = "ACTIVE"

	// Should be able to use as string
	if string(status) != "ACTIVE" {
		t.Errorf("Enum conversion failed: got %s, want ACTIVE", status)
	}
}

func TestEmptyStruct(t *testing.T) {
	// Test empty struct like ProductItemInvalidationError
	item := &ProductItem{
		ID: "prod123",
		InvalidationErrors: []*ProductItemInvalidationError{
			{}, // Empty struct should work
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal ProductItem: %v", err)
	}

	// Should produce valid JSON
	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestComplexMapTypes(t *testing.T) {
	// Test complex map handling
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
	_, err := json.Marshal(template)
	if err != nil {
		t.Fatalf("Failed to marshal complex type: %v", err)
	}
}

func TestFieldsWithNumbers(t *testing.T) {
	// Test fields that start with numbers
	stats := &AdsActionStats{
		X1dClick:  "100",
		X28dClick: "500",
	}

	// Marshal to JSON - should use original field names
	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal AdsActionStats: %v", err)
	}

	// Check that JSON has original field names
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, ok := m["1d_click"]; !ok {
		t.Error("Expected JSON to have '1d_click' field")
	}
}

package generated

import (
	"context"
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

func TestTypedHandlers_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("GET", "/v23.0/123456789/", func(w http.ResponseWriter, r *http.Request) {
		// Check for access token
		if r.URL.Query().Get("access_token") == "" {
			env.Server().WriteAuthError(w)
			return
		}
		// Return a simple response
		env.Server().WriteSuccess(w, map[string]interface{}{
			"id":     "123456789",
			"name":   "Test Ad",
			"status": "ACTIVE",
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

	t.Run("GetAdHandler with typed arguments", func(t *testing.T) {
		args := ad_getArgs{
			ID:     "123456789",
			Fields: []string{"id", "name", "status"},
			Limit:  10,
		}

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"id":     args.ID,
					"fields": args.Fields,
					"limit":  args.Limit,
				},
			},
		}

		// Test without access token first
		oldToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
		os.Unsetenv("FACEBOOK_ACCESS_TOKEN")

		result, err := AdGetHandler(context.Background(), request, args)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		// Should get error due to missing token
		testutil.AssertResult(t, result).
			IsError().
			HasErrorContaining("FACEBOOK_ACCESS_TOKEN")

		// Now test with access token
		os.Setenv("FACEBOOK_ACCESS_TOKEN", "test_token")
		defer func() {
			if oldToken != "" {
				os.Setenv("FACEBOOK_ACCESS_TOKEN", oldToken)
			} else {
				os.Unsetenv("FACEBOOK_ACCESS_TOKEN")
			}
		}()

		result2, err := AdGetHandler(context.Background(), request, args)
		if err != nil {
			t.Fatalf("Handler returned error with token: %v", err)
		}

		// Should succeed now
		data := testutil.AssertResult(t, result2).
			IsSuccess().
			ParseJSON()

		if data["id"] != "123456789" {
			t.Errorf("Expected id '123456789', got %v", data["id"])
		}
	})
}

func TestTypedBatchHandler_WithFramework(t *testing.T) {
	// Test that we can create typed args
	args := ad_getArgs{
		ID:     "test123",
		Fields: []string{"id", "name"},
		Limit:  5,
	}

	// Verify args are correctly structured
	if args.ID != "test123" {
		t.Errorf("Expected ID 'test123', got '%s'", args.ID)
	}

	if len(args.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(args.Fields))
	}

	if args.Limit != 5 {
		t.Errorf("Expected limit 5, got %d", args.Limit)
	}
}

func TestTypedHandlersTableDriven_WithFramework(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func(*testutil.TestServer)
		setupEnv      func()
		cleanupEnv    func()
		args          ad_getArgs
		wantError     bool
		errorContains string
		validate      func(*testing.T, map[string]interface{})
	}{
		{
			name: "Success with all fields",
			setupServer: func(s *testutil.TestServer) {
				s.AddRoute("GET", "/v23.0/test_ad_123/", func(w http.ResponseWriter, r *http.Request) {
					s.WriteSuccess(w, testutil.CreateMockAdResponse("test_ad_123"))
				})
			},
			args: ad_getArgs{
				ID:     "test_ad_123",
				Fields: []string{"id", "name", "status", "creative", "adlabels"},
			},
			validate: func(t *testing.T, data map[string]interface{}) {
				if data["id"] != "test_ad_123" {
					t.Errorf("Expected id 'test_ad_123', got %v", data["id"])
				}
				if data["name"] != "Test Ad" {
					t.Errorf("Expected name 'Test Ad', got %v", data["name"])
				}
				if creative, ok := data["creative"].(map[string]interface{}); ok {
					if creative["id"] != testutil.TestCreativeID {
						t.Errorf("Expected creative id '%s', got %v", testutil.TestCreativeID, creative["id"])
					}
				} else {
					t.Error("Expected creative to be an object")
				}
			},
		},
		{
			name: "Error with invalid ID",
			setupServer: func(s *testutil.TestServer) {
				s.SetDefaultHandler(func(w http.ResponseWriter, r *http.Request) {
					s.WriteError(w, 400, "GraphMethodException", "Invalid ad ID")
				})
			},
			args: ad_getArgs{
				ID: "invalid_id",
			},
			wantError:     true,
			errorContains: "Invalid ad ID",
		},
		{
			name: "Missing access token",
			setupEnv: func() {
				os.Unsetenv("FACEBOOK_ACCESS_TOKEN")
			},
			cleanupEnv: func() {
				os.Setenv("FACEBOOK_ACCESS_TOKEN", testutil.TestAccessToken)
			},
			args: ad_getArgs{
				ID: "123",
			},
			wantError:     true,
			errorContains: "FACEBOOK_ACCESS_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.NewTestEnvironment(t)
			defer env.Teardown()

			// Setup server
			if tt.setupServer != nil {
				tt.setupServer(env.Server())
			}

			// Override URLs
			oldHost := graphAPIHost
			oldBase := baseGraphURL
			defer func() {
				graphAPIHost = oldHost
				baseGraphURL = oldBase
			}()
			graphAPIHost = env.Server().URL
			baseGraphURL = env.Server().URL

			// Setup environment
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			// Create request
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"id":     tt.args.ID,
						"fields": tt.args.Fields,
						"limit":  tt.args.Limit,
					},
				},
			}

			// Execute
			result, err := AdGetHandler(context.Background(), request, tt.args)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			// Assert
			assertion := testutil.AssertResult(t, result)
			if tt.wantError {
				assertion.IsError()
				if tt.errorContains != "" {
					assertion.HasErrorContaining(tt.errorContains)
				}
			} else {
				data := assertion.IsSuccess().ParseJSON()
				if tt.validate != nil {
					tt.validate(t, data)
				}
			}
		})
	}
}

func TestComplexTypedArgs_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("POST", "/v23.0/test_campaign/", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateSuccessResponse("test_campaign"))
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

	// Test with complex typed arguments
	args := campaign_updateArgs{
		ID:          "test_campaign",
		Name:        "Updated Campaign",
		Status:      "ACTIVE",
		DailyBudget: 500,
		Adlabels: []*AdLabel{
			{
				ID:   "label1",
				Name: "Test Label 1",
			},
			{
				ID:   "label2",
				Name: "Test Label 2",
			},
		},
		PromotedObject: &AdPromotedObject{
			PageID:          "page123",
			CustomEventType: "PURCHASE",
			ApplicationID:   "app456",
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id":           args.ID,
				"name":         args.Name,
				"status":       args.Status,
				"daily_budget": args.DailyBudget,
				"adlabels": []map[string]interface{}{
					{"id": "label1", "name": "Test Label 1"},
					{"id": "label2", "name": "Test Label 2"},
				},
				"promoted_object": map[string]interface{}{
					"page_id":           "page123",
					"custom_event_type": "PURCHASE",
					"application_id":    "app456",
				},
			},
		},
	}

	// Execute
	result, err := CampaignUpdateHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Assert success
	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	if success, ok := data["success"].(bool); !ok || !success {
		t.Error("Expected success=true")
	}
}

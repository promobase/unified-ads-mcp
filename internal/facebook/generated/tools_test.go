package generated

import (
	"context"
	"encoding/json"
	"io"
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

func TestListAdAccountActivitiesHandler_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("GET", "/v23.0/act_123456789/activities", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"event_time": "2024-01-01T12:00:00+0000",
					"event_type": "update_status",
					"extra_data": map[string]interface{}{
						"old_status": "ACTIVE",
						"new_status": "PAUSED",
					},
				},
				{
					"event_time": "2024-01-01T11:00:00+0000",
					"event_type": "create_campaign",
					"extra_data": map[string]interface{}{
						"campaign_id": "123456789",
					},
				},
			},
			"paging": map[string]interface{}{
				"cursors": map[string]string{
					"before": "BEFORE_CURSOR",
					"after":  "AFTER_CURSOR",
				},
			},
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
		name           string
		args           ad_account_list_activitiesArgs
		wantError      bool
		errorContains  string
		validateResult func(*testing.T, map[string]interface{})
	}{
		{
			name: "Success",
			args: ad_account_list_activitiesArgs{
				ID:     "act_123456789",
				Limit:  10,
				Fields: []string{"event_time", "event_type", "extra_data"},
			},
			validateResult: func(t *testing.T, data map[string]interface{}) {
				activities, ok := data["data"].([]interface{})
				if !ok {
					t.Error("Response missing 'data' field")
					return
				}
				if len(activities) != 2 {
					t.Errorf("Expected 2 activities, got %d", len(activities))
				}
			},
		},
		{
			name: "MissingID",
			args: ad_account_list_activitiesArgs{
				ID: "",
			},
			wantError:     true,
			errorContains: "id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"id":     tt.args.ID,
						"limit":  tt.args.Limit,
						"fields": tt.args.Fields,
					},
				},
			}

			result, err := AdAccountListActivitiesHandler(context.Background(), request, tt.args)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			assertion := testutil.AssertResult(t, result)
			if tt.wantError {
				assertion.IsError()
				if tt.errorContains != "" {
					assertion.HasErrorContaining(tt.errorContains)
				}
			} else {
				data := assertion.IsSuccess().ParseJSON()
				if tt.validateResult != nil {
					tt.validateResult(t, data)
				}
			}
		})
	}
}

func TestGetAdSetInsightsHandler_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock insights response
	env.Server().AddRoute("GET", "/v23.0/123456789/insights", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, testutil.CreateMockInsightsResponse())
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

	args := ad_set_get_insightsArgs{
		ID:         "123456789",
		DatePreset: "yesterday",
		Fields:     []string{"impressions", "clicks", "spend", "reach", "frequency"},
		Level:      "adset",
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id":          args.ID,
				"date_preset": args.DatePreset,
				"fields":      args.Fields,
				"level":       args.Level,
			},
		},
	}

	result, err := AdSetGetInsightsHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	// Verify insights data
	if insights, ok := data["data"].([]interface{}); !ok {
		t.Error("Response missing 'data' field")
	} else if len(insights) > 0 {
		insight := insights[0].(map[string]interface{})
		expectedFields := []string{"impressions", "clicks", "spend", "reach", "frequency"}
		for _, field := range expectedFields {
			if _, exists := insight[field]; !exists {
				t.Errorf("Insight missing expected field: %s", field)
			}
		}
	}
}

func TestCreateAdSetAdlabelHandler_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("POST", "/v23.0/123456789/adlabels", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var params map[string]interface{}
		json.Unmarshal(body, &params)

		env.Server().WriteSuccess(w, map[string]interface{}{
			"success":  true,
			"adlabels": params["adlabels"],
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

	args := ad_set_create_adlabelArgs{
		ID: "123456789",
		Adlabels: []*AdLabel{
			{
				Name: "Test Label",
				ID:   "label_123",
			},
		},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id": args.ID,
				"adlabels": []map[string]interface{}{
					{
						"name": "Test Label",
						"id":   "label_123",
					},
				},
			},
		},
	}

	result, err := AdSetCreateAdlabelHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	data := testutil.AssertResult(t, result).
		IsSuccess().
		ParseJSON()

	// Verify success
	if success, ok := data["success"].(bool); !ok || !success {
		t.Error("Expected success=true in response")
	}
}

func TestAPIErrorHandling_WithFramework(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup error response
	env.Server().SetDefaultHandler(func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteError(w, 400, "GraphMethodException", "Invalid parameter")
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

	args := ad_set_getArgs{
		ID: "invalid_id",
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id": args.ID,
			},
		},
	}

	result, err := AdSetGetHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	testutil.AssertResult(t, result).
		IsError().
		HasErrorContaining("GraphMethodException").
		HasErrorContaining("Invalid parameter")
}

func TestNoAccessToken_WithFramework(t *testing.T) {
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

	// Clear access token
	oldToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	os.Unsetenv("FACEBOOK_ACCESS_TOKEN")
	oldAccessToken := accessToken
	accessToken = ""
	defer func() {
		if oldToken != "" {
			os.Setenv("FACEBOOK_ACCESS_TOKEN", oldToken)
		}
		accessToken = oldAccessToken
	}()

	args := ad_account_list_activitiesArgs{
		ID: "act_123456789",
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id": args.ID,
			},
		},
	}

	result, err := AdAccountListActivitiesHandler(context.Background(), request, args)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	testutil.AssertResult(t, result).
		IsError().
		HasErrorContaining("FACEBOOK_ACCESS_TOKEN")
}

// Benchmark test using the framework
func BenchmarkListAdAccountActivitiesHandler_WithFramework(b *testing.B) {
	env := testutil.NewTestEnvironment(&testing.T{})
	defer env.Teardown()

	// Setup mock response
	env.Server().AddRoute("GET", "/v23.0/act_123456789/activities", func(w http.ResponseWriter, r *http.Request) {
		env.Server().WriteSuccess(w, map[string]interface{}{
			"data": []map[string]interface{}{
				{"event_time": "2024-01-01T12:00:00+0000", "event_type": "update_status"},
			},
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

	args := ad_account_list_activitiesArgs{
		ID:     "act_123456789",
		Limit:  10,
		Fields: []string{"event_time", "event_type", "extra_data"},
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id":     args.ID,
				"limit":  args.Limit,
				"fields": args.Fields,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AdAccountListActivitiesHandler(context.Background(), request, args)
	}
}

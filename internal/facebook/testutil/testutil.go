// Package testutil provides utilities for testing Facebook Graph API handlers
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Common test data constants
const (
	TestAccountID   = "act_123"
	TestCampaignID  = "1234321"
	TestAdsetID     = "12345"
	TestAdID        = "125475"
	TestPageID      = "13531"
	TestCreativeID  = "15742462"
	TestLabelID     = "label_123"
	TestAccessToken = "test_token"
)

// TestServer wraps httptest.Server with common Facebook Graph API mock functionality
type TestServer struct {
	*httptest.Server
	t              *testing.T
	routes         map[string]HandlerFunc
	defaultHandler HandlerFunc
}

// HandlerFunc is a function that handles HTTP requests
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// NewTestServer creates a new test server with standard authentication checking
func NewTestServer(t *testing.T) *TestServer {
	ts := &TestServer{
		t:      t,
		routes: make(map[string]HandlerFunc),
	}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request for debugging
		t.Logf("Mock API request: %s %s", r.Method, r.URL.Path)

		// Common auth check
		if !ts.checkAuth(w, r) {
			return
		}

		// Route to specific handlers
		key := r.Method + " " + r.URL.Path
		if handler, ok := ts.routes[key]; ok {
			handler(w, r)
			return
		}

		// Check for default handler
		if ts.defaultHandler != nil {
			ts.defaultHandler(w, r)
			return
		}

		// Default 404
		ts.WriteError(w, http.StatusNotFound, "GraphMethodException", "Endpoint not found")
	}))

	return ts
}

// AddRoute adds a handler for a specific method and path
func (ts *TestServer) AddRoute(method, path string, handler HandlerFunc) {
	ts.routes[method+" "+path] = handler
}

// SetDefaultHandler sets a default handler for unmatched routes
func (ts *TestServer) SetDefaultHandler(handler HandlerFunc) {
	ts.defaultHandler = handler
}

// checkAuth performs standard authentication checking
func (ts *TestServer) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	// Batch requests use form data
	if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/v23.0/") {
		if err := r.ParseForm(); err != nil {
			ts.WriteError(w, http.StatusBadRequest, "ParseError", "Failed to parse form")
			return false
		}
		if r.FormValue("access_token") == "" {
			ts.WriteAuthError(w)
			return false
		}
		return true
	}

	// Regular requests use query params
	if r.URL.Query().Get("access_token") == "" {
		ts.WriteAuthError(w)
		return false
	}
	return true
}

// WriteAuthError writes a standard authentication error
func (ts *TestServer) WriteAuthError(w http.ResponseWriter) {
	ts.WriteError(w, http.StatusUnauthorized, "OAuthException",
		"An access token is required to request this resource.")
}

// WriteError writes a standard error response
func (ts *TestServer) WriteError(w http.ResponseWriter, code int, errorType, message string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errorType,
			"message": message,
			"code":    code,
		},
	})
}

// WriteSuccess writes a successful JSON response
func (ts *TestServer) WriteSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// MockResponse provides common response structures
type MockResponse struct {
	ID      string                 `json:"id,omitempty"`
	Name    string                 `json:"name,omitempty"`
	Status  string                 `json:"status,omitempty"`
	Success bool                   `json:"success,omitempty"`
	Data    []interface{}          `json:"data,omitempty"`
	Error   *MockError             `json:"error,omitempty"`
	Fields  map[string]interface{} `json:"-"`
}

// MockError represents an error response
type MockError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// ResultAssertion provides fluent assertions for MCP results
type ResultAssertion struct {
	t      *testing.T
	result *mcp.CallToolResult
}

// AssertResult creates a new result assertion
func AssertResult(t *testing.T, result *mcp.CallToolResult) *ResultAssertion {
	return &ResultAssertion{t: t, result: result}
}

// IsSuccess asserts the result is not an error
func (ra *ResultAssertion) IsSuccess() *ResultAssertion {
	if ra.result == nil {
		ra.t.Fatal("Result is nil")
	}
	if ra.result.IsError {
		content := ra.extractTextContent()
		ra.t.Fatalf("Expected success, got error: %s", content)
	}
	return ra
}

// IsError asserts the result is an error
func (ra *ResultAssertion) IsError() *ResultAssertion {
	if ra.result == nil {
		ra.t.Fatal("Result is nil")
	}
	if !ra.result.IsError {
		ra.t.Error("Expected error result")
	}
	return ra
}

// HasErrorContaining asserts the error contains a string
func (ra *ResultAssertion) HasErrorContaining(substr string) *ResultAssertion {
	ra.IsError()
	content := ra.extractTextContent()
	if !strings.Contains(content, substr) {
		ra.t.Errorf("Expected error containing '%s', got: %s", substr, content)
	}
	return ra
}

// ParseJSON parses the result as JSON and returns it
func (ra *ResultAssertion) ParseJSON() map[string]interface{} {
	content := ra.extractTextContent()
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		ra.t.Fatalf("Failed to parse JSON result: %v\nContent: %s", err, content)
	}
	return data
}

// ParseJSONArray parses the result as a JSON array
func (ra *ResultAssertion) ParseJSONArray() []interface{} {
	content := ra.extractTextContent()
	var data []interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		ra.t.Fatalf("Failed to parse JSON array result: %v\nContent: %s", err, content)
	}
	return data
}

// GetContent returns the raw text content
func (ra *ResultAssertion) GetContent() string {
	return ra.extractTextContent()
}

// extractTextContent extracts text from the first content item
func (ra *ResultAssertion) extractTextContent() string {
	if len(ra.result.Content) == 0 {
		ra.t.Fatal("Result has no content")
	}
	textContent, ok := mcp.AsTextContent(ra.result.Content[0])
	if !ok {
		ra.t.Fatal("Expected text content in result")
	}
	return textContent.Text
}

// TestCase provides a structured way to define test cases
type TestCase struct {
	Name           string
	Handler        interface{}            // The handler function to test
	Args           interface{}            // The typed args struct
	RequestArgs    map[string]interface{} // Raw request arguments
	Setup          func(*TestServer)
	ExpectError    bool
	ExpectContains string
	Validate       func(*testing.T, *mcp.CallToolResult)
}

// TestEnvironment manages test environment setup and teardown
type TestEnvironment struct {
	t             *testing.T
	server        *TestServer
	originalHost  string
	originalBase  string
	originalToken string
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	env := &TestEnvironment{
		t:      t,
		server: NewTestServer(t),
	}
	env.Setup()
	return env
}

// Setup configures the test environment
func (env *TestEnvironment) Setup() {
	// Store original values - we need to import these from generated package
	// For now, we'll use environment variables
	env.originalToken = os.Getenv("FACEBOOK_ACCESS_TOKEN")

	// Set test values
	os.Setenv("FACEBOOK_ACCESS_TOKEN", TestAccessToken)
}

// Teardown restores the original environment
func (env *TestEnvironment) Teardown() {
	env.server.Close()

	// Restore original values
	if env.originalToken != "" {
		os.Setenv("FACEBOOK_ACCESS_TOKEN", env.originalToken)
	} else {
		os.Unsetenv("FACEBOOK_ACCESS_TOKEN")
	}
}

// Server returns the test server
func (env *TestEnvironment) Server() *TestServer {
	return env.server
}

// WithGraphAPIOverride temporarily overrides the Graph API host and base URL
type WithGraphAPIOverride struct {
	Host string
	Base string
}

// CreateMockCampaignResponse creates a standard campaign response
func CreateMockCampaignResponse(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":                  id,
		"name":                "Test Campaign",
		"status":              "PAUSED",
		"objective":           "LINK_CLICKS",
		"buying_type":         "AUCTION",
		"daily_budget":        "200",
		"lifetime_budget":     "10000",
		"bid_strategy":        "LOWEST_COST_WITHOUT_CAP",
		"special_ad_category": "EMPLOYMENT",
		"promoted_object": map[string]interface{}{
			"page_id": TestPageID,
		},
		"adlabels": []map[string]interface{}{
			{
				"id":   TestLabelID,
				"name": "Test Label",
			},
		},
	}
}

// CreateMockAdSetResponse creates a standard ad set response
func CreateMockAdSetResponse(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":                id,
		"name":              "Test AdSet",
		"status":            "PAUSED",
		"campaign_id":       TestCampaignID,
		"daily_budget":      "200",
		"lifetime_budget":   "10000",
		"bid_strategy":      "LOWEST_COST_WITHOUT_CAP",
		"optimization_goal": "LINK_CLICKS",
		"targeting": map[string]interface{}{
			"geo_locations": map[string]interface{}{
				"countries": []string{"US"},
			},
		},
		"promoted_object": map[string]interface{}{
			"page_id": TestPageID,
		},
		"adlabels": []map[string]interface{}{
			{
				"id":   "label_456",
				"name": "AdSet Label",
			},
		},
	}
}

// CreateMockAdResponse creates a standard ad response
func CreateMockAdResponse(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":          id,
		"name":        "Test Ad",
		"status":      "PAUSED",
		"campaign_id": TestCampaignID,
		"adset_id":    TestAdsetID,
		"creative": map[string]interface{}{
			"id":   TestCreativeID,
			"name": "test creative",
		},
		"adlabels": []map[string]interface{}{
			{
				"id":   "label_789",
				"name": "Ad Label",
			},
		},
	}
}

// CreateMockInsightsResponse creates a standard insights response
func CreateMockInsightsResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"date_start":   "2024-01-01",
				"date_stop":    "2024-01-01",
				"impressions":  "1000",
				"clicks":       "50",
				"spend":        "25.50",
				"reach":        "800",
				"frequency":    "1.25",
				"account_id":   TestAccountID,
				"account_name": "Test Account",
			},
		},
		"paging": map[string]interface{}{
			"cursors": map[string]interface{}{
				"before": "BEFORE_CURSOR",
				"after":  "AFTER_CURSOR",
			},
		},
	}
}

// CreateSuccessResponse creates a standard success response
func CreateSuccessResponse(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":      id,
		"success": true,
	}
}

// CallHandler is a helper to call handlers with proper context and request setup
func CallHandler(ctx context.Context, handler interface{}, args interface{}) (*mcp.CallToolResult, error) {
	// This is a placeholder - handlers need to be called directly in tests
	// The function signature is kept for future implementation with reflection
	return nil, fmt.Errorf("CallHandler not implemented - call handlers directly")
}

// Assert provides fluent assertion methods for tests
type Assert struct {
	t *testing.T
}

// NewAssert creates a new Assert instance
func NewAssert(t *testing.T) *Assert {
	return &Assert{t: t}
}

// Equal asserts that two values are equal
func (a *Assert) Equal(actual, expected interface{}, message string) *Assert {
	if actual != expected {
		a.t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
	return a
}

// EqualInt asserts that two int values are equal
func (a *Assert) EqualInt(actual, expected int, message string) *Assert {
	if actual != expected {
		a.t.Errorf("%s: expected %d, got %d", message, expected, actual)
	}
	return a
}

// NotNil asserts that a value is not nil
func (a *Assert) NotNil(value interface{}, message string) *Assert {
	if value == nil {
		a.t.Errorf("%s: expected non-nil value", message)
	}
	return a
}

// True asserts that a boolean value is true
func (a *Assert) True(value bool, message string) *Assert {
	if !value {
		a.t.Errorf("%s: expected true, got false", message)
	}
	return a
}

// False asserts that a boolean value is false
func (a *Assert) False(value bool, message string) *Assert {
	if value {
		a.t.Errorf("%s: expected false, got true", message)
	}
	return a
}

// NoError asserts that an error is nil
func (a *Assert) NoError(err error, message string) *Assert {
	if err != nil {
		a.t.Errorf("%s: unexpected error: %v", message, err)
	}
	return a
}

// CreateMockHandler creates a handler that returns the given response
func CreateMockHandler(response interface{}) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSONSuccess(w, response)
	}
}

// WriteJSONSuccess writes a successful JSON response
func WriteJSONSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// WriteJSONError writes a JSON error response
func WriteJSONError(w http.ResponseWriter, code int, errorType, message string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errorType,
			"message": message,
			"code":    code,
		},
	})
}

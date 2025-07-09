package testutil

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandlerTestFunc is a function type that represents a handler to be tested
type HandlerTestFunc func(ctx context.Context, request mcp.CallToolRequest, args interface{}) (*mcp.CallToolResult, error)

// HandlerTestConfig configures how to test a handler
type HandlerTestConfig struct {
	// SetupServer is called to configure the test server before each test
	SetupServer func(*TestServer)

	// DefaultFields to request if not specified in test case
	DefaultFields []string

	// RequiredEnvVars that should be set for the handler
	RequiredEnvVars map[string]string
}

// HandlerTestSuite provides a higher-order function approach to testing handlers
type HandlerTestSuite struct {
	t      *testing.T
	config HandlerTestConfig
}

// NewHandlerTestSuite creates a new handler test suite
func NewHandlerTestSuite(t *testing.T, config HandlerTestConfig) *HandlerTestSuite {
	return &HandlerTestSuite{
		t:      t,
		config: config,
	}
}

// TestHandler runs a set of test cases against a handler using higher-order functions
func (suite *HandlerTestSuite) TestHandler(handler HandlerTestFunc, cases []TestCase) {
	for _, tc := range cases {
		suite.t.Run(tc.Name, func(t *testing.T) {
			// Create test environment
			env := NewTestEnvironment(t)
			defer env.Teardown()

			// Apply required env vars
			for k, v := range suite.config.RequiredEnvVars {
				os.Setenv(k, v)
			}

			// Setup server with default config
			if suite.config.SetupServer != nil {
				suite.config.SetupServer(env.Server())
			}

			// Apply test-specific setup
			if tc.Setup != nil {
				tc.Setup(env.Server())
			}

			// Create request
			ctx := context.Background()
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tc.RequestArgs,
				},
			}

			// Execute handler
			result, err := handler(ctx, request, tc.Args)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			// Assert results
			assertion := AssertResult(t, result)
			if tc.ExpectError {
				assertion.IsError()
				if tc.ExpectContains != "" {
					assertion.HasErrorContaining(tc.ExpectContains)
				}
			} else {
				assertion.IsSuccess()
			}

			// Run custom validation
			if tc.Validate != nil {
				tc.Validate(t, result)
			}
		})
	}
}

// CreateStandardGetTests creates standard test cases for GET handlers
func CreateStandardGetTests(objectType string, objectID string) []TestCase {
	return []TestCase{
		{
			Name: "Success",
			RequestArgs: map[string]interface{}{
				"id":     objectID,
				"fields": []string{"id", "name", "status"},
			},
			ExpectError: false,
			Validate: func(t *testing.T, result *mcp.CallToolResult) {
				data := AssertResult(t, result).ParseJSON()
				if data["id"] != objectID {
					t.Errorf("Expected id %s, got %v", objectID, data["id"])
				}
			},
		},
		{
			Name: "MissingID",
			RequestArgs: map[string]interface{}{
				"id":     "",
				"fields": []string{"id", "name"},
			},
			ExpectError:    true,
			ExpectContains: "id is required",
		},
		{
			Name: "NoAccessToken",
			RequestArgs: map[string]interface{}{
				"id": objectID,
			},
			Setup: func(s *TestServer) {
				os.Unsetenv("FACEBOOK_ACCESS_TOKEN")
			},
			ExpectError:    true,
			ExpectContains: "FACEBOOK_ACCESS_TOKEN",
		},
	}
}

// CreateStandardUpdateTests creates standard test cases for UPDATE handlers
func CreateStandardUpdateTests(objectType string, objectID string) []TestCase {
	return []TestCase{
		{
			Name: "UpdateSuccess",
			RequestArgs: map[string]interface{}{
				"id":     objectID,
				"name":   "Updated Name",
				"status": "ACTIVE",
			},
			ExpectError: false,
			Validate: func(t *testing.T, result *mcp.CallToolResult) {
				data := AssertResult(t, result).ParseJSON()
				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("Expected success=true in response")
				}
			},
		},
		{
			Name: "UpdateMissingID",
			RequestArgs: map[string]interface{}{
				"name": "Updated Name",
			},
			ExpectError:    true,
			ExpectContains: "id is required",
		},
	}
}

// CreateStandardDeleteTests creates standard test cases for DELETE handlers
func CreateStandardDeleteTests(objectType string, objectID string) []TestCase {
	return []TestCase{
		{
			Name: "DeleteSuccess",
			RequestArgs: map[string]interface{}{
				"id": objectID,
			},
			ExpectError: false,
			Validate: func(t *testing.T, result *mcp.CallToolResult) {
				data := AssertResult(t, result).ParseJSON()
				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("Expected success=true in response")
				}
			},
		},
		{
			Name: "DeleteMissingID",
			RequestArgs: map[string]interface{}{
				"id": "",
			},
			ExpectError:    true,
			ExpectContains: "id is required",
		},
	}
}

// WithMockResponse returns a setup function that adds a mock response for a route
func WithMockResponse(method, path string, response interface{}) func(*TestServer) {
	return func(s *TestServer) {
		s.AddRoute(method, path, func(w http.ResponseWriter, r *http.Request) {
			s.WriteSuccess(w, response)
		})
	}
}

// WithErrorResponse returns a setup function that adds an error response for a route
func WithErrorResponse(method, path string, code int, errorType, message string) func(*TestServer) {
	return func(s *TestServer) {
		s.AddRoute(method, path, func(w http.ResponseWriter, r *http.Request) {
			s.WriteError(w, code, errorType, message)
		})
	}
}

// CombineSetups combines multiple setup functions into one
func CombineSetups(setups ...func(*TestServer)) func(*TestServer) {
	return func(s *TestServer) {
		for _, setup := range setups {
			setup(s)
		}
	}
}

// TestBatchHandler provides specialized testing for batch handlers
type TestBatchHandler struct {
	suite *HandlerTestSuite
}

// NewTestBatchHandler creates a batch handler tester
func NewTestBatchHandler(t *testing.T) *TestBatchHandler {
	config := HandlerTestConfig{
		RequiredEnvVars: map[string]string{
			"FACEBOOK_ACCESS_TOKEN": TestAccessToken,
		},
	}
	return &TestBatchHandler{
		suite: NewHandlerTestSuite(t, config),
	}
}

// TestBatchRequest tests a batch request handler
func (tb *TestBatchHandler) TestBatchRequest(handler HandlerTestFunc, requests []interface{}, expectedResponses []interface{}) {
	cases := []TestCase{
		{
			Name: "BatchSuccess",
			RequestArgs: map[string]interface{}{
				"requests": requests,
			},
			Setup: func(s *TestServer) {
				// Mock batch endpoint
				s.AddRoute("POST", "/v23.0/", func(w http.ResponseWriter, r *http.Request) {
					// Return mocked batch responses
					s.WriteSuccess(w, expectedResponses)
				})
			},
			Validate: func(t *testing.T, result *mcp.CallToolResult) {
				data := AssertResult(t, result).ParseJSONArray()
				if len(data) != len(expectedResponses) {
					t.Errorf("Expected %d responses, got %d", len(expectedResponses), len(data))
				}
			},
		},
		{
			Name: "EmptyBatch",
			RequestArgs: map[string]interface{}{
				"requests": []interface{}{},
			},
			ExpectError:    true,
			ExpectContains: "empty",
		},
		{
			Name: "TooManyRequests",
			RequestArgs: map[string]interface{}{
				"requests": make([]interface{}, 51), // Over 50 limit
			},
			ExpectError:    true,
			ExpectContains: "50",
		},
	}

	tb.suite.TestHandler(handler, cases)
}

// MockBatchResponse creates a mock batch response
type MockBatchResponse struct {
	Code int                    `json:"code"`
	Body map[string]interface{} `json:"body"`
}

// CreateBatchSuccessResponse creates a successful batch response
func CreateBatchSuccessResponse(body interface{}) MockBatchResponse {
	return MockBatchResponse{
		Code: 200,
		Body: body.(map[string]interface{}),
	}
}

// CreateBatchErrorResponse creates an error batch response
func CreateBatchErrorResponse(code int, errorType, message string) MockBatchResponse {
	return MockBatchResponse{
		Code: code,
		Body: map[string]interface{}{
			"error": map[string]interface{}{
				"type":    errorType,
				"message": message,
			},
		},
	}
}

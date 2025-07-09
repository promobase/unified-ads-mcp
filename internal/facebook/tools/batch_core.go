// Core batch operations for Facebook API

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// BatchOperationArgs defines arguments for a single operation in a batch
type BatchOperationArgs struct {
	Method      string                 `json:"method" jsonschema:"required,description=HTTP method (GET/POST/PUT/DELETE),enum=GET,enum=POST,enum=PUT,enum=DELETE"`
	RelativeURL string                 `json:"relative_url" jsonschema:"required,description=Relative URL path (e.g. '123456789' or '123456789/ads')"`
	Body        map[string]interface{} `json:"body,omitempty" jsonschema:"description=Request body for POST/PUT requests"`
	Headers     map[string]string      `json:"headers,omitempty" jsonschema:"description=Custom headers for this request"`
	Name        string                 `json:"name,omitempty" jsonschema:"description=Optional name for referencing this operation in responses"`
}

// BatchRequestArgs defines the typed arguments for the batch tool
type BatchRequestArgs struct {
	Operations []BatchOperationArgs `json:"operations" jsonschema:"required,description=Array of operations to execute (max 50),minItems=1,maxItems=50"`
}

// BatchOperationResult represents the result of a single operation
type BatchOperationResult struct {
	Code       int                    `json:"code"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       json.RawMessage        `json:"body,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	ParsedBody map[string]interface{} `json:"parsed_body,omitempty"`
}

// BatchResult represents the complete batch operation result
type BatchResult struct {
	TotalOperations      int                    `json:"total_operations"`
	SuccessfulOperations int                    `json:"successful_operations"`
	FailedOperations     int                    `json:"failed_operations"`
	Results              []BatchOperationResult `json:"results"`
	Summary              map[string]interface{} `json:"summary"`
}

// BatchRequest represents a single request in a batch
type BatchRequest struct {
	Method      string                 `json:"method"`
	RelativeURL string                 `json:"relative_url"`
	Body        map[string]interface{} `json:"body,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Name        string                 `json:"name,omitempty"`
}

// BatchResponse represents the response from a single batch operation
type BatchResponse struct {
	Code    int               `json:"code"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    json.RawMessage   `json:"body,omitempty"`
}

// Graph API configuration
const graphAPIVersion = "v23.0"

var (
	graphAPIHost = "https://graph.facebook.com"
	accessToken  = ""
)

// getAccessToken retrieves the Facebook access token from environment or test setting
func getAccessToken() string {
	if accessToken != "" {
		return accessToken
	}
	return os.Getenv("FACEBOOK_ACCESS_TOKEN")
}

// SetAccessToken sets the access token for testing purposes
func SetAccessToken(token string) {
	accessToken = token
}

// SetGraphAPIHost sets the Graph API host for testing purposes
func SetGraphAPIHost(host string) {
	graphAPIHost = host
}

// MakeBatchRequest performs a batch request to the Facebook Graph API
func MakeBatchRequest(requests []BatchRequest) ([]BatchResponse, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests provided")
	}
	if len(requests) > 50 {
		return nil, fmt.Errorf("batch request limited to 50 requests, got %d", len(requests))
	}

	// Check access token
	accessToken := getAccessToken()
	if accessToken == "" {
		return nil, fmt.Errorf("FACEBOOK_ACCESS_TOKEN environment variable is not set")
	}

	// Convert batch requests to JSON
	batchJSON, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch requests: %w", err)
	}

	// Create form data
	formData := url.Values{}
	formData.Set("batch", string(batchJSON))
	formData.Set("access_token", accessToken)

	// Build URL
	batchURL := fmt.Sprintf("%s/%s/", graphAPIHost, graphAPIVersion)
	log.Printf("[DEBUG] Making batch request to %s with %d requests", batchURL, len(requests))

	// Test guardrail: panic if hitting real Facebook API during tests
	if strings.Contains(batchURL, "facebook.com") && os.Getenv("TESTING") == "true" {
		panic(fmt.Sprintf("TEST GUARDRAIL: Attempted to hit real Facebook API during tests! URL: %s", batchURL))
	}

	// Create request
	req, err := http.NewRequest("POST", batchURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("batch request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read batch response: %w", err)
	}

	// Parse response
	var batchResponses []BatchResponse
	if err := json.Unmarshal(body, &batchResponses); err != nil {
		return nil, fmt.Errorf("failed to parse batch response: %w", err)
	}

	return batchResponses, nil
}

// FacebookBatchHandler handles batch requests with enhanced error handling and result processing
func FacebookBatchHandler(ctx context.Context, request mcp.CallToolRequest, args BatchRequestArgs) (*mcp.CallToolResult, error) {
	// Validate input
	if len(args.Operations) == 0 {
		return mcp.NewToolResultError("At least one operation is required"), nil
	}

	if len(args.Operations) > 50 {
		return mcp.NewToolResultError("Maximum 50 operations allowed per batch"), nil
	}

	// Convert to internal batch request format
	batchRequests := make([]BatchRequest, len(args.Operations))
	for i, op := range args.Operations {
		batchRequests[i] = BatchRequest{
			Method:      op.Method,
			RelativeURL: op.RelativeURL,
			Body:        op.Body,
			Headers:     op.Headers,
			Name:        op.Name,
		}
	}

	// Execute batch request
	responses, err := MakeBatchRequest(batchRequests)
	if err != nil {
		return mcp.NewToolResultErrorf("Batch request failed: %v", err), nil
	}

	// Process results with enhanced error handling
	result := BatchResult{
		TotalOperations: len(args.Operations),
		Results:         make([]BatchOperationResult, len(responses)),
		Summary:         make(map[string]interface{}),
	}

	successCount := 0
	failedCount := 0

	for i, resp := range responses {
		opResult := BatchOperationResult{
			Code:    resp.Code,
			Headers: resp.Headers,
			Body:    resp.Body,
			Success: resp.Code >= 200 && resp.Code < 300,
		}

		// Add operation name if provided
		if i < len(args.Operations) && args.Operations[i].Name != "" {
			opResult.Name = args.Operations[i].Name
		}

		// Try to parse response body
		if len(resp.Body) > 0 {
			var parsedBody map[string]interface{}
			if err := json.Unmarshal(resp.Body, &parsedBody); err == nil {
				opResult.ParsedBody = parsedBody
			}
		}

		// Handle errors
		if !opResult.Success {
			failedCount++
			if opResult.ParsedBody != nil {
				if errData, ok := opResult.ParsedBody["error"].(map[string]interface{}); ok {
					if message, ok := errData["message"].(string); ok {
						opResult.Error = message
					}
				}
			}
			if opResult.Error == "" {
				opResult.Error = fmt.Sprintf("HTTP %d error", resp.Code)
			}
		} else {
			successCount++
		}

		result.Results[i] = opResult
	}

	result.SuccessfulOperations = successCount
	result.FailedOperations = failedCount

	// Add summary information
	result.Summary["success_rate"] = float64(successCount) / float64(len(args.Operations))
	result.Summary["total_operations"] = len(args.Operations)
	result.Summary["successful_operations"] = successCount
	result.Summary["failed_operations"] = failedCount

	// Convert to JSON for response
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to serialize batch result: %v", err), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// RegisterCoreBatchTools registers the core batch tool with the MCP server
func RegisterCoreBatchTools(s *server.MCPServer) error {
	// Register facebook_batch using raw JSON schema
	facebookBatchTool := mcp.NewToolWithRawSchema(
		"facebook_batch",
		"Execute multiple Facebook Graph API operations in a single batch request. Supports GET, POST, PUT, DELETE operations with enhanced error handling and result processing.",
		json.RawMessage(`{
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"operations": {
					"type": "array",
					"description": "Array of operations to execute (max 50)",
					"minItems": 1,
					"maxItems": 50,
					"items": {
						"type": "object",
						"additionalProperties": false,
						"properties": {
							"method": {
								"type": "string",
								"description": "HTTP method (GET/POST/PUT/DELETE)",
								"enum": ["GET", "POST", "PUT", "DELETE"]
							},
							"relative_url": {
								"type": "string",
								"description": "Relative URL path (e.g. '123456789' or '123456789/ads')"
							},
							"body": {
								"type": "object",
								"description": "Request body for POST/PUT requests",
								"additionalProperties": true
							},
							"headers": {
								"type": "object",
								"description": "Custom headers for this request",
								"additionalProperties": {"type": "string"}
							},
							"name": {
								"type": "string",
								"description": "Optional name for referencing this operation in responses"
							}
						},
						"required": ["method", "relative_url"]
					}
				}
			},
			"required": ["operations"]
		}`),
	)

	s.AddTool(facebookBatchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args BatchRequestArgs
		if err := request.BindArguments(&args); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid arguments: %v", err)), nil
		}
		return FacebookBatchHandler(ctx, request, args)
	})

	return nil
}

package testutil

import (
	"encoding/json"
	"net/http"
	"testing"
)

// BatchRequest represents a batch request
type BatchRequest struct {
	Method      string                 `json:"method"`
	RelativeURL string                 `json:"relative_url"`
	Body        map[string]interface{} `json:"body,omitempty"`
}

// BatchResponse represents a batch response
type BatchResponse struct {
	Code int             `json:"code"`
	Body json.RawMessage `json:"body"`
}

// BatchTestBuilder helps build batch request tests
type BatchTestBuilder struct {
	t        *testing.T
	requests []BatchRequest
}

// NewBatchTestBuilder creates a new batch test builder
func NewBatchTestBuilder(t *testing.T) *BatchTestBuilder {
	return &BatchTestBuilder{
		t:        t,
		requests: []BatchRequest{},
	}
}

// AddGetRequest adds a GET request to the batch
func (b *BatchTestBuilder) AddGetRequest(path string, params map[string]string) *BatchTestBuilder {
	url := path
	if len(params) > 0 {
		url += "?"
		first := true
		for k, v := range params {
			if !first {
				url += "&"
			}
			url += k + "=" + v
			first = false
		}
	}

	b.requests = append(b.requests, BatchRequest{
		Method:      "GET",
		RelativeURL: url,
	})
	return b
}

// AddPostRequest adds a POST request to the batch
func (b *BatchTestBuilder) AddPostRequest(path string, body map[string]interface{}) *BatchTestBuilder {
	b.requests = append(b.requests, BatchRequest{
		Method:      "POST",
		RelativeURL: path,
		Body:        body,
	})
	return b
}

// AddDeleteRequest adds a DELETE request to the batch
func (b *BatchTestBuilder) AddDeleteRequest(path string) *BatchTestBuilder {
	b.requests = append(b.requests, BatchRequest{
		Method:      "DELETE",
		RelativeURL: path,
	})
	return b
}

// Build returns the batch requests
func (b *BatchTestBuilder) Build() []BatchRequest {
	return b.requests
}

// CreateMockBatchHandler creates a handler for batch requests
func CreateMockBatchHandler(responses map[string]interface{}) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse batch parameter
		if err := r.ParseForm(); err != nil {
			WriteJSONError(w, 400, "ParseError", "Failed to parse form")
			return
		}

		batchJSON := r.FormValue("batch")
		if batchJSON == "" {
			WriteJSONError(w, 400, "ValidationError", "Missing batch parameter")
			return
		}

		var requests []BatchRequest
		if err := json.Unmarshal([]byte(batchJSON), &requests); err != nil {
			WriteJSONError(w, 400, "JSONError", "Invalid batch JSON")
			return
		}

		// Create responses based on requests
		batchResponses := make([]BatchResponse, len(requests))
		for i, req := range requests {
			// Check if we have a mock response for this path
			if resp, ok := responses[req.RelativeURL]; ok {
				respJSON, _ := json.Marshal(resp)
				batchResponses[i] = BatchResponse{
					Code: 200,
					Body: respJSON,
				}
			} else {
				// Default response
				defaultResp := map[string]interface{}{
					"id":   req.RelativeURL,
					"name": "Test Object",
				}
				respJSON, _ := json.Marshal(defaultResp)
				batchResponses[i] = BatchResponse{
					Code: 200,
					Body: respJSON,
				}
			}
		}

		WriteJSONSuccess(w, batchResponses)
	}
}

// BatchResponseBuilder helps build expected batch responses
type BatchResponseBuilder struct {
	responses []BatchResponse
}

// NewBatchResponseBuilder creates a new batch response builder
func NewBatchResponseBuilder() *BatchResponseBuilder {
	return &BatchResponseBuilder{
		responses: []BatchResponse{},
	}
}

// AddSuccess adds a successful response
func (b *BatchResponseBuilder) AddSuccess(data interface{}) *BatchResponseBuilder {
	body, _ := json.Marshal(data)
	b.responses = append(b.responses, BatchResponse{
		Code: 200,
		Body: body,
	})
	return b
}

// AddError adds an error response
func (b *BatchResponseBuilder) AddError(code int, errorType, message string) *BatchResponseBuilder {
	errorData := map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errorType,
			"message": message,
			"code":    code,
		},
	}
	body, _ := json.Marshal(errorData)
	b.responses = append(b.responses, BatchResponse{
		Code: code,
		Body: body,
	})
	return b
}

// Build returns the batch responses
func (b *BatchResponseBuilder) Build() []BatchResponse {
	return b.responses
}

// AssertBatchResponse provides assertions for batch responses
type AssertBatchResponse struct {
	t         *testing.T
	responses []BatchResponse
}

// NewAssertBatchResponse creates batch response assertions
func NewAssertBatchResponse(t *testing.T, responses []BatchResponse) *AssertBatchResponse {
	return &AssertBatchResponse{t: t, responses: responses}
}

// HasCount asserts the number of responses
func (a *AssertBatchResponse) HasCount(expected int) *AssertBatchResponse {
	if len(a.responses) != expected {
		a.t.Errorf("Expected %d responses, got %d", expected, len(a.responses))
	}
	return a
}

// ResponseAt returns assertions for a specific response
func (a *AssertBatchResponse) ResponseAt(index int) *BatchResponseAssertion {
	if index >= len(a.responses) {
		a.t.Fatalf("Response index %d out of bounds (have %d responses)", index, len(a.responses))
	}
	return &BatchResponseAssertion{
		t:        a.t,
		response: &a.responses[index],
	}
}

// AllSuccessful asserts all responses are successful
func (a *AssertBatchResponse) AllSuccessful() *AssertBatchResponse {
	for i, resp := range a.responses {
		if resp.Code != 200 {
			a.t.Errorf("Response %d has code %d, expected 200", i, resp.Code)
		}
	}
	return a
}

// BatchResponseAssertion provides assertions for a single batch response
type BatchResponseAssertion struct {
	t        *testing.T
	response *BatchResponse
}

// HasCode asserts the response code
func (b *BatchResponseAssertion) HasCode(expected int) *BatchResponseAssertion {
	if b.response.Code != expected {
		b.t.Errorf("Expected response code %d, got %d", expected, b.response.Code)
	}
	return b
}

// HasField asserts a field exists in the response body
func (b *BatchResponseAssertion) HasField(field string, expected interface{}) *BatchResponseAssertion {
	var body map[string]interface{}
	if err := json.Unmarshal(b.response.Body, &body); err != nil {
		b.t.Fatalf("Failed to parse response body: %v", err)
	}

	if val, ok := body[field]; !ok {
		b.t.Errorf("Response missing field '%s'", field)
	} else if val != expected {
		b.t.Errorf("Expected field '%s' to be %v, got %v", field, expected, val)
	}
	return b
}

// IsError asserts the response is an error
func (b *BatchResponseAssertion) IsError() *BatchResponseAssertion {
	if b.response.Code == 200 {
		b.t.Error("Expected error response, got success")
	}
	return b
}

// IsSuccess asserts the response is successful
func (b *BatchResponseAssertion) IsSuccess() *BatchResponseAssertion {
	if b.response.Code != 200 {
		b.t.Errorf("Expected success response, got code %d", b.response.Code)
	}
	return b
}

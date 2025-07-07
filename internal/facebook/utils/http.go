package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// FacebookAPIResponse represents the standard Facebook API response structure
type FacebookAPIResponse struct {
	Data   interface{}       `json:"data,omitempty"`
	Error  *FacebookAPIError `json:"error,omitempty"`
	Paging *PagingInfo       `json:"paging,omitempty"`
}

// FacebookAPIError represents Facebook API error structure
type FacebookAPIError struct {
	Message      string `json:"message"`
	Type         string `json:"type"`
	Code         int    `json:"code"`
	ErrorSubcode int    `json:"error_subcode,omitempty"`
	FBTraceID    string `json:"fbtrace_id,omitempty"`
}

// PagingInfo represents pagination information
type PagingInfo struct {
	Cursors *struct {
		Before string `json:"before"`
		After  string `json:"after"`
	} `json:"cursors,omitempty"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// HTTPClient is a reusable HTTP client for Facebook API
var HTTPClient = &http.Client{}

// ExecuteAPIRequest executes a Facebook API request with proper error handling
func ExecuteAPIRequest(method, baseURL string, params url.Values) (interface{}, error) {
	var resp *http.Response
	var err error

	switch method {
	case "GET":
		resp, err = HTTPClient.Get(baseURL + "?" + params.Encode())
	case "POST":
		resp, err = HTTPClient.PostForm(baseURL, params)
	case "DELETE":
		req, err := http.NewRequest("DELETE", baseURL+"?"+params.Encode(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create DELETE request: %w", err)
		}
		resp, err = HTTPClient.Do(req)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// If we can't parse as JSON, return the raw body as string
		return string(body), nil
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		// Try to extract error from parsed result
		if errMsg := ExtractAPIError(result, resp.StatusCode); errMsg != nil {
			return nil, errMsg
		}
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	return result, nil
}

// ExtractAPIError extracts a detailed error message from Facebook API response
func ExtractAPIError(result interface{}, httpStatus int) error {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil
	}

	errorInfo, ok := resultMap["error"].(map[string]interface{})
	if !ok {
		return nil
	}

	fbError := &FacebookAPIError{
		Code:    httpStatus,
		Type:    "unknown",
		Message: "unknown error",
	}

	if msg, ok := errorInfo["message"].(string); ok {
		fbError.Message = msg
	}
	if code, ok := errorInfo["code"].(float64); ok {
		fbError.Code = int(code)
	}
	if errorType, ok := errorInfo["type"].(string); ok {
		fbError.Type = errorType
	}
	if subcode, ok := errorInfo["error_subcode"].(float64); ok {
		fbError.ErrorSubcode = int(subcode)
	}
	if traceID, ok := errorInfo["fbtrace_id"].(string); ok {
		fbError.FBTraceID = traceID
	}

	return fmt.Errorf("Facebook API error: %s (code: %d, type: %s, http_status: %d)",
		fbError.Message, fbError.Code, fbError.Type, httpStatus)
}

// BuildURLParams builds URL parameters from a map, handling special cases
func BuildURLParams(accessToken string, args map[string]interface{}, skipParams ...string) url.Values {
	params := url.Values{}
	params.Set("access_token", accessToken)

	// Create a map of parameters to skip for quick lookup
	skipMap := make(map[string]bool)
	for _, param := range skipParams {
		skipMap[param] = true
	}

	for key, value := range args {
		// Skip parameters that are already in the URL path
		if skipMap[key] {
			continue
		}

		// Skip params object as it's already expanded
		if key == "params" {
			continue
		}

		params.Set(key, fmt.Sprintf("%v", value))
	}

	return params
}

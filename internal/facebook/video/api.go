package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	graphAPIBaseURL      = "https://graph.facebook.com/v23.0"
	graphVideoAPIBaseURL = "https://graph-video.facebook.com/v23.0"
)

// makeGraphAPIRequest makes a request to the Facebook Graph API
func makeGraphAPIRequest(ctx context.Context, method, endpoint string, params map[string]string, body interface{}) ([]byte, error) {
	return makeAPIRequest(ctx, graphAPIBaseURL, method, endpoint, params, body, "")
}

// makeGraphVideoAPIRequest makes a request to the Facebook Graph Video API
func makeGraphVideoAPIRequest(ctx context.Context, method, endpoint string, body io.Reader, contentType string) ([]byte, error) {
	// Get access token
	accessToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("FACEBOOK_ACCESS_TOKEN not set")
	}

	// Build URL
	fullURL := fmt.Sprintf("%s/%s", graphVideoAPIBaseURL, endpoint)
	if !strings.Contains(fullURL, "?") {
		fullURL += "?"
	} else {
		fullURL += "&"
	}
	fullURL += fmt.Sprintf("access_token=%s", url.QueryEscape(accessToken))

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var fbError struct {
			Error FacebookError `json:"error"`
		}
		if err := json.Unmarshal(respBody, &fbError); err == nil && fbError.Error.Message != "" {
			return nil, &fbError.Error
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// makeAPIRequest is a generic API request function
func makeAPIRequest(ctx context.Context, baseURL, method, endpoint string, params map[string]string, body interface{}, contentType string) ([]byte, error) {
	// Get access token
	accessToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("FACEBOOK_ACCESS_TOKEN not set")
	}

	// Build URL with parameters
	fullURL := fmt.Sprintf("%s/%s", baseURL, endpoint)
	urlParams := url.Values{}
	urlParams.Set("access_token", accessToken)
	
	for key, value := range params {
		urlParams.Set(key, value)
	}

	if method == "GET" || (method == "POST" && body == nil) {
		if strings.Contains(fullURL, "?") {
			fullURL += "&" + urlParams.Encode()
		} else {
			fullURL += "?" + urlParams.Encode()
		}
	}

	// Prepare request body
	var reqBody io.Reader
	if body != nil && method != "GET" {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
		if contentType == "" {
			contentType = "application/json"
		}
	} else if method == "POST" && body == nil {
		// For POST with params in URL, we need an empty body
		reqBody = strings.NewReader("")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var fbError struct {
			Error FacebookError `json:"error"`
		}
		if err := json.Unmarshal(respBody, &fbError); err == nil && fbError.Error.Message != "" {
			return nil, &fbError.Error
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
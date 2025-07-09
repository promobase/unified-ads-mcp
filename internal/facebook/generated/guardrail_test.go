package generated

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"unified-ads-mcp/internal/facebook/testutil"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")
}

// TestGuardrailPreventsRealAPI verifies that our guardrail panics when trying to hit real Facebook API
func TestGuardrailPreventsRealAPI(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	tests := []struct {
		name        string
		setupFunc   func()
		cleanupFunc func()
		testFunc    func(*testing.T)
	}{
		{
			name: "Direct_Facebook_URL",
			setupFunc: func() {
				// Temporarily set the URLs to real Facebook
				graphAPIHost = "https://graph.facebook.com"
				baseGraphURL = "https://graph.facebook.com"
			},
			cleanupFunc: func() {
				// Reset to safe defaults
				graphAPIHost = "http://test-mock-server"
				baseGraphURL = "http://test-mock-server"
			},
			testFunc: func(t *testing.T) {
				// This should panic due to our guardrail
				defer func() {
					if r := recover(); r != nil {
						// Expected panic
						panicMsg, ok := r.(string)
						assert := testutil.NewAssert(t)
						assert.True(ok, "Panic should be a string")
						assert.True(strings.Contains(panicMsg, "TEST GUARDRAIL"), "Panic message should contain TEST GUARDRAIL")
						assert.True(strings.Contains(panicMsg, "facebook.com"), "Panic message should mention facebook.com")
						t.Logf("Guardrail successfully prevented real API call: %v", r)
					} else {
						t.Error("Expected panic from guardrail, but no panic occurred")
					}
				}()

				// Try to make a request that would hit real Facebook API
				_, err := makeGraphRequest("GET", "https://graph.facebook.com/v23.0/test", nil)
				if err != nil {
					t.Errorf("Should have panicked before returning error: %v", err)
				}
			},
		},
		{
			name: "Graph_Facebook_Subdomain",
			setupFunc: func() {
				// Set to a different Facebook subdomain
				graphAPIHost = "https://graph-video.facebook.com"
				baseGraphURL = "https://graph-video.facebook.com"
			},
			cleanupFunc: func() {
				graphAPIHost = "http://test-mock-server"
				baseGraphURL = "http://test-mock-server"
			},
			testFunc: func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic
						assert := testutil.NewAssert(t)
						panicMsg, ok := r.(string)
						assert.True(ok, "Panic should be a string")
						assert.True(strings.Contains(panicMsg, "TEST GUARDRAIL"), "Should panic for any facebook.com domain")
					} else {
						t.Error("Expected panic for graph-video.facebook.com")
					}
				}()

				_, _ = makeGraphRequest("GET", "https://graph-video.facebook.com/v23.0/video", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			oldHost := graphAPIHost
			oldBase := baseGraphURL

			// Setup
			tt.setupFunc()

			// Ensure cleanup happens
			defer func() {
				tt.cleanupFunc()
				// Extra safety: restore original values
				graphAPIHost = oldHost
				baseGraphURL = oldBase
			}()

			// Run test
			tt.testFunc(t)
		})
	}
}

// TestGuardrailAllowsMockServers verifies that mock servers work fine
func TestGuardrailAllowsMockServers(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	// Setup a working mock server
	env.Server().SetDefaultHandler(func(w http.ResponseWriter, r *http.Request) {
		testutil.WriteJSONSuccess(w, map[string]interface{}{
			"success": true,
		})
	})

	tests := []struct {
		name     string
		url      string
		testFunc func(*testing.T, string)
	}{
		{
			name: "Localhost_URL",
			url:  "http://localhost:12345/v23.0/test",
			testFunc: func(t *testing.T, url string) {
				// This should NOT panic
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Guardrail should not panic for localhost, but got: %v", r)
					}
				}()

				// This will fail with connection error but shouldn't panic
				_, err := makeGraphRequest("GET", url, nil)
				testutil.NewAssert(t).NotNil(err, "Should get connection error for non-existent server")
			},
		},
		{
			name: "Test_Server_URL",
			url:  env.Server().URL + "/v23.0/test",
			testFunc: func(t *testing.T, url string) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Guardrail should not panic for test server, but got: %v", r)
					}
				}()

				// This should succeed without panic
				resp, err := makeGraphRequest("GET", url, nil)
				assert := testutil.NewAssert(t)
				assert.NoError(err, "Should succeed with test server")
				assert.NotNil(resp, "Should get response from test server")
			},
		},
		{
			name: "Custom_Domain",
			url:  "http://mycompany.internal/api/test",
			testFunc: func(t *testing.T, url string) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Guardrail should not panic for custom domains, but got: %v", r)
					}
				}()

				// This will fail but shouldn't panic
				_, err := makeGraphRequest("GET", url, nil)
				assert := testutil.NewAssert(t)
				assert.NotNil(err, "Should get error for non-existent domain")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t, tt.url)
		})
	}
}

// TestGuardrailEnvironmentVariable tests the TESTING environment variable behavior
func TestGuardrailEnvironmentVariable(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Teardown()

	tests := []struct {
		name        string
		envValue    string
		shouldPanic bool
		setupFunc   func()
		cleanupFunc func()
	}{
		{
			name:        "TESTING_true",
			envValue:    "true",
			shouldPanic: true,
			setupFunc:   func() { os.Setenv("TESTING", "true") },
			cleanupFunc: func() { os.Setenv("TESTING", "true") }, // Reset to true
		},
		{
			name:        "TESTING_false",
			envValue:    "false",
			shouldPanic: false,
			setupFunc:   func() { os.Setenv("TESTING", "false") },
			cleanupFunc: func() { os.Setenv("TESTING", "true") }, // Reset to true
		},
		{
			name:        "TESTING_unset",
			envValue:    "",
			shouldPanic: false,
			setupFunc:   func() { os.Unsetenv("TESTING") },
			cleanupFunc: func() { os.Setenv("TESTING", "true") }, // Reset to true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupFunc()
			defer tt.cleanupFunc()

			// Test with Facebook URL
			didPanic := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic = true
					}
				}()

				_, _ = makeGraphRequest("GET", "https://graph.facebook.com/test", nil)
			}()

			assert := testutil.NewAssert(t)
			assert.Equal(didPanic, tt.shouldPanic, "Panic behavior for TESTING="+tt.envValue)
		})
	}
}

package generated

import (
	"os"
	"strings"
	"testing"
)

func init() {
	// Set testing environment variable to enable guardrails
	os.Setenv("TESTING", "true")
}

// TestGuardrailPreventsRealAPI verifies that our guardrail panics when trying to hit real Facebook API
func TestGuardrailPreventsRealAPI(t *testing.T) {
	// Temporarily set the URLs to real Facebook
	oldHost := graphAPIHost
	oldBase := baseGraphURL
	graphAPIHost = "https://graph.facebook.com"
	baseGraphURL = "https://graph.facebook.com"
	defer func() {
		graphAPIHost = oldHost
		baseGraphURL = oldBase
	}()

	// This should panic due to our guardrail
	defer func() {
		if r := recover(); r != nil {
			// Expected panic
			panicMsg := r.(string)
			if !strings.Contains(panicMsg, "TEST GUARDRAIL") {
				t.Errorf("Expected guardrail panic, got: %v", r)
			}
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
}

// TestGuardrailAllowsMockServers verifies that mock servers work fine
func TestGuardrailAllowsMockServers(t *testing.T) {
	// Use a mock server URL
	mockURL := "http://localhost:12345/v23.0/test"

	// This should NOT panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Guardrail should not panic for mock servers, but got: %v", r)
		}
	}()

	// Set up a simple test - this will fail with connection error but shouldn't panic
	_, err := makeGraphRequest("GET", mockURL, nil)
	if err == nil {
		t.Error("Expected connection error for non-existent mock server")
	}
	// The important thing is it didn't panic
}

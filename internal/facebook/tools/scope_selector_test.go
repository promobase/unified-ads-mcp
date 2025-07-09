package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopeSelectorHandler(t *testing.T) {
	// Create a test server and register the scope selector tool
	testServer := server.NewMCPServer(
		"test-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	err := RegisterScopeSelectorTool(testServer)
	require.NoError(t, err)

	tests := []struct {
		name           string
		args           ScopeSelectorArgs
		expectedError  bool
		expectedResult string
	}{
		{
			name: "Get initial scopes",
			args: ScopeSelectorArgs{
				Action: "get_scopes",
			},
			expectedError:  false,
			expectedResult: "Active domains: adaccount",
		},
		{
			name: "Add campaign scope",
			args: ScopeSelectorArgs{
				Action:  "add_scopes",
				Domains: []string{"campaign"},
			},
			expectedError:  false,
			expectedResult: "Active domains: adaccount, campaign",
		},
		{
			name: "Set scopes to ad and adset",
			args: ScopeSelectorArgs{
				Action:  "set_scopes",
				Domains: []string{"ad", "adset"},
			},
			expectedError:  false,
			expectedResult: "Active domains: ad, adset",
		},
		{
			name: "Remove ad scope",
			args: ScopeSelectorArgs{
				Action:  "remove_scopes",
				Domains: []string{"ad"},
			},
			expectedError:  false,
			expectedResult: "Active domains: adset",
		},
		{
			name: "Invalid action",
			args: ScopeSelectorArgs{
				Action: "invalid_action",
			},
			expectedError: true,
		},
		{
			name: "Set scopes without domains",
			args: ScopeSelectorArgs{
				Action: "set_scopes",
			},
			expectedError: true,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "scope_selector",
					Arguments: map[string]interface{}{
						"action":  tt.args.Action,
						"domains": tt.args.Domains,
					},
				},
			}

			// Call the handler
			result, err := ScopeSelectorHandler(ctx, request, tt.args)
			require.NoError(t, err)

			if tt.expectedError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
				if tt.expectedResult != "" {
					// Check the first content item contains expected text
					assert.NotEmpty(t, result.Content)
					textContent, ok := mcp.AsTextContent(result.Content[0])
					assert.True(t, ok)
					assert.Contains(t, textContent.Text, tt.expectedResult)
				}
			}
		})
	}
}

func TestScopeManager(t *testing.T) {
	// Create a test server
	testServer := server.NewMCPServer(
		"test-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Create a scope manager
	sm := NewScopeManager(testServer)

	t.Run("Initial state", func(t *testing.T) {
		domains := sm.GetActiveDomains()
		assert.Empty(t, domains)
	})

	t.Run("Add domains", func(t *testing.T) {
		err := sm.AddDomains([]string{"adaccount", "campaign"})
		assert.NoError(t, err)

		domains := sm.GetActiveDomains()
		assert.Equal(t, []string{"adaccount", "campaign"}, domains)
	})

	t.Run("Set domains", func(t *testing.T) {
		err := sm.SetDomains([]string{"ad", "adset"})
		assert.NoError(t, err)

		domains := sm.GetActiveDomains()
		assert.Equal(t, []string{"ad", "adset"}, domains)
	})

	t.Run("Remove domains", func(t *testing.T) {
		err := sm.RemoveDomains([]string{"ad"})
		assert.NoError(t, err)

		domains := sm.GetActiveDomains()
		assert.Equal(t, []string{"adset"}, domains)
	})

	t.Run("Invalid domain", func(t *testing.T) {
		err := sm.AddDomains([]string{"invalid_domain"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown domain")
	})
}

func TestFormatScopeResult(t *testing.T) {
	tests := []struct {
		name     string
		result   map[string]interface{}
		expected string
	}{
		{
			name: "Get scopes result",
			result: map[string]interface{}{
				"active_domains":    []string{"adaccount", "campaign"},
				"available_domains": []string{"ad", "adaccount", "adcreative", "adset", "campaign", "user"},
			},
			expected: "Active domains: adaccount, campaign\nAvailable domains: ad, adaccount, adcreative, adset, campaign, user\n",
		},
		{
			name: "Action result",
			result: map[string]interface{}{
				"action":         "add_scopes",
				"active_domains": []string{"adaccount", "campaign"},
			},
			expected: "Action: add_scopes\nActive domains: adaccount, campaign\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatScopeResult(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

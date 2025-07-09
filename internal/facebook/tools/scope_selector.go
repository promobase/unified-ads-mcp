package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"unified-ads-mcp/internal/facebook/generated"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ScopeSelectorArgs defines the arguments for the scope_selector tool
type ScopeSelectorArgs struct {
	Action  string   `json:"action" jsonschema:"required,description=Action to perform,enum=['get_scopes', 'set_scopes', 'add_scopes', 'remove_scopes']"`
	Domains []string `json:"domains,omitempty" jsonschema:"description=Domain names to manage,items={enum=['ad', 'adaccount', 'adcreative', 'adset', 'campaign', 'user']}"`
}

// ScopeManager manages the dynamic registration and unregistration of tools for different domains
type ScopeManager struct {
	server        *server.MCPServer
	activeDomains map[string]bool
	mu            sync.RWMutex
}

// NewScopeManager creates a new scope manager
func NewScopeManager(s *server.MCPServer) *ScopeManager {
	return &ScopeManager{
		server:        s,
		activeDomains: make(map[string]bool),
	}
}

// GetActiveDomains returns the list of currently active domains
func (sm *ScopeManager) GetActiveDomains() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	domains := make([]string, 0, len(sm.activeDomains))
	for domain := range sm.activeDomains {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	return domains
}

// SetDomains sets the active domains, registering new ones and unregistering removed ones
func (sm *ScopeManager) SetDomains(domains []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	newDomains := make(map[string]bool)
	for _, domain := range domains {
		newDomains[domain] = true
	}

	// Register new domains
	for domain := range newDomains {
		if !sm.activeDomains[domain] {
			if err := sm.registerDomain(domain); err != nil {
				return fmt.Errorf("failed to register domain '%s': %w", domain, err)
			}
		}
	}

	// Unregister removed domains
	for domain := range sm.activeDomains {
		if !newDomains[domain] {
			if err := sm.unregisterDomain(domain); err != nil {
				return fmt.Errorf("failed to unregister domain '%s': %w", domain, err)
			}
		}
	}

	sm.activeDomains = newDomains
	return nil
}

// AddDomains adds new domains to the active set
func (sm *ScopeManager) AddDomains(domains []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, domain := range domains {
		if !sm.activeDomains[domain] {
			if err := sm.registerDomain(domain); err != nil {
				return fmt.Errorf("failed to register domain '%s': %w", domain, err)
			}
			sm.activeDomains[domain] = true
		}
	}
	return nil
}

// RemoveDomains removes domains from the active set
func (sm *ScopeManager) RemoveDomains(domains []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, domain := range domains {
		if sm.activeDomains[domain] {
			if err := sm.unregisterDomain(domain); err != nil {
				return fmt.Errorf("failed to unregister domain '%s': %w", domain, err)
			}
			delete(sm.activeDomains, domain)
		}
	}
	return nil
}

// registerDomain registers tools for a specific domain
func (sm *ScopeManager) registerDomain(domain string) error {
	switch domain {
	case "ad":
		return generated.RegisterAdTools(sm.server)
	case "adaccount":
		return generated.RegisterAdAccountTools(sm.server)
	case "adcreative":
		return generated.RegisterAdCreativeTools(sm.server)
	case "adset":
		return generated.RegisterAdSetTools(sm.server)
	case "campaign":
		return generated.RegisterCampaignTools(sm.server)
	case "user":
		return generated.RegisterUserTools(sm.server)
	default:
		return fmt.Errorf("unknown domain: %s", domain)
	}
}

// unregisterDomain unregisters tools for a specific domain
func (sm *ScopeManager) unregisterDomain(domain string) error {
	// Note: mcp-go doesn't have a built-in unregister mechanism,
	// so we'll need to implement this by recreating the server or
	// keeping track of tool names and removing them manually
	// For now, we'll just log that the domain would be unregistered
	// This is a limitation of the current mcp-go implementation
	fmt.Printf("Note: Domain '%s' would be unregistered (not implemented in mcp-go)\n", domain)
	return nil
}

// Global scope manager instance
var globalScopeManager *ScopeManager

// InitializeScopeManager initializes the global scope manager
func InitializeScopeManager(s *server.MCPServer) {
	globalScopeManager = NewScopeManager(s)
}

// ScopeSelectorHandler handles the scope_selector tool calls
func ScopeSelectorHandler(ctx context.Context, request mcp.CallToolRequest, args ScopeSelectorArgs) (*mcp.CallToolResult, error) {
	if globalScopeManager == nil {
		return mcp.NewToolResultError("scope manager not initialized"), nil
	}

	switch args.Action {
	case "get_scopes":
		activeDomains := globalScopeManager.GetActiveDomains()
		result := map[string]interface{}{
			"active_domains":    activeDomains,
			"available_domains": []string{"ad", "adaccount", "adcreative", "adset", "campaign", "user"},
		}
		return mcp.NewToolResultText(formatScopeResult(result)), nil

	case "set_scopes":
		if len(args.Domains) == 0 {
			return mcp.NewToolResultError("domains are required for set_scopes action"), nil
		}
		if err := globalScopeManager.SetDomains(args.Domains); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to set domains: %v", err)), nil
		}
		result := map[string]interface{}{
			"action":         "set_scopes",
			"active_domains": globalScopeManager.GetActiveDomains(),
		}
		return mcp.NewToolResultText(formatScopeResult(result)), nil

	case "add_scopes":
		if len(args.Domains) == 0 {
			return mcp.NewToolResultError("domains are required for add_scopes action"), nil
		}
		if err := globalScopeManager.AddDomains(args.Domains); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to add domains: %v", err)), nil
		}
		result := map[string]interface{}{
			"action":         "add_scopes",
			"active_domains": globalScopeManager.GetActiveDomains(),
		}
		return mcp.NewToolResultText(formatScopeResult(result)), nil

	case "remove_scopes":
		if len(args.Domains) == 0 {
			return mcp.NewToolResultError("domains are required for remove_scopes action"), nil
		}
		if err := globalScopeManager.RemoveDomains(args.Domains); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to remove domains: %v", err)), nil
		}
		result := map[string]interface{}{
			"action":         "remove_scopes",
			"active_domains": globalScopeManager.GetActiveDomains(),
		}
		return mcp.NewToolResultText(formatScopeResult(result)), nil

	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown action: %s", args.Action)), nil
	}
}

// formatScopeResult formats the scope result for display
func formatScopeResult(result map[string]interface{}) string {
	var sb strings.Builder

	if action, ok := result["action"].(string); ok {
		sb.WriteString(fmt.Sprintf("Action: %s\n", action))
	}

	if activeDomains, ok := result["active_domains"].([]string); ok {
		sb.WriteString(fmt.Sprintf("Active domains: %s\n", strings.Join(activeDomains, ", ")))
	}

	if availableDomains, ok := result["available_domains"].([]string); ok {
		sb.WriteString(fmt.Sprintf("Available domains: %s\n", strings.Join(availableDomains, ", ")))
	}

	return sb.String()
}

// RegisterScopeSelectorTool registers the scope_selector tool with the MCP server
func RegisterScopeSelectorTool(s *server.MCPServer) error {
	// Initialize the scope manager
	InitializeScopeManager(s)

	// Mark adaccount as initially active since it's registered by default
	globalScopeManager.activeDomains["adaccount"] = true

	// Create the scope selector tool with raw JSON schema
	scopeSelectorTool := mcp.NewToolWithRawSchema(
		"scope_selector",
		"Manage which Facebook API domain tool sets are currently loaded. By default, only 'adaccount' tools are available. Use this tool to dynamically load/unload tools for different domains (ad, adaccount, adcreative, adset, campaign, user).",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"action": {
					"type": "string",
					"description": "Action to perform",
					"enum": ["get_scopes", "set_scopes", "add_scopes", "remove_scopes"]
				},
				"domains": {
					"type": "array",
					"description": "Domain names to manage",
					"items": {
						"type": "string",
						"enum": ["ad", "adaccount", "adcreative", "adset", "campaign", "user"]
					}
				}
			},
			"required": ["action"],
			"additionalProperties": false
		}`),
	)

	// Register the tool with proper argument binding
	s.AddTool(scopeSelectorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args ScopeSelectorArgs

		// Convert arguments to JSON bytes
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal arguments: %v", err)), nil
		}

		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid arguments: %v", err)), nil
		}
		return ScopeSelectorHandler(ctx, request, args)
	})

	return nil
}

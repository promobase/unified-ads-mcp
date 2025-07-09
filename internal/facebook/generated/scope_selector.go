package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

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
		return RegisterAdTools(sm.server)
	case "adaccount":
		return RegisterAdAccountTools(sm.server)
	case "adcreative":
		return RegisterAdCreativeTools(sm.server)
	case "adset":
		return RegisterAdSetTools(sm.server)
	case "campaign":
		return RegisterCampaignTools(sm.server)
	case "user":
		return RegisterUserTools(sm.server)
	default:
		return fmt.Errorf("unknown domain: %s", domain)
	}
}

// getDomainToolNames returns the list of tool names for a given domain
func getDomainToolNames(domain string) []string {
	switch domain {
	case "ad":
		return []string{
			"ad_get", "ad_update", "ad_delete", "ad_list_adlabels", "ad_create_adlabel",
			"ad_get_insights", "ad_create_insights_report", "ad_list_previews", "ad_create_preview",
		}
	case "adaccount":
		return []string{
			"ad_account_get", "ad_account_update", "ad_account_list_activities", "ad_account_list_ad_creatives",
			"ad_account_create_ad_creative", "ad_account_list_ads", "ad_account_create_ad", "ad_account_list_adsets",
			"ad_account_create_adset", "ad_account_list_campaigns", "ad_account_create_campaign",
			"ad_account_list_custom_audiences", "ad_account_create_custom_audience",
		}
	case "adcreative":
		return []string{
			"ad_creative_get", "ad_creative_update", "ad_creative_delete", "ad_creative_list_previews",
			"ad_creative_create_preview",
		}
	case "adset":
		return []string{
			"ad_set_get", "ad_set_update", "ad_set_delete", "ad_set_list_ads", "ad_set_create_ad",
			"ad_set_list_adlabels", "ad_set_create_adlabel", "ad_set_get_insights", "ad_set_create_insights_report",
		}
	case "campaign":
		return []string{
			"campaign_get", "campaign_update", "campaign_delete", "campaign_list_ads", "campaign_list_adsets",
			"campaign_list_ad_studies", "campaign_create_adlabel", "campaign_get_adrules_governed",
			"campaign_create_budget_schedule", "campaign_list_copies", "campaign_create_copie",
			"campaign_get_insights", "campaign_create_insights_report",
		}
	case "user":
		return []string{
			"user_get", "user_list_ad_accounts", "user_list_pages",
		}
	default:
		return []string{}
	}
}

// unregisterDomain unregisters tools for a specific domain
func (sm *ScopeManager) unregisterDomain(domain string) error {
	// Get the list of tool names for this domain
	toolNames := getDomainToolNames(domain)
	if len(toolNames) == 0 {
		return fmt.Errorf("no tools found for domain: %s", domain)
	}

	// Remove tools from the server using the DeleteTools method
	// This will also send a tools list changed notification to clients
	sm.server.DeleteTools(toolNames...)

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
			"action":          "remove_scopes",
			"removed_domains": args.Domains,
			"active_domains":  globalScopeManager.GetActiveDomains(),
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

	if removedDomains, ok := result["removed_domains"].([]string); ok {
		sb.WriteString(fmt.Sprintf("Removed domains: %s\n", strings.Join(removedDomains, ", ")))
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

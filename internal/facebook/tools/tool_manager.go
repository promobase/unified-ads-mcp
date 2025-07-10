package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"unified-ads-mcp/internal/facebook/generated/ad"
	"unified-ads-mcp/internal/facebook/generated/adaccount"
	"unified-ads-mcp/internal/facebook/generated/adcreative"
	"unified-ads-mcp/internal/facebook/generated/adset"
	"unified-ads-mcp/internal/facebook/generated/campaign"
	"unified-ads-mcp/internal/facebook/generated/customaudience"
	"unified-ads-mcp/internal/facebook/generated/user"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolManager manages the dynamic loading and unloading of tools
type ToolManager struct {
	server       *server.MCPServer
	loadedScopes map[string]bool
	mu           sync.RWMutex
}

// Global tool manager instance
var toolManager *ToolManager

// InitToolManager initializes the global tool manager
func InitToolManager(s *server.MCPServer) {
	toolManager = &ToolManager{
		server:       s,
		loadedScopes: make(map[string]bool),
	}
}

// Available scopes/domains
var availableScopes = []string{
	// Generated scopes (all tools from a single object)
	"ad",
	"adaccount",
	"adcreative",
	"adset",
	"campaign",
	"customaudience",
	"user",
	// Custom scopes (curated tool sets)
	"essentials",
	"campaign_management",
	"reporting",
	"audience",
	"creative",
	"optimization",
	"video",
}

const (
	ActionSet = "set"
	ActionGet = "get"
)

var ToolManagerActions = []string{ActionGet, ActionSet}

// RegisterToolManagerTool registers the tool manager tool
func RegisterToolManagerTool(s *server.MCPServer) error {
	// Initialize the tool manager
	InitToolManager(s)

	tool := mcp.NewTool(
		"tool_manager",
		mcp.WithDescription("Manage which Facebook API tool sets are loaded. This allows efficient memory usage by only loading the tools you need. RECOMMENDED: Use custom scopes (essentials, campaign_management, reporting, etc.) for common workflows. Codegen scopes load ALL tools for an object type and should only be used when you need comprehensive access. "),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("Action to perform"),
			mcp.Enum(ToolManagerActions...),
		),
		mcp.WithArray("scopes",
			mcp.Description("Tool scopes to add or remove. PREFER CUSTOM SCOPES: 'essentials', 'campaign_management', 'reporting', 'audience', 'creative', 'optimization'. Codegen scopes ('ad', 'adaccount', etc.) load ALL tools and should be used sparingly."),
			mcp.Items(map[string]interface{}{
				"type": "string",
				"enum": availableScopes,
			}),
		),
	)

	s.AddTool(tool, handleToolManager)
	return nil
}

func handleToolManager(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if toolManager == nil {
		return mcp.NewToolResultError("Tool manager not initialized"), nil
	}

	var args struct {
		Action string   `json:"action"`
		Scopes []string `json:"scopes,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return mcp.NewToolResultErrorf("Invalid arguments: %v", err), nil
	}

	switch args.Action {
	case ActionGet:
		return handleListAvailableScopes()
	case ActionSet:
		return handleSetScopes(args.Scopes)
	default:
		return mcp.NewToolResultErrorf("Invalid action: %s. Valid actions are: %s", args.Action, strings.Join(ToolManagerActions, ", ")), nil
	}
}

func handleListAvailableScopes() (*mcp.CallToolResult, error) {
	toolManager.mu.RLock()
	defer toolManager.mu.RUnlock()

	loaded := make([]string, 0, len(toolManager.loadedScopes))
	for scope := range toolManager.loadedScopes {
		loaded = append(loaded, scope)
	}

	// Combine generated and custom scope descriptions
	descriptions := map[string]string{
		// Generated scopes (low-level, comprehensive)
		"ad":             "[LOW-LEVEL] Ad object - loads ALL 13 tools. Use 'essentials' or 'campaign_management' instead",
		"adaccount":      "[LOW-LEVEL] Ad account object - loads ALL 120+ tools! Use custom scopes instead",
		"adcreative":     "[LOW-LEVEL] Ad creative object - loads ALL 6 tools. Use 'creative' scope instead",
		"adset":          "[LOW-LEVEL] Ad set object - loads ALL 18 tools. Use 'essentials' or 'campaign_management' instead",
		"campaign":       "[LOW-LEVEL] Campaign object - loads ALL 13 tools. Use 'campaign_management' instead",
		"customaudience": "[LOW-LEVEL] Custom audience object - loads ALL 15 tools. Use 'audience' scope instead",
		"user":           "[LOW-LEVEL] User object - loads ALL user management tools",
	}

	// Add custom scope descriptions
	for name, desc := range getCustomScopeDescriptions() {
		descriptions[name] = desc
	}

	// Separate scopes by type for clarity
	generatedScopes := []string{"ad", "adaccount", "adcreative", "adset", "campaign", "customaudience", "user"}
	customScopesList := getCustomScopeNames()

	result := map[string]interface{}{
		"recommendation": "Use custom scopes for most workflows. They provide curated tool sets optimized for specific tasks.",

		"custom_scopes":         customScopesList,
		"codegen_scopes":        generatedScopes,
		"loaded_scopes":         loaded,
		"descriptions":          descriptions,
		"custom_scope_benefits": "Memory efficient, task-focused, easier to use",
		"codegen_scope_warning": "Low-level, loads ALL tools (can be 120+ tools!), use only when needed",
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func handleSetScopes(scopes []string) (*mcp.CallToolResult, error) {
	if len(scopes) == 0 {
		return mcp.NewToolResultError("No scopes provided to add"), nil
	}
	// lock
	toolManager.mu.Lock()
	defer toolManager.mu.Unlock()
	// reset server tools
	toolManager.server.SetTools()
	toolManager.loadedScopes = make(map[string]bool)

	added := []string{}
	errors := []string{}
	codegenWarnings := []string{}

	for _, scope := range scopes {
		scope = strings.ToLower(strings.TrimSpace(scope))

		// Check if valid scope
		if !isValidScope(scope) {
			errors = append(errors, fmt.Sprintf("Invalid scope: %s", scope))
			continue
		}

		// Load the scope
		err := loadScope(scope, toolManager.server)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to load %s: %v", scope, err))
			continue
		}

		toolManager.loadedScopes[scope] = true
		added = append(added, scope)
	}

	result := map[string]interface{}{
		"added":        added,
		"errors":       errors,
		"warnings":     codegenWarnings,
		"total_loaded": len(toolManager.loadedScopes),
	}

	// Add recommendation if codegen scopes were loaded
	if len(codegenWarnings) > 0 {
		result["recommendation"] = "TIP: Use 'tool_manager action=list' to see available custom scopes optimized for specific workflows"
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func isValidScope(scope string) bool {
	// Check both generated and custom scopes
	for _, valid := range availableScopes {
		if scope == valid {
			return true
		}
	}
	return false
}

func loadScope(scope string, s *server.MCPServer) error {
	// Check if it's a custom scope first
	if isCustomScope(scope) {
		return loadCustomScope(scope, s)
	}

	// Otherwise, load generated scope
	switch scope {
	case "ad":
		return ad.RegisterAllAdTools(s)
	case "adaccount":
		return adaccount.RegisterAllAdAccountTools(s)
	case "adcreative":
		return adcreative.RegisterAllAdCreativeTools(s)
	case "adset":
		return adset.RegisterAllAdSetTools(s)
	case "campaign":
		return campaign.RegisterAllCampaignTools(s)
	case "customaudience":
		return customaudience.RegisterAllCustomAudienceTools(s)
	case "user":
		return user.RegisterAllUserTools(s)
	default:
		return fmt.Errorf("unknown scope: %s", scope)
	}
}

// GetLoadedScopes returns the currently loaded scopes
func GetLoadedScopes() []string {
	if toolManager == nil {
		return []string{}
	}

	toolManager.mu.RLock()
	defer toolManager.mu.RUnlock()

	scopes := make([]string, 0, len(toolManager.loadedScopes))
	for scope := range toolManager.loadedScopes {
		scopes = append(scopes, scope)
	}
	return scopes
}

// IsLoaded checks if a scope is loaded
func IsLoaded(scope string) bool {
	if toolManager == nil {
		return false
	}

	toolManager.mu.RLock()
	defer toolManager.mu.RUnlock()

	return toolManager.loadedScopes[scope]
}

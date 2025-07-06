package mcp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"unified-ads-mcp/internal/config"
	"unified-ads-mcp/internal/facebook"
	"unified-ads-mcp/internal/shared"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ImprovedMCPServer wraps the mcp-go server with context support
type ImprovedMCPServer struct {
	server         *server.MCPServer
	config         *config.Config
	facebookToken  string
	enabledObjects map[string]bool
}

// ObjectCategories defines categories of Facebook objects
var ObjectCategories = map[string][]string{
	"core_ads": {
		"Ad", "AdAccount", "AdSet", "Campaign", "AdCreative",
	},
	"targeting": {
		"AdSavedLocation", "AdSavedKeywords", "SavedAudience", "CustomAudience",
	},
	"reporting": {
		"AdReportRun", "AdsInsights", "AdStudy", "AdSavedReport",
	},
	"creative_assets": {
		"AdImage", "AdVideo", "AdCreativeAssetGroup", "AdCreativeTemplate",
	},
	"business": {
		"Business", "BusinessAssetGroup", "BusinessProject", "BusinessUser",
	},
	"page": {
		"Page", "PagePost", "PageInsights", "PageVideos",
	},
}

// NewImprovedMCPServer creates a new MCP server with improved context handling
func NewImprovedMCPServer(cfg *config.Config, enabledCategories []string) (*ImprovedMCPServer, error) {
	// Create enabled objects map
	enabledObjects := make(map[string]bool)
	if len(enabledCategories) == 0 {
		// If no categories specified, enable core_ads by default
		enabledCategories = []string{"core_ads"}
	}

	// Build enabled objects map from categories
	for _, category := range enabledCategories {
		if objects, ok := ObjectCategories[category]; ok {
			for _, obj := range objects {
				enabledObjects[strings.ToLower(obj)] = true
			}
		} else if category == "all" {
			// Enable all objects
			for _, objects := range ObjectCategories {
				for _, obj := range objects {
					enabledObjects[strings.ToLower(obj)] = true
				}
			}
		}
	}

	// Create the mcp-go server with session support
	mcpServer := server.NewMCPServer(
		"unified-ads-mcp",
		"2.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	s := &ImprovedMCPServer{
		server:         mcpServer,
		config:         cfg,
		facebookToken:  cfg.Facebook.AccessToken,
		enabledObjects: enabledObjects,
	}

	// Log initialization
	log.Printf("Initializing MCP server with enabled categories: %v", enabledCategories)

	// Register Facebook tools if token is available
	if cfg.Facebook.AccessToken != "" {
		if err := s.registerFacebookTools(mcpServer); err != nil {
			return nil, fmt.Errorf("failed to register Facebook tools: %w", err)
		}
		log.Printf("Facebook Business API tools registered for objects: %v", s.getEnabledObjectsList())
	}

	return s, nil
}

// registerFacebookTools registers Facebook tools with context-aware handlers
func (s *ImprovedMCPServer) registerFacebookTools(mcpServer *server.MCPServer) error {
	// Get all tools but filter based on enabled objects
	allTools := facebook.GetFilteredMCPTools(s.enabledObjects)
	handlers := facebook.GetContextAwareHandlers(s.facebookToken)

	// Register each tool with its context-aware handler
	for _, tool := range allTools {
		if handler, ok := handlers[tool.Name]; ok {
			// Wrap handler to inject context
			wrappedHandler := s.wrapHandlerWithContext(handler)
			mcpServer.AddTool(tool, wrappedHandler)
		}
	}

	return nil
}

// wrapHandlerWithContext wraps a handler to inject access token into context
func (s *ImprovedMCPServer) wrapHandlerWithContext(handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Add access token to context
		ctx = shared.WithFacebookAccessToken(ctx, s.facebookToken)
		// Add enabled objects to context
		ctx = shared.WithEnabledObjectTypes(ctx, s.enabledObjects)

		// Call the original handler
		return handler(ctx, request)
	}
}

// getEnabledObjectsList returns a list of enabled object names
func (s *ImprovedMCPServer) getEnabledObjectsList() []string {
	var enabled []string
	for obj, isEnabled := range s.enabledObjects {
		if isEnabled {
			enabled = append(enabled, obj)
		}
	}
	return enabled
}

// Start runs the MCP server
func (s *ImprovedMCPServer) Start() error {
	log.Println("Starting improved MCP server with context support...")
	return server.ServeStdio(s.server)
}

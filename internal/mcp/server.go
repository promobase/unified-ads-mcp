package mcp

import (
	"context"
	"fmt"
	"log"

	"unified-ads-mcp/internal/config"
	"unified-ads-mcp/internal/facebook"
	"unified-ads-mcp/internal/shared"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ObjectCategory represents a category of Facebook objects
type ObjectCategory string

const (
	// CategoryCoreAds is the default category for core advertising objects
	CategoryCoreAds ObjectCategory = "core_ads"
	// CategoryTargeting includes targeting-related objects
	CategoryTargeting ObjectCategory = "targeting"
	// CategoryReporting includes reporting and insights objects
	CategoryReporting ObjectCategory = "reporting"
	// CategoryCreativeAssets includes creative assets like images and videos
	CategoryCreativeAssets ObjectCategory = "creative_assets"
	// CategoryBusiness includes business management objects
	CategoryBusiness ObjectCategory = "business"
	// CategoryPage includes page-related objects
	CategoryPage ObjectCategory = "page"
	// CategoryAll is a special category that includes all objects
	CategoryAll ObjectCategory = "all"
)

// WrappedMCPServer wraps the mcp-go server with context support
type WrappedMCPServer struct {
	server         *server.MCPServer
	config         *config.Config
	facebookToken  string
	enabledObjects map[string]bool
}

// ObjectCategories defines categories of Facebook objects
var ObjectCategories = map[ObjectCategory][]string{
	CategoryCoreAds: {
		"Ad", "AdAccount", "AdSet", "Campaign", "AdCreative",
	},
	CategoryTargeting: {
		"AdSavedLocation", "AdSavedKeywords", "SavedAudience", "CustomAudience",
	},
	CategoryReporting: {
		"AdReportRun", "AdsInsights", "AdStudy", "AdSavedReport",
	},
	CategoryCreativeAssets: {
		"AdImage", "AdVideo", "AdCreativeAssetGroup", "AdCreativeTemplate",
	},
	CategoryBusiness: {
		"Business", "BusinessAssetGroup", "BusinessProject", "BusinessUser",
	},
	CategoryPage: {
		"Page", "PagePost", "PageInsights", "PageVideos",
	},
}

// GetValidCategories returns a list of all valid category names
func GetValidCategories() []string {
	return []string{
		string(CategoryCoreAds),
		string(CategoryTargeting),
		string(CategoryReporting),
		string(CategoryCreativeAssets),
		string(CategoryBusiness),
		string(CategoryPage),
		string(CategoryAll),
	}
}

// InitMCPServer creates a new MCP server with context handling
func InitMCPServer(cfg *config.Config, enabledCategories []string) (*WrappedMCPServer, error) {
	// Create enabled objects map
	enabledObjects := make(map[string]bool)
	if len(enabledCategories) == 0 {
		// If no categories specified, enable core_ads by default
		enabledCategories = []string{string(CategoryCoreAds)}
	}

	// Build enabled objects map from categories
	for _, categoryStr := range enabledCategories {
		category := ObjectCategory(categoryStr)
		
		if category == CategoryAll {
			// Enable all objects
			for _, objects := range ObjectCategories {
				for _, obj := range objects {
					enabledObjects[obj] = true
				}
			}
		} else if objects, ok := ObjectCategories[category]; ok {
			for _, obj := range objects {
				enabledObjects[obj] = true
			}
		} else {
			return nil, fmt.Errorf("unsupported category: '%s'. Valid categories are: %v", categoryStr, GetValidCategories())
		}
	}

	// Create the mcp-go server with session support
	mcpServer := server.NewMCPServer(
		"unified-ads-mcp",
		"0.0.1",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	s := &WrappedMCPServer{
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
func (s *WrappedMCPServer) registerFacebookTools(mcpServer *server.MCPServer) error {
	// Get all tools but filter based on enabled objects
	allTools := facebook.GetFilteredMCPTools(s.enabledObjects)
	handlers := facebook.GetContextAwareHandlers(s.facebookToken)

	log.Printf("Found %d filtered tools", len(allTools))
	log.Printf("Found %d handlers", len(handlers))
	log.Printf("Enabled objects: %v", s.enabledObjects)

	// Count registered tools
	registeredCount := 0

	// Register each tool with its context-aware handler
	for _, tool := range allTools {
		if handler, ok := handlers[tool.Name]; ok {
			// Wrap handler to inject context
			wrappedHandler := s.wrapHandlerWithContext(handler)
			mcpServer.AddTool(tool, wrappedHandler)
			registeredCount++
		} else {
			// Log first few missing handlers for debugging
			if registeredCount == 0 && len(allTools)-registeredCount < 5 {
				log.Printf("No handler found for tool: %s", tool.Name)
			}
		}
	}

	log.Printf("Registered %d tools out of %d filtered tools", registeredCount, len(allTools))
	return nil
}

// wrapHandlerWithContext wraps a handler to inject access token into context
func (s *WrappedMCPServer) wrapHandlerWithContext(handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
func (s *WrappedMCPServer) getEnabledObjectsList() []string {
	var enabled []string
	for obj, isEnabled := range s.enabledObjects {
		if isEnabled {
			enabled = append(enabled, obj)
		}
	}
	return enabled
}

// Start runs the MCP server
func (s *WrappedMCPServer) Start() error {
	log.Println("Starting MCP server with context support...")
	return server.ServeStdio(s.server)
}

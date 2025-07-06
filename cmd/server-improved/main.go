package main

import (
	"log"
	"os"
	"strings"

	"unified-ads-mcp/internal/config"
	"unified-ads-mcp/internal/mcp"
)

func main() {
	// Load configuration
	cfg := &config.Config{
		Facebook: &config.FacebookConfig{
			AccessToken: os.Getenv("FACEBOOK_ACCESS_TOKEN"),
		},
	}

	// Parse enabled categories from environment variable
	// Example: ENABLED_CATEGORIES="core_ads,reporting" or ENABLED_CATEGORIES="all"
	enabledCategories := []string{"core_ads"} // default
	if categories := os.Getenv("ENABLED_CATEGORIES"); categories != "" {
		enabledCategories = strings.Split(categories, ",")
		for i := range enabledCategories {
			enabledCategories[i] = strings.TrimSpace(enabledCategories[i])
		}
	}

	// Create improved MCP server with context support
	server, err := mcp.NewImprovedMCPServer(cfg, enabledCategories)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"unified-ads-mcp/internal/config"
	"unified-ads-mcp/internal/mcp"
	"unified-ads-mcp/internal/utils"
)

func main() {
	// Load .env file
	utils.LoadEnv()

	// Define command-line flags
	var (
		categories = flag.String("categories", "core_ads", 
			fmt.Sprintf("Comma-separated list of enabled categories. Valid options: %s", 
				strings.Join(mcp.GetValidCategories(), ", ")))
		help = flag.Bool("help", false, "Show help message")
	)

	// Parse command-line arguments
	flag.Parse()

	// Show help if requested
	if *help {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Unified Ads MCP Server - A Model Context Protocol server for ad platforms\n\n")
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nENVIRONMENT VARIABLES:\n")
		fmt.Fprintf(os.Stderr, "  FACEBOOK_ACCESS_TOKEN    Facebook access token (required)\n")
		fmt.Fprintf(os.Stderr, "  ENABLED_CATEGORIES       Alternative to --categories flag\n")
		fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
		fmt.Fprintf(os.Stderr, "  %s --categories=core_ads,reporting\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --categories=all\n", os.Args[0])
		os.Exit(0)
	}

	// Load configuration
	cfg := &config.Config{
		Facebook: &config.FacebookConfig{
			AccessToken: os.Getenv("FACEBOOK_ACCESS_TOKEN"),
		},
	}

	// Parse enabled categories
	// Priority: CLI flag > Environment variable > Default
	var enabledCategories []string
	if *categories != "" {
		enabledCategories = strings.Split(*categories, ",")
	} else if envCategories := os.Getenv("ENABLED_CATEGORIES"); envCategories != "" {
		enabledCategories = strings.Split(envCategories, ",")
	} else {
		enabledCategories = []string{"core_ads"}
	}

	// Trim whitespace from categories
	for i := range enabledCategories {
		enabledCategories[i] = strings.TrimSpace(enabledCategories[i])
	}

	log.Printf("Starting server with categories: %v", enabledCategories)

	// Create MCP server with context support
	server, err := mcp.InitMCPServer(cfg, enabledCategories)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

package main

import (
	"log"
	"os"

	"unified-ads-mcp/internal/facebook"
	"unified-ads-mcp/internal/utils"
)

func main() {
	// Load .env file
	utils.LoadEnv()

	// Check if access token is provided
	if os.Getenv("FACEBOOK_ACCESS_TOKEN") == "" {
		log.Fatal("FACEBOOK_ACCESS_TOKEN environment variable must be set")
	}

	// Run the MCP server
	if err := facebook.RunServer(); err != nil {
		log.Fatal(err)
	}
}

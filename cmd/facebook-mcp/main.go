package main

import (
	"log"
	"os"

	"unified-ads-mcp/internal/utils"
)

func main() {
	// Load .env file
	utils.LoadEnv()

	// Check if access token is provided
	if os.Getenv("FACEBOOK_ACCESS_TOKEN") == "" {
		log.Fatal("FACEBOOK_ACCESS_TOKEN environment variable must be set")
	}
}

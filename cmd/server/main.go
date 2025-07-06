package main

import (
	"log"

	"unified-ads-mcp/internal/config"
	"unified-ads-mcp/internal/mcp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	server, err := mcp.NewMCPServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// The mcp-go ServeStdio handles signals internally
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

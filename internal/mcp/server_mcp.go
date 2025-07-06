package mcp

import (
	"fmt"
	"log"

	"unified-ads-mcp/internal/config"
	"unified-ads-mcp/internal/facebook"

	"github.com/mark3labs/mcp-go/server"
)

// MCPServer wraps the mcp-go server
type MCPServer struct {
	server *server.MCPServer
	config *config.Config
}

// NewMCPServer creates a new MCP server using mcp-go
func NewMCPServer(cfg *config.Config) (*MCPServer, error) {
	// Create the mcp-go server
	mcpServer := server.NewMCPServer(
		"unified-ads-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	s := &MCPServer{
		server: mcpServer,
		config: cfg,
	}

	// Register Facebook tools
	if cfg.Facebook.AccessToken != "" {
		if err := facebook.RegisterMCPTools(mcpServer, cfg.Facebook.AccessToken); err != nil {
			return nil, fmt.Errorf("failed to register Facebook tools: %w", err)
		}
		log.Println("Facebook Business API tools registered")
	}

	// TODO: Register Google Ads tools when implemented
	// TODO: Register TikTok tools when implemented

	return s, nil
}

// Start runs the MCP server
func (s *MCPServer) Start() error {
	log.Println("Starting MCP server with mcp-go...")
	return server.ServeStdio(s.server)
}

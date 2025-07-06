package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"unified-ads-mcp/internal/config"
	"unified-ads-mcp/internal/facebook"
	"unified-ads-mcp/internal/google"
	"unified-ads-mcp/internal/tiktok"
	"unified-ads-mcp/pkg/types"
)

type Server struct {
	config         *config.Config
	googleClient   *google.Client
	facebookClient *facebook.Client
	tiktokClient   *tiktok.Client
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		config:         cfg,
		googleClient:   google.NewClient(cfg.Google),
		facebookClient: facebook.NewClient(cfg.Facebook),
		tiktokClient:   tiktok.NewClient(cfg.TikTok),
	}
}

func (s *Server) Start(ctx context.Context) error {
	log.Println("Starting MCP server...")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := s.handleRequest(); err != nil {
				if err == io.EOF {
					return nil
				}
				log.Printf("Error handling request: %v", err)
			}
		}
	}
}

func (s *Server) handleRequest() error {
	var request types.MCPRequest
	decoder := json.NewDecoder(os.Stdin)

	if err := decoder.Decode(&request); err != nil {
		return err
	}

	response, err := s.processRequest(&request)
	if err != nil {
		response = &types.MCPResponse{
			ID: request.ID,
			Error: &types.MCPError{
				Code:    -1,
				Message: err.Error(),
			},
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	return encoder.Encode(response)
}

func (s *Server) processRequest(request *types.MCPRequest) (*types.MCPResponse, error) {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(request)
	default:
		return nil, fmt.Errorf("unknown method: %s", request.Method)
	}
}

func (s *Server) handleInitialize(request *types.MCPRequest) (*types.MCPResponse, error) {
	return &types.MCPResponse{
		ID: request.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "unified-ads-mcp",
				"version": "1.0.0",
			},
		},
	}, nil
}

func (s *Server) handleToolsList(request *types.MCPRequest) (*types.MCPResponse, error) {
	var tools []map[string]interface{}

	// Add Google Ads tools (placeholder for now)
	googleTools := []map[string]interface{}{
		{
			"name":        "google_ads_create_campaign",
			"description": "Create a new Google Ads campaign",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":   map[string]interface{}{"type": "string"},
					"budget": map[string]interface{}{"type": "number"},
					"type":   map[string]interface{}{"type": "string"},
				},
				"required": []string{"name", "budget", "type"},
			},
		},
	}

	// Add Facebook Business API tools (generated)
	facebookTools := s.facebookClient.GetMCPTools()

	// Add TikTok Ads tools (placeholder for now)
	tiktokTools := []map[string]interface{}{
		{
			"name":        "tiktok_ads_create_campaign",
			"description": "Create a new TikTok Ads campaign",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":      map[string]interface{}{"type": "string"},
					"objective": map[string]interface{}{"type": "string"},
					"budget":    map[string]interface{}{"type": "number"},
				},
				"required": []string{"name", "objective", "budget"},
			},
		},
	}

	// Combine all tools
	tools = append(tools, googleTools...)
	tools = append(tools, facebookTools...)
	tools = append(tools, tiktokTools...)

	return &types.MCPResponse{
		ID: request.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}, nil
}

func (s *Server) handleToolsCall(request *types.MCPRequest) (*types.MCPResponse, error) {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params")
	}

	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing tool name")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing tool arguments")
	}

	var result interface{}
	var err error

	switch name {
	case "google_ads_create_campaign":
		result, err = s.googleClient.CreateCampaign(arguments)
	case "facebook_ads_create_campaign":
		result, err = s.facebookClient.CreateCampaign(arguments)
	case "tiktok_ads_create_campaign":
		result, err = s.tiktokClient.CreateCampaign(arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	if err != nil {
		return nil, err
	}

	return &types.MCPResponse{
		ID: request.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Tool %s executed successfully: %v", name, result),
				},
			},
		},
	}, nil
}

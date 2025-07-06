package facebook

import (
	"fmt"
	"unified-ads-mcp/internal/config"
)

type Client struct {
	config *config.FacebookConfig
}

func NewClient(cfg *config.FacebookConfig) *Client {
	return &Client{
		config: cfg,
	}
}

func (c *Client) CreateCampaign(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing campaign name")
	}

	objective, ok := args["objective"].(string)
	if !ok {
		return nil, fmt.Errorf("missing campaign objective")
	}

	budget, ok := args["budget"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing campaign budget")
	}

	campaign := map[string]interface{}{
		"id":        fmt.Sprintf("facebook_campaign_%d", generateID()),
		"name":      name,
		"objective": objective,
		"budget":    budget,
		"status":    "PAUSED",
		"platform":  "facebook_ads",
	}

	return campaign, nil
}

func generateID() int64 {
	return 67890
}

// GetMCPTools returns all available MCP tools for Facebook Business API
func (c *Client) GetMCPTools() []map[string]interface{} {
	// This method is deprecated, use GetMCPToolsV2 for mcp-go compatibility
	return []map[string]interface{}{}
}

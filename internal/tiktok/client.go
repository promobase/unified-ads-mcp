package tiktok

import (
	"fmt"
	"unified-ads-mcp/internal/config"
)

type Client struct {
	config *config.TikTokConfig
}

func NewClient(cfg *config.TikTokConfig) *Client {
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
		"id":        fmt.Sprintf("tiktok_campaign_%d", generateID()),
		"name":      name,
		"objective": objective,
		"budget":    budget,
		"status":    "PAUSED",
		"platform":  "tiktok_ads",
	}

	return campaign, nil
}

func generateID() int64 {
	return 11111
}

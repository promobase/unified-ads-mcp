package google

import (
	"fmt"
	"unified-ads-mcp/internal/config"
)

type Client struct {
	config *config.GoogleConfig
}

func NewClient(cfg *config.GoogleConfig) *Client {
	return &Client{
		config: cfg,
	}
}

func (c *Client) CreateCampaign(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing campaign name")
	}

	budget, ok := args["budget"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing campaign budget")
	}

	campaignType, ok := args["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing campaign type")
	}

	campaign := map[string]interface{}{
		"id":       fmt.Sprintf("google_campaign_%d", generateID()),
		"name":     name,
		"budget":   budget,
		"type":     campaignType,
		"status":   "PAUSED",
		"platform": "google_ads",
	}

	return campaign, nil
}

func generateID() int64 {
	return 12345
}

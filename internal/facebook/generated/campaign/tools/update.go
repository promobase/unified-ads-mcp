// Code generated by codegen. DO NOT EDIT.

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"unified-ads-mcp/internal/facebook/generated/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// campaign_updateArgs defines the typed arguments for campaign_update
type campaign_updateArgs struct {
	ID                        string                   `json:"id" jsonschema:"required,description=Campaign ID,pattern=^[0-9]+$"`
	Adlabels                  []*common.AdLabel        `json:"adlabels,omitempty" jsonschema:"description=Adlabels"`
	AdsetBidAmounts           map[string]interface{}   `json:"adset_bid_amounts,omitempty" jsonschema:"description=Adset Bid Amounts,minimum=1"`
	AdsetBudgets              []map[string]interface{} `json:"adset_budgets,omitempty" jsonschema:"description=Adset Budgets,minimum=1"`
	BidStrategy               string                   `json:"bid_strategy,omitempty" jsonschema:"description=Bid Strategy"`
	BudgetRebalanceFlag       bool                     `json:"budget_rebalance_flag,omitempty" jsonschema:"description=Budget Rebalance Flag"`
	DailyBudget               int                      `json:"daily_budget,omitempty" jsonschema:"description=Daily Budget,minimum=1"`
	ExecutionOptions          []string                 `json:"execution_options,omitempty" jsonschema:"description=Execution Options"`
	IsSkadnetworkAttribution  bool                     `json:"is_skadnetwork_attribution,omitempty" jsonschema:"description=Is Skadnetwork Attribution"`
	IterativeSplitTestConfigs []map[string]interface{} `json:"iterative_split_test_configs,omitempty" jsonschema:"description=Iterative Split Test Configs"`
	LifetimeBudget            int                      `json:"lifetime_budget,omitempty" jsonschema:"description=Lifetime Budget,minimum=1"`
	Name                      string                   `json:"name,omitempty" jsonschema:"description=Name"`
	Objective                 string                   `json:"objective,omitempty" jsonschema:"description=Objective"`
	PacingType                []string                 `json:"pacing_type,omitempty" jsonschema:"description=Pacing Type"`
	PromotedObject            *common.AdPromotedObject `json:"promoted_object,omitempty" jsonschema:"description=Promoted Object"`
	SmartPromotionType        string                   `json:"smart_promotion_type,omitempty" jsonschema:"description=Smart Promotion Type"`
	SpecialAdCategories       []string                 `json:"special_ad_categories,omitempty" jsonschema:"description=Special Ad Categories"`
	SpecialAdCategory         string                   `json:"special_ad_category,omitempty" jsonschema:"description=Special Ad Category"`
	SpecialAdCategoryCountry  []string                 `json:"special_ad_category_country,omitempty" jsonschema:"description=Special Ad Category Country"`
	SpendCap                  int                      `json:"spend_cap,omitempty" jsonschema:"description=Spend Cap"`
	StartTime                 string                   `json:"start_time,omitempty" jsonschema:"description=Start Time,format=date-time"`
	Status                    string                   `json:"status,omitempty" jsonschema:"description=Status,enum=ACTIVE,enum=PAUSED,enum=DELETED,enum=ARCHIVED"`
	StopTime                  string                   `json:"stop_time,omitempty" jsonschema:"description=Stop Time,format=date-time"`
}

// RegisterCampaignUpdateHandler registers the campaign_update tool
func RegisterCampaignUpdateHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"campaign_update",
		"Update a Campaign Returns Campaign.",
		json.RawMessage(`{"additionalProperties":false,"properties":{"adlabels":{"description":"Adlabels","items":{"additionalProperties":true,"type":"object"},"type":"array"},"adset_bid_amounts":{"description":"Adset Bid Amounts","type":"string"},"adset_budgets":{"description":"Adset Budgets","items":{"additionalProperties":true,"type":"object"},"type":"array"},"bid_strategy":{"description":"Bid Strategy (enum: adcampaigngroup_bid_strategy)","enum":["COST_CAP","LOWEST_COST_WITHOUT_CAP","LOWEST_COST_WITH_BID_CAP","LOWEST_COST_WITH_MIN_ROAS"],"type":"string"},"budget_rebalance_flag":{"description":"Budget Rebalance Flag","type":"boolean"},"daily_budget":{"description":"Daily Budget","type":"integer"},"execution_options":{"description":"Execution Options","items":{"type":"string"},"type":"array"},"id":{"description":"Campaign ID","pattern":"^[0-9]+$","type":"string"},"is_skadnetwork_attribution":{"description":"Is Skadnetwork Attribution","type":"boolean"},"iterative_split_test_configs":{"description":"Iterative Split Test Configs","items":{"additionalProperties":true,"type":"object"},"type":"array"},"lifetime_budget":{"description":"Lifetime Budget","type":"integer"},"name":{"description":"Name","type":"string"},"objective":{"description":"Objective (enum: adcampaigngroup_objective)","enum":["APP_INSTALLS","BRAND_AWARENESS","CONVERSIONS","EVENT_RESPONSES","LEAD_GENERATION","LINK_CLICKS","LOCAL_AWARENESS","MESSAGES","OFFER_CLAIMS","OUTCOME_APP_PROMOTION","OUTCOME_AWARENESS","OUTCOME_ENGAGEMENT","OUTCOME_LEADS","OUTCOME_SALES","OUTCOME_TRAFFIC","PAGE_LIKES","POST_ENGAGEMENT","PRODUCT_CATALOG_SALES","REACH","STORE_VISITS","VIDEO_VIEWS"],"type":"string"},"pacing_type":{"description":"Pacing Type","items":{"type":"string"},"type":"array"},"promoted_object":{"additionalProperties":true,"description":"Promoted Object","type":"object"},"smart_promotion_type":{"description":"Smart Promotion Type (enum: adcampaigngroup_smart_promotion_type)","enum":["GUIDED_CREATION","SMART_APP_PROMOTION"],"type":"string"},"special_ad_categories":{"description":"Special Ad Categories","items":{"type":"string"},"type":"array"},"special_ad_category":{"description":"Special Ad Category (enum: adcampaigngroup_special_ad_category)","enum":["CREDIT","EMPLOYMENT","FINANCIAL_PRODUCTS_SERVICES","HOUSING","ISSUES_ELECTIONS_POLITICS","NONE","ONLINE_GAMBLING_AND_GAMING"],"type":"string"},"special_ad_category_country":{"description":"Special Ad Category Country","items":{"type":"string"},"type":"array"},"spend_cap":{"description":"Spend Cap","type":"integer"},"start_time":{"description":"Start Time","type":"string"},"status":{"description":"Status (enum: adcampaigngroup_status)","enum":["ACTIVE","ARCHIVED","DELETED","PAUSED"],"type":"string"},"stop_time":{"description":"Stop Time","type":"string"}},"required":["id"],"type":"object"}`),
	)

	s.AddTool(tool, CampaignUpdateHandler)
	return nil
}

// CampaignUpdateHandler handles the campaign_update tool
func CampaignUpdateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args campaign_updateArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s", args.ID)
	// Prepare request body
	body := make(map[string]interface{})
	if len(args.Adlabels) > 0 {
		body["adlabels"] = args.Adlabels
	}
	body["adset_bid_amounts"] = args.AdsetBidAmounts
	if len(args.AdsetBudgets) > 0 {
		body["adset_budgets"] = args.AdsetBudgets
	}
	if args.BidStrategy != "" {
		body["bid_strategy"] = args.BidStrategy
	}
	body["budget_rebalance_flag"] = args.BudgetRebalanceFlag
	if args.DailyBudget > 0 {
		body["daily_budget"] = args.DailyBudget
	}
	if len(args.ExecutionOptions) > 0 {
		body["execution_options"] = args.ExecutionOptions
	}
	body["is_skadnetwork_attribution"] = args.IsSkadnetworkAttribution
	if len(args.IterativeSplitTestConfigs) > 0 {
		body["iterative_split_test_configs"] = args.IterativeSplitTestConfigs
	}
	if args.LifetimeBudget > 0 {
		body["lifetime_budget"] = args.LifetimeBudget
	}
	if args.Name != "" {
		body["name"] = args.Name
	}
	if args.Objective != "" {
		body["objective"] = args.Objective
	}
	if len(args.PacingType) > 0 {
		body["pacing_type"] = args.PacingType
	}
	if args.PromotedObject != nil {
		body["promoted_object"] = args.PromotedObject
	}
	if args.SmartPromotionType != "" {
		body["smart_promotion_type"] = args.SmartPromotionType
	}
	if len(args.SpecialAdCategories) > 0 {
		body["special_ad_categories"] = args.SpecialAdCategories
	}
	if args.SpecialAdCategory != "" {
		body["special_ad_category"] = args.SpecialAdCategory
	}
	if len(args.SpecialAdCategoryCountry) > 0 {
		body["special_ad_category_country"] = args.SpecialAdCategoryCountry
	}
	if args.SpendCap > 0 {
		body["spend_cap"] = args.SpendCap
	}
	if args.StartTime != "" {
		body["start_time"] = args.StartTime
	}
	if args.Status != "" {
		body["status"] = args.Status
	}
	if args.StopTime != "" {
		body["stop_time"] = args.StopTime
	}

	result, err := common.MakeGraphAPIRequest(ctx, "POST", endpoint, nil, body)

	if err != nil {
		return common.HandleAPIError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(result)),
		},
	}, nil
}

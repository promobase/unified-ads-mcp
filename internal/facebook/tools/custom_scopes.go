package tools

import (
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	adtools "unified-ads-mcp/internal/facebook/generated/ad/tools"
	adaccounttools "unified-ads-mcp/internal/facebook/generated/adaccount/tools"
	adcreativetools "unified-ads-mcp/internal/facebook/generated/adcreative/tools"
	adsettools "unified-ads-mcp/internal/facebook/generated/adset/tools"
	campaigntools "unified-ads-mcp/internal/facebook/generated/campaign/tools"
	customaudiencetools "unified-ads-mcp/internal/facebook/generated/customaudience/tools"
)

// CustomScope represents a curated set of tools for specific workflows
type CustomScope struct {
	Name        string
	Description string
	RegisterFn  func(*server.MCPServer) error
}

// customScopes defines all available custom scopes
var customScopes = map[string]CustomScope{
	"essentials": {
		Name:        "essentials",
		Description: "[RECOMMENDED] Core CRUD operations for campaigns, ad sets, and ads. Perfect starting point for basic ad management (16 tools)",
		RegisterFn:  registerEssentialsScope,
	},
	"campaign_management": {
		Name:        "campaign_management",
		Description: "[POPULAR] Complete campaign lifecycle management - create, update, monitor campaigns and their hierarchy (17 tools)",
		RegisterFn:  registerCampaignManagementScope,
	},
	"reporting": {
		Name:        "reporting",
		Description: "[ANALYTICS] Comprehensive insights and reporting across all levels - account, campaign, ad set, ad (11 tools)",
		RegisterFn:  registerReportingScope,
	},
	"audience": {
		Name:        "audience",
		Description: "[TARGETING] Audience creation, custom audiences, targeting browse/search, and reach estimates (14 tools)",
		RegisterFn:  registerAudienceScope,
	},
	"creative": {
		Name:        "creative",
		Description: "[CONTENT] Creative assets, images, videos, and ad preview management (15 tools)",
		RegisterFn:  registerCreativeScope,
	},
	"optimization": {
		Name:        "optimization",
		Description: "[PERFORMANCE] Optimization tools - delivery estimates, recommendations, budgets, and bidding (10 tools)",
		RegisterFn:  registerOptimizationScope,
	},
	"video": {
		Name:        "video",
		Description: "[VIDEO] Video upload and management tools - upload single or batch videos, check encoding status (3 tools)",
		RegisterFn:  registerVideoScope,
	},
}

// registerEssentialsScope loads core CRUD operations
func registerEssentialsScope(s *server.MCPServer) error {
	// AdAccount essentials
	if err := adaccounttools.RegisterAdAccountGetHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListCampaignsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_campaigns: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListAdsetsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_adsets: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListAdsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_ads: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountCreateCampaignHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_campaign: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountCreateAdsetHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_adset: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountCreateAdHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_ad: %w", err)
	}

	// Campaign essentials
	if err := campaigntools.RegisterCampaignGetHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_get: %w", err)
	}
	if err := campaigntools.RegisterCampaignUpdateHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_update: %w", err)
	}
	if err := campaigntools.RegisterCampaignDeleteHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_delete: %w", err)
	}

	// AdSet essentials
	if err := adsettools.RegisterAdSetGetHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_get: %w", err)
	}
	if err := adsettools.RegisterAdSetUpdateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_update: %w", err)
	}
	if err := adsettools.RegisterAdSetDeleteHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_delete: %w", err)
	}

	// Ad essentials
	if err := adtools.RegisterAdGetHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_get: %w", err)
	}
	if err := adtools.RegisterAdUpdateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_update: %w", err)
	}
	if err := adtools.RegisterAdDeleteHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_delete: %w", err)
	}

	return nil
}

// registerCampaignManagementScope loads campaign lifecycle management tools
func registerCampaignManagementScope(s *server.MCPServer) error {
	// Campaign creation and management
	if err := adaccounttools.RegisterAdAccountCreateCampaignHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_campaign: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListCampaignsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_campaigns: %w", err)
	}
	if err := campaigntools.RegisterCampaignGetHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_get: %w", err)
	}
	if err := campaigntools.RegisterCampaignUpdateHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_update: %w", err)
	}
	if err := campaigntools.RegisterCampaignDeleteHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_delete: %w", err)
	}
	if err := campaigntools.RegisterCampaignListAdsetsHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_list_adsets: %w", err)
	}
	if err := campaigntools.RegisterCampaignListAdsHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_list_ads: %w", err)
	}
	if err := campaigntools.RegisterCampaignGetInsightsHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_get_insights: %w", err)
	}
	if err := campaigntools.RegisterCampaignCreateBudgetScheduleHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_create_budget_schedule: %w", err)
	}

	// AdSet management
	if err := adaccounttools.RegisterAdAccountCreateAdsetHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_adset: %w", err)
	}
	if err := adsettools.RegisterAdSetGetHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_get: %w", err)
	}
	if err := adsettools.RegisterAdSetUpdateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_update: %w", err)
	}
	if err := adsettools.RegisterAdSetListAdsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_list_ads: %w", err)
	}

	// Ad management
	if err := adaccounttools.RegisterAdAccountCreateAdHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_ad: %w", err)
	}
	if err := adtools.RegisterAdGetHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_get: %w", err)
	}
	if err := adtools.RegisterAdUpdateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_update: %w", err)
	}

	return nil
}

// registerReportingScope loads analytics and insights tools
func registerReportingScope(s *server.MCPServer) error {
	// Account level insights
	if err := adaccounttools.RegisterAdAccountGetInsightsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_insights: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountCreateInsightsReportHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_insights_report: %w", err)
	}

	// Campaign insights
	if err := campaigntools.RegisterCampaignGetInsightsHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_get_insights: %w", err)
	}
	if err := campaigntools.RegisterCampaignCreateInsightsReportHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_create_insights_report: %w", err)
	}

	// AdSet insights
	if err := adsettools.RegisterAdSetGetInsightsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_get_insights: %w", err)
	}
	if err := adsettools.RegisterAdSetCreateInsightsReportHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_create_insights_report: %w", err)
	}

	// Ad insights
	if err := adtools.RegisterAdGetInsightsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_get_insights: %w", err)
	}
	if err := adtools.RegisterAdCreateInsightsReportHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_create_insights_report: %w", err)
	}

	// Creative insights
	if err := adcreativetools.RegisterAdCreativeListCreativeInsightsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_creative_list_creative_insights: %w", err)
	}

	// Additional reporting tools
	if err := adaccounttools.RegisterAdAccountListActivitiesHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_activities: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountGetAdsVolumeHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_ads_volume: %w", err)
	}

	return nil
}

// registerAudienceScope loads audience and targeting tools
func registerAudienceScope(s *server.MCPServer) error {
	// Custom audience management
	if err := adaccounttools.RegisterAdAccountCreateCustomaudienceHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_customaudience: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListCustomaudiencesHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_customaudiences: %w", err)
	}
	if err := customaudiencetools.RegisterCustomAudienceGetHandler(s); err != nil {
		return fmt.Errorf("failed to register custom_audience_get: %w", err)
	}
	if err := customaudiencetools.RegisterCustomAudienceUpdateHandler(s); err != nil {
		return fmt.Errorf("failed to register custom_audience_update: %w", err)
	}
	if err := customaudiencetools.RegisterCustomAudienceCreateUserHandler(s); err != nil {
		return fmt.Errorf("failed to register custom_audience_create_user: %w", err)
	}
	if err := customaudiencetools.RegisterCustomAudienceRemoveUsersHandler(s); err != nil {
		return fmt.Errorf("failed to register custom_audience_remove_users: %w", err)
	}

	// Targeting tools
	if err := adaccounttools.RegisterAdAccountGetTargetingbrowseHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_targetingbrowse: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountGetTargetingsearchHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_targetingsearch: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountGetTargetingvalidationHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_targetingvalidation: %w", err)
	}

	// Reach and delivery estimates
	if err := adaccounttools.RegisterAdAccountGetReachestimateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_reachestimate: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountGetDeliveryEstimateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_delivery_estimate: %w", err)
	}

	// Saved audiences
	if err := adaccounttools.RegisterAdAccountListSavedAudiencesHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_saved_audiences: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListBroadtargetingcategoriesHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_broadtargetingcategories: %w", err)
	}

	return nil
}

// registerCreativeScope loads creative management tools
func registerCreativeScope(s *server.MCPServer) error {
	// Ad creative management
	if err := adaccounttools.RegisterAdAccountCreateAdcreativeHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_adcreative: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListAdcreativesHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_adcreatives: %w", err)
	}
	if err := adcreativetools.RegisterAdCreativeGetHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_creative_get: %w", err)
	}
	if err := adcreativetools.RegisterAdCreativeUpdateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_creative_update: %w", err)
	}
	if err := adcreativetools.RegisterAdCreativeDeleteHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_creative_delete: %w", err)
	}
	if err := adcreativetools.RegisterAdCreativeListPreviewsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_creative_list_previews: %w", err)
	}

	// Image management
	if err := adaccounttools.RegisterAdAccountCreateAdimageHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_adimage: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListAdimagesHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_adimages: %w", err)
	}

	// Video management
	if err := adaccounttools.RegisterAdAccountCreateAdvideoHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_advideo: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListAdvideosHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_advideos: %w", err)
	}

	// Preview tools
	if err := adtools.RegisterAdListPreviewsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_list_previews: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListGeneratepreviewsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_generatepreviews: %w", err)
	}

	// Video tools (also in creative scope)
	if err := RegisterVideoUploadTool(s); err != nil {
		return fmt.Errorf("failed to register video upload tool: %w", err)
	}
	if err := RegisterVideoStatusTool(s); err != nil {
		return fmt.Errorf("failed to register video status tool: %w", err)
	}
	if err := RegisterVideoUploadBatchTool(s); err != nil {
		return fmt.Errorf("failed to register video batch upload tool: %w", err)
	}

	return nil
}

// registerOptimizationScope loads performance optimization tools
func registerOptimizationScope(s *server.MCPServer) error {
	// Delivery and reach estimates
	if err := adaccounttools.RegisterAdAccountGetDeliveryEstimateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_delivery_estimate: %w", err)
	}
	if err := adsettools.RegisterAdSetGetDeliveryEstimateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_get_delivery_estimate: %w", err)
	}
	if err := adsettools.RegisterAdSetGetMessageDeliveryEstimateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_get_message_delivery_estimate: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountGetReachestimateHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_reachestimate: %w", err)
	}

	// Recommendations
	if err := adaccounttools.RegisterAdAccountListRecommendationsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_recommendations: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountCreateRecommendationHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_create_recommendation: %w", err)
	}

	// Budget optimization
	if err := adaccounttools.RegisterAdAccountGetMaxBidHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_get_max_bid: %w", err)
	}
	if err := adaccounttools.RegisterAdAccountListMinimumBudgetsHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_account_list_minimum_budgets: %w", err)
	}
	if err := campaigntools.RegisterCampaignCreateBudgetScheduleHandler(s); err != nil {
		return fmt.Errorf("failed to register campaign_create_budget_schedule: %w", err)
	}
	if err := adsettools.RegisterAdSetCreateBudgetScheduleHandler(s); err != nil {
		return fmt.Errorf("failed to register ad_set_create_budget_schedule: %w", err)
	}

	return nil
}

// isCustomScope checks if a scope is a custom scope
func isCustomScope(scope string) bool {
	_, exists := customScopes[scope]
	return exists
}

// loadCustomScope loads a custom scope
func loadCustomScope(scope string, s *server.MCPServer) error {
	customScope, exists := customScopes[scope]
	if !exists {
		return fmt.Errorf("unknown custom scope: %s", scope)
	}
	return customScope.RegisterFn(s)
}

// getCustomScopeNames returns all available custom scope names
func getCustomScopeNames() []string {
	names := make([]string, 0, len(customScopes))
	for name := range customScopes {
		names = append(names, name)
	}
	return names
}

// getCustomScopeDescriptions returns a map of custom scope names to descriptions
func getCustomScopeDescriptions() map[string]string {
	descriptions := make(map[string]string)
	for name, scope := range customScopes {
		descriptions[name] = scope.Description
	}
	return descriptions
}

// registerVideoScope loads video upload and management tools
func registerVideoScope(s *server.MCPServer) error {
	// Video upload tools
	if err := RegisterVideoUploadTool(s); err != nil {
		return fmt.Errorf("failed to register video upload tool: %w", err)
	}
	if err := RegisterVideoStatusTool(s); err != nil {
		return fmt.Errorf("failed to register video status tool: %w", err)
	}
	if err := RegisterVideoUploadBatchTool(s); err != nil {
		return fmt.Errorf("failed to register video batch upload tool: %w", err)
	}

	return nil
}

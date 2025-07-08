package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"unified-ads-mcp/internal/facebook/generated"
)

// RegisterBatchTools registers batch-related tools with the MCP server
func RegisterBatchTools(s *server.MCPServer) {
	// Execute batch requests tool
	s.AddTool(
		mcp.NewTool(
			"execute_batch_requests",
			mcp.WithDescription("Execute multiple Facebook Graph API requests in a single batch. Maximum 50 requests per batch."),
			mcp.WithString("requests",
				mcp.Required(),
				mcp.Description("JSON array of batch requests. Each request should have: method, relative_url, and optional body/name fields"),
			),
		),
		ExecuteBatchRequestsHandler,
	)

	// Batch get multiple objects
	s.AddTool(
		mcp.NewTool(
			"batch_get_objects",
			mcp.WithDescription("Get multiple Facebook objects in a single batch request"),
			mcp.WithArray("object_ids",
				mcp.Required(),
				mcp.Description("Array of object IDs to fetch (e.g., ['act_123', 'campaign_456'])"),
				mcp.Items(map[string]interface{}{"type": "string"}),
			),
			mcp.WithArray("fields",
				mcp.Description("Fields to return for each object"),
				mcp.Items(map[string]interface{}{"type": "string"}),
			),
		),
		BatchGetObjectsHandler,
	)

	// Batch update multiple campaigns
	s.AddTool(
		mcp.NewTool(
			"batch_update_campaigns",
			mcp.WithDescription("Update multiple campaigns in a single batch request"),
			mcp.WithString("updates",
				mcp.Required(),
				mcp.Description("JSON object mapping campaign IDs to their update parameters"),
			),
		),
		BatchUpdateCampaignsHandler,
	)

	// Batch create ad sets
	s.AddTool(
		mcp.NewTool(
			"batch_create_adsets",
			mcp.WithDescription("Create multiple ad sets in a single batch request"),
			mcp.WithString("campaign_id",
				mcp.Required(),
				mcp.Description("Campaign ID to create ad sets under"),
			),
			mcp.WithString("adsets",
				mcp.Required(),
				mcp.Description("JSON array of ad set configurations"),
			),
		),
		BatchCreateAdsetsHandler,
	)

	// Batch insights requests
	s.AddTool(
		mcp.NewTool(
			"batch_get_insights",
			mcp.WithDescription("Get insights for multiple objects in a single batch request"),
			mcp.WithArray("object_ids",
				mcp.Required(),
				mcp.Description("Array of object IDs to get insights for"),
				mcp.Items(map[string]interface{}{"type": "string"}),
			),
			mcp.WithString("level",
				mcp.Description("Insights level: account, campaign, adset, or ad"),
			),
			mcp.WithString("date_preset",
				mcp.Description("Date preset like 'yesterday', 'last_7d', 'last_30d'"),
			),
			mcp.WithArray("fields",
				mcp.Description("Insight fields to return"),
				mcp.Items(map[string]interface{}{"type": "string"}),
			),
		),
		BatchGetInsightsHandler,
	)
}

// ExecuteBatchRequestsHandler handles generic batch requests
func ExecuteBatchRequestsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get the requests parameter
	requestsJSON := request.GetString("requests", "")
	if requestsJSON == "" {
		return mcp.NewToolResultErrorf("requests parameter is required"), nil
	}

	// Parse the requests
	var batchRequests []generated.BatchRequest
	if err := json.Unmarshal([]byte(requestsJSON), &batchRequests); err != nil {
		return mcp.NewToolResultErrorf("invalid requests JSON: %v", err), nil
	}

	// Validate request count
	if len(batchRequests) == 0 {
		return mcp.NewToolResultErrorf("at least one request is required"), nil
	}
	if len(batchRequests) > 50 {
		return mcp.NewToolResultErrorf("maximum 50 requests allowed, got %d", len(batchRequests)), nil
	}

	// Execute batch request
	responses, err := generated.MakeBatchRequest(batchRequests)
	if err != nil {
		return mcp.NewToolResultErrorf("batch request failed: %v", err), nil
	}

	// Format responses
	result := make([]map[string]interface{}, len(responses))
	for i, resp := range responses {
		resultItem := map[string]interface{}{
			"code": resp.Code,
		}

		if resp.Headers != nil {
			resultItem["headers"] = resp.Headers
		}

		if resp.Body != nil {
			var bodyData interface{}
			if err := json.Unmarshal(resp.Body, &bodyData); err == nil {
				resultItem["body"] = bodyData
			} else {
				resultItem["body"] = string(resp.Body)
			}
		}

		result[i] = resultItem
	}

	// Return as JSON
	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// BatchGetObjectsHandler handles batch GET requests for multiple objects
func BatchGetObjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	objectIDs := request.GetStringSlice("object_ids", nil)
	if len(objectIDs) == 0 {
		return mcp.NewToolResultErrorf("object_ids is required"), nil
	}

	fields := request.GetStringSlice("fields", nil)

	// Build batch requests
	builder := generated.NewBatchRequestBuilder()

	for i, objectID := range objectIDs {
		params := make(map[string]interface{})
		if len(fields) > 0 {
			params["fields"] = fields
		}

		builder.AddGET(objectID, "", params, fmt.Sprintf("get_%d", i))
	}

	// Execute batch
	responses, err := builder.Execute()
	if err != nil {
		return mcp.NewToolResultErrorf("batch request failed: %v", err), nil
	}

	// Format responses
	result := make(map[string]interface{})
	for i, resp := range responses {
		if i < len(objectIDs) {
			if resp.Code == 200 && resp.Body != nil {
				var data interface{}
				if err := json.Unmarshal(resp.Body, &data); err == nil {
					result[objectIDs[i]] = data
				} else {
					result[objectIDs[i]] = map[string]interface{}{
						"error": fmt.Sprintf("failed to parse response: %v", err),
					}
				}
			} else {
				result[objectIDs[i]] = map[string]interface{}{
					"error": fmt.Sprintf("request failed with code %d", resp.Code),
					"body":  string(resp.Body),
				}
			}
		}
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// BatchUpdateCampaignsHandler handles batch updates for multiple campaigns
func BatchUpdateCampaignsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	updatesJSON := request.GetString("updates", "")
	if updatesJSON == "" {
		return mcp.NewToolResultErrorf("updates parameter is required"), nil
	}

	// Parse updates
	var updates map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(updatesJSON), &updates); err != nil {
		return mcp.NewToolResultErrorf("invalid updates JSON: %v", err), nil
	}

	if len(updates) == 0 {
		return mcp.NewToolResultErrorf("at least one update is required"), nil
	}

	// Build batch requests
	builder := generated.NewBatchRequestBuilder()
	campaignIDs := make([]string, 0, len(updates))

	for campaignID, updateParams := range updates {
		campaignIDs = append(campaignIDs, campaignID)
		builder.AddPOST(campaignID, "", updateParams, fmt.Sprintf("update_%s", campaignID))
	}

	// Execute batch
	responses, err := builder.Execute()
	if err != nil {
		return mcp.NewToolResultErrorf("batch request failed: %v", err), nil
	}

	// Format responses
	result := make(map[string]interface{})
	for i, resp := range responses {
		if i < len(campaignIDs) {
			campaignID := campaignIDs[i]
			if resp.Code == 200 {
				result[campaignID] = map[string]interface{}{
					"success": true,
					"code":    resp.Code,
				}
			} else {
				var errorData interface{}
				if resp.Body != nil {
					json.Unmarshal(resp.Body, &errorData)
				}
				result[campaignID] = map[string]interface{}{
					"success": false,
					"code":    resp.Code,
					"error":   errorData,
				}
			}
		}
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// BatchCreateAdsetsHandler handles batch creation of ad sets
func BatchCreateAdsetsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID := request.GetString("campaign_id", "")
	if campaignID == "" {
		return mcp.NewToolResultErrorf("campaign_id is required"), nil
	}

	adsetsJSON := request.GetString("adsets", "")
	if adsetsJSON == "" {
		return mcp.NewToolResultErrorf("adsets parameter is required"), nil
	}

	// Parse ad sets
	var adsets []map[string]interface{}
	if err := json.Unmarshal([]byte(adsetsJSON), &adsets); err != nil {
		return mcp.NewToolResultErrorf("invalid adsets JSON: %v", err), nil
	}

	if len(adsets) == 0 {
		return mcp.NewToolResultErrorf("at least one ad set is required"), nil
	}

	// Build batch requests
	builder := generated.NewBatchRequestBuilder()

	for i, adset := range adsets {
		// Add campaign_id to each ad set
		adset["campaign_id"] = campaignID
		builder.AddPOST("act_"+campaignID, "adsets", adset, fmt.Sprintf("create_adset_%d", i))
	}

	// Execute batch
	responses, err := builder.Execute()
	if err != nil {
		return mcp.NewToolResultErrorf("batch request failed: %v", err), nil
	}

	// Format responses
	result := make([]map[string]interface{}, len(responses))
	for i, resp := range responses {
		if resp.Code == 200 && resp.Body != nil {
			var data map[string]interface{}
			if err := json.Unmarshal(resp.Body, &data); err == nil {
				result[i] = map[string]interface{}{
					"success": true,
					"id":      data["id"],
					"data":    data,
				}
			} else {
				result[i] = map[string]interface{}{
					"success": false,
					"error":   "failed to parse response",
				}
			}
		} else {
			var errorData interface{}
			if resp.Body != nil {
				json.Unmarshal(resp.Body, &errorData)
			}
			result[i] = map[string]interface{}{
				"success": false,
				"code":    resp.Code,
				"error":   errorData,
			}
		}
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// BatchGetInsightsHandler handles batch insights requests
func BatchGetInsightsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	objectIDs := request.GetStringSlice("object_ids", nil)
	if len(objectIDs) == 0 {
		return mcp.NewToolResultErrorf("object_ids is required"), nil
	}

	level := request.GetString("level", "")
	datePreset := request.GetString("date_preset", "yesterday")
	fields := request.GetStringSlice("fields", []string{"impressions", "clicks", "spend", "reach"})

	// Build batch requests
	builder := generated.NewBatchRequestBuilder()

	for i, objectID := range objectIDs {
		params := map[string]interface{}{
			"fields":      fields,
			"date_preset": datePreset,
		}
		if level != "" {
			params["level"] = level
		}

		builder.AddGET(objectID, "insights", params, fmt.Sprintf("insights_%d", i))
	}

	// Execute batch
	responses, err := builder.Execute()
	if err != nil {
		return mcp.NewToolResultErrorf("batch request failed: %v", err), nil
	}

	// Format responses
	result := make(map[string]interface{})
	for i, resp := range responses {
		if i < len(objectIDs) {
			if resp.Code == 200 && resp.Body != nil {
				var data interface{}
				if err := json.Unmarshal(resp.Body, &data); err == nil {
					result[objectIDs[i]] = data
				} else {
					result[objectIDs[i]] = map[string]interface{}{
						"error": fmt.Sprintf("failed to parse response: %v", err),
					}
				}
			} else {
				var errorData interface{}
				if resp.Body != nil {
					json.Unmarshal(resp.Body, &errorData)
				}
				result[objectIDs[i]] = map[string]interface{}{
					"error": errorData,
					"code":  resp.Code,
				}
			}
		}
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterBatchTools registers batch-related tools with the MCP server
func RegisterBatchTools(s *server.MCPServer) error {
	// Register the core batch tool first
	if err := RegisterCoreBatchTools(s); err != nil {
		return err
	}

	// Register convenience tools that build on top of the main batch tool

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

	// Batch get insights
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

	return nil
}

// BatchGetObjectsHandler handles batch GET requests for multiple objects
// This is now a convenience wrapper around the main facebook_batch tool
func BatchGetObjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	objectIDs := request.GetStringSlice("object_ids", nil)
	if len(objectIDs) == 0 {
		return mcp.NewToolResultErrorf("object_ids is required"), nil
	}

	if len(objectIDs) > 50 {
		return mcp.NewToolResultErrorf("maximum 50 object IDs allowed, got %d", len(objectIDs)), nil
	}

	fields := request.GetStringSlice("fields", nil)

	// Build operations for the new batch framework
	operations := make([]BatchOperationArgs, len(objectIDs))
	for i, objectID := range objectIDs {
		relativeURL := objectID
		if len(fields) > 0 {
			relativeURL = fmt.Sprintf("%s?fields=%s", objectID, strings.Join(fields, ","))
		}

		operations[i] = BatchOperationArgs{
			Method:      "GET",
			RelativeURL: relativeURL,
			Name:        fmt.Sprintf("get_%s", objectID),
		}
	}

	// Create batch request
	batchArgs := BatchRequestArgs{
		Operations: operations,
	}

	// Execute using the new batch handler
	batchRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"operations": operations,
			},
		},
	}

	result, err := FacebookBatchHandler(ctx, batchRequest, batchArgs)
	if err != nil {
		return mcp.NewToolResultErrorf("batch request failed: %v", err), nil
	}

	// Parse the result to extract object data
	var batchResult BatchResult

	// Extract text content from the result
	if len(result.Content) == 0 {
		return mcp.NewToolResultErrorf("no content in batch result"), nil
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		return mcp.NewToolResultErrorf("expected text content in batch result"), nil
	}

	if err := json.Unmarshal([]byte(textContent.Text), &batchResult); err != nil {
		return mcp.NewToolResultErrorf("failed to parse batch result: %v", err), nil
	}

	// Format as object ID -> data mapping
	objectResults := make(map[string]interface{})
	for i, opResult := range batchResult.Results {
		if i < len(objectIDs) {
			objectID := objectIDs[i]
			if opResult.Success {
				objectResults[objectID] = opResult.ParsedBody
			} else {
				objectResults[objectID] = map[string]interface{}{
					"error": opResult.Error,
					"code":  opResult.Code,
				}
			}
		}
	}

	// Add summary information
	summary := map[string]interface{}{
		"total_objects":      len(objectIDs),
		"successful_objects": batchResult.SuccessfulOperations,
		"failed_objects":     batchResult.FailedOperations,
		"success_rate":       batchResult.Summary["success_rate"],
		"objects":            objectResults,
	}

	resultJSON, _ := json.Marshal(summary)
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

	if len(updates) > 50 {
		return mcp.NewToolResultErrorf("maximum 50 updates allowed, got %d", len(updates)), nil
	}

	// Build operations for the new batch framework
	operations := make([]BatchOperationArgs, 0, len(updates))
	campaignIDs := make([]string, 0, len(updates))

	for campaignID, updateParams := range updates {
		campaignIDs = append(campaignIDs, campaignID)
		operations = append(operations, BatchOperationArgs{
			Method:      "POST",
			RelativeURL: campaignID,
			Body:        updateParams,
			Name:        fmt.Sprintf("update_%s", campaignID),
		})
	}

	// Create batch request
	batchArgs := BatchRequestArgs{
		Operations: operations,
	}

	// Execute using the new batch handler
	batchRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"operations": operations,
			},
		},
	}

	result, err := FacebookBatchHandler(ctx, batchRequest, batchArgs)
	if err != nil {
		return mcp.NewToolResultErrorf("batch request failed: %v", err), nil
	}

	// Parse the result to extract campaign update results
	var batchResult BatchResult

	// Extract text content from the result
	if len(result.Content) == 0 {
		return mcp.NewToolResultErrorf("no content in batch result"), nil
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		return mcp.NewToolResultErrorf("expected text content in batch result"), nil
	}

	if err := json.Unmarshal([]byte(textContent.Text), &batchResult); err != nil {
		return mcp.NewToolResultErrorf("failed to parse batch result: %v", err), nil
	}

	// Format as campaign ID -> update result mapping
	campaignResults := make(map[string]interface{})
	for i, opResult := range batchResult.Results {
		if i < len(campaignIDs) {
			campaignID := campaignIDs[i]
			if opResult.Success {
				campaignResults[campaignID] = map[string]interface{}{
					"success": true,
					"code":    opResult.Code,
					"data":    opResult.ParsedBody,
				}
			} else {
				campaignResults[campaignID] = map[string]interface{}{
					"success": false,
					"error":   opResult.Error,
					"code":    opResult.Code,
				}
			}
		}
	}

	// Add summary information
	summary := map[string]interface{}{
		"total_campaigns":      len(campaignIDs),
		"successful_campaigns": batchResult.SuccessfulOperations,
		"failed_campaigns":     batchResult.FailedOperations,
		"success_rate":         batchResult.Summary["success_rate"],
		"campaigns":            campaignResults,
	}

	resultJSON, _ := json.Marshal(summary)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// BatchGetInsightsHandler handles batch insights requests
func BatchGetInsightsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	objectIDs := request.GetStringSlice("object_ids", nil)
	if len(objectIDs) == 0 {
		return mcp.NewToolResultErrorf("object_ids is required"), nil
	}

	if len(objectIDs) > 50 {
		return mcp.NewToolResultErrorf("maximum 50 object IDs allowed, got %d", len(objectIDs)), nil
	}

	level := request.GetString("level", "")
	datePreset := request.GetString("date_preset", "yesterday")
	fields := request.GetStringSlice("fields", []string{"impressions", "clicks", "spend", "reach"})

	// Build operations for the new batch framework
	operations := make([]BatchOperationArgs, len(objectIDs))
	for i, objectID := range objectIDs {
		// Build query parameters
		params := []string{
			fmt.Sprintf("fields=%s", strings.Join(fields, ",")),
			fmt.Sprintf("date_preset=%s", datePreset),
		}
		if level != "" {
			params = append(params, fmt.Sprintf("level=%s", level))
		}

		relativeURL := fmt.Sprintf("%s/insights?%s", objectID, strings.Join(params, "&"))

		operations[i] = BatchOperationArgs{
			Method:      "GET",
			RelativeURL: relativeURL,
			Name:        fmt.Sprintf("insights_%s", objectID),
		}
	}

	// Create batch request
	batchArgs := BatchRequestArgs{
		Operations: operations,
	}

	// Execute using the new batch handler
	batchRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"operations": operations,
			},
		},
	}

	result, err := FacebookBatchHandler(ctx, batchRequest, batchArgs)
	if err != nil {
		return mcp.NewToolResultErrorf("batch request failed: %v", err), nil
	}

	// Parse the result to extract insights data
	var batchResult BatchResult

	// Extract text content from the result
	if len(result.Content) == 0 {
		return mcp.NewToolResultErrorf("no content in batch result"), nil
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		return mcp.NewToolResultErrorf("expected text content in batch result"), nil
	}

	if err := json.Unmarshal([]byte(textContent.Text), &batchResult); err != nil {
		return mcp.NewToolResultErrorf("failed to parse batch result: %v", err), nil
	}

	// Format as object ID -> insights data mapping
	insightsResults := make(map[string]interface{})
	for i, opResult := range batchResult.Results {
		if i < len(objectIDs) {
			objectID := objectIDs[i]
			if opResult.Success {
				insightsResults[objectID] = opResult.ParsedBody
			} else {
				insightsResults[objectID] = map[string]interface{}{
					"error": opResult.Error,
					"code":  opResult.Code,
				}
			}
		}
	}

	// Add summary information
	summary := map[string]interface{}{
		"total_objects":      len(objectIDs),
		"successful_objects": batchResult.SuccessfulOperations,
		"failed_objects":     batchResult.FailedOperations,
		"success_rate":       batchResult.Summary["success_rate"],
		"date_preset":        datePreset,
		"level":              level,
		"fields":             fields,
		"insights":           insightsResults,
	}

	resultJSON, _ := json.Marshal(summary)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

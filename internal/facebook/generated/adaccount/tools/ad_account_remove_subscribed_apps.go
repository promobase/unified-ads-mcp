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

// ad_account_remove_subscribed_appsArgs defines the typed arguments for ad_account_remove_subscribed_apps
type ad_account_remove_subscribed_appsArgs struct {
	ID    string `json:"id" jsonschema:"required,description=AdAccount ID,pattern=^[0-9]+$"`
	AppId string `json:"app_id,omitempty" jsonschema:"description=ID of the App,pattern=^[0-9]+$"`
}

// RegisterAdAccountRemoveSubscribedAppsHandler registers the ad_account_remove_subscribed_apps tool
func RegisterAdAccountRemoveSubscribedAppsHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"ad_account_remove_subscribed_apps",
		"Remove subscribed_apps from this AdAccount",
		json.RawMessage(`{"additionalProperties":false,"properties":{"app_id":{"description":"ID of the App","pattern":"^[0-9]+$","type":"string"},"id":{"description":"AdAccount ID","pattern":"^[0-9]+$","type":"string"}},"required":["id"],"type":"object"}`),
	)

	s.AddTool(tool, AdAccountRemoveSubscribedAppsHandler)
	return nil
}

// AdAccountRemoveSubscribedAppsHandler handles the ad_account_remove_subscribed_apps tool
func AdAccountRemoveSubscribedAppsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ad_account_remove_subscribed_appsArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s/subscribed_apps", args.ID)
	// Prepare query parameters
	params := make(map[string]string)
	// ID is part of path, not query params
	if args.AppId != "" {
		params["app_id"] = args.AppId
	}

	result, err := common.MakeGraphAPIRequest(ctx, "DELETE", endpoint, params, nil)

	if err != nil {
		return common.HandleAPIError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(result)),
		},
	}, nil
}

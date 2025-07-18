// Code generated by codegen. DO NOT EDIT.

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"unified-ads-mcp/internal/facebook/generated/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ad_account_list_ad_place_page_setsArgs defines the typed arguments for ad_account_list_ad_place_page_sets
type ad_account_list_ad_place_page_setsArgs struct {
	ID     string   `json:"id" jsonschema:"required,description=AdAccount ID,pattern=^[0-9]+$"`
	Fields []string `json:"fields,omitempty" jsonschema:"description=Fields to return"`
	Limit  int      `json:"limit,omitempty" jsonschema:"description=Maximum number of results,minimum=1,maximum=100"`
	After  string   `json:"after,omitempty" jsonschema:"description=Cursor for pagination (next page)"`
	Before string   `json:"before,omitempty" jsonschema:"description=Cursor for pagination (previous page)"`
}

// RegisterAdAccountListAdPlacePageSetsHandler registers the ad_account_list_ad_place_page_sets tool
func RegisterAdAccountListAdPlacePageSetsHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"ad_account_list_ad_place_page_sets",
		"List ad_place_page_sets for this AdAccount Returns AdPlacePageSet.",
		json.RawMessage(`{"additionalProperties":false,"properties":{"after":{"description":"Cursor for pagination (next page)","type":"string"},"before":{"description":"Cursor for pagination (previous page)","type":"string"},"fields":{"description":"Fields to return","items":{"type":"string"},"type":"array"},"id":{"description":"AdAccount ID","pattern":"^[0-9]+$","type":"string"},"limit":{"description":"Maximum number of results","maximum":100,"minimum":1,"type":"integer"}},"required":["id"],"type":"object"}`),
	)

	s.AddTool(tool, AdAccountListAdPlacePageSetsHandler)
	return nil
}

// AdAccountListAdPlacePageSetsHandler handles the ad_account_list_ad_place_page_sets tool
func AdAccountListAdPlacePageSetsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ad_account_list_ad_place_page_setsArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s/ad_place_page_sets", args.ID)
	// Prepare query parameters
	params := make(map[string]string)
	// ID is part of path, not query params
	if len(args.Fields) > 0 {
		params["fields"] = strings.Join(args.Fields, ",")
	}
	if args.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", args.Limit)
	}
	if args.After != "" {
		params["after"] = args.After
	}
	if args.Before != "" {
		params["before"] = args.Before
	}

	result, err := common.MakeGraphAPIRequest(ctx, "GET", endpoint, params, nil)

	if err != nil {
		return common.HandleAPIError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(result)),
		},
	}, nil
}

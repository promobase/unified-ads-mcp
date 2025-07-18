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

// user_list_groupsArgs defines the typed arguments for user_list_groups
type user_list_groupsArgs struct {
	ID        string   `json:"id" jsonschema:"required,description=User ID,pattern=^[0-9]+$"`
	Fields    []string `json:"fields,omitempty" jsonschema:"description=Fields to return"`
	Limit     int      `json:"limit,omitempty" jsonschema:"description=Maximum number of results,minimum=1,maximum=100"`
	After     string   `json:"after,omitempty" jsonschema:"description=Cursor for pagination (next page)"`
	Before    string   `json:"before,omitempty" jsonschema:"description=Cursor for pagination (previous page)"`
	AdminOnly bool     `json:"admin_only,omitempty" jsonschema:"description=Admin Only"`
	Parent    string   `json:"parent,omitempty" jsonschema:"description=Parent"`
}

// RegisterUserListGroupsHandler registers the user_list_groups tool
func RegisterUserListGroupsHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"user_list_groups",
		"List groups for this User Returns Group.",
		json.RawMessage(`{"additionalProperties":false,"properties":{"admin_only":{"description":"Admin Only","type":"boolean"},"after":{"description":"Cursor for pagination (next page)","type":"string"},"before":{"description":"Cursor for pagination (previous page)","type":"string"},"fields":{"description":"Fields to return","items":{"type":"string"},"type":"array"},"id":{"description":"User ID","pattern":"^[0-9]+$","type":"string"},"limit":{"description":"Maximum number of results","maximum":100,"minimum":1,"type":"integer"},"parent":{"description":"Parent","type":"string"}},"required":["id"],"type":"object"}`),
	)

	s.AddTool(tool, UserListGroupsHandler)
	return nil
}

// UserListGroupsHandler handles the user_list_groups tool
func UserListGroupsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args user_list_groupsArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s/groups", args.ID)
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
	params["admin_only"] = fmt.Sprintf("%v", args.AdminOnly)
	if args.Parent != "" {
		params["parent"] = args.Parent
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

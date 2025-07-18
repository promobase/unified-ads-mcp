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

// user_list_business_usersArgs defines the typed arguments for user_list_business_users
type user_list_business_usersArgs struct {
	Fields []string `json:"fields,omitempty" jsonschema:"description=Fields to return"`
	Limit  int      `json:"limit,omitempty" jsonschema:"description=Maximum number of results,minimum=1,maximum=100"`
	After  string   `json:"after,omitempty" jsonschema:"description=Cursor for pagination (next page)"`
	Before string   `json:"before,omitempty" jsonschema:"description=Cursor for pagination (previous page)"`
}

// RegisterUserListBusinessUsersHandler registers the user_list_business_users tool
func RegisterUserListBusinessUsersHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"user_list_business_users",
		"List business_users for this User Returns BusinessUser.",
		json.RawMessage(`{"additionalProperties":false,"properties":{"after":{"description":"Cursor for pagination (next page)","type":"string"},"before":{"description":"Cursor for pagination (previous page)","type":"string"},"fields":{"description":"Fields to return","items":{"type":"string"},"type":"array"},"limit":{"description":"Maximum number of results","maximum":100,"minimum":1,"type":"integer"}},"required":[],"type":"object"}`),
	)

	s.AddTool(tool, UserListBusinessUsersHandler)
	return nil
}

// UserListBusinessUsersHandler handles the user_list_business_users tool
func UserListBusinessUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args user_list_business_usersArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := "/business_users"
	// Prepare query parameters
	params := make(map[string]string)
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

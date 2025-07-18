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

// user_list_conversationsArgs defines the typed arguments for user_list_conversations
type user_list_conversationsArgs struct {
	ID       string   `json:"id" jsonschema:"required,description=User ID,pattern=^[0-9]+$"`
	Fields   []string `json:"fields,omitempty" jsonschema:"description=Fields to return"`
	Limit    int      `json:"limit,omitempty" jsonschema:"description=Maximum number of results,minimum=1,maximum=100"`
	After    string   `json:"after,omitempty" jsonschema:"description=Cursor for pagination (next page)"`
	Before   string   `json:"before,omitempty" jsonschema:"description=Cursor for pagination (previous page)"`
	Folder   string   `json:"folder,omitempty" jsonschema:"description=Folder"`
	Platform string   `json:"platform,omitempty" jsonschema:"description=Platform"`
	Tags     []string `json:"tags,omitempty" jsonschema:"description=Tags"`
	UserId   string   `json:"user_id,omitempty" jsonschema:"description=ID of the User,pattern=^[0-9]+$"`
}

// RegisterUserListConversationsHandler registers the user_list_conversations tool
func RegisterUserListConversationsHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"user_list_conversations",
		"List conversations for this User Returns UnifiedThread.",
		json.RawMessage(`{"additionalProperties":false,"properties":{"after":{"description":"Cursor for pagination (next page)","type":"string"},"before":{"description":"Cursor for pagination (previous page)","type":"string"},"fields":{"description":"Fields to return","items":{"type":"string"},"type":"array"},"folder":{"description":"Folder","type":"string"},"id":{"description":"User ID","pattern":"^[0-9]+$","type":"string"},"limit":{"description":"Maximum number of results","maximum":100,"minimum":1,"type":"integer"},"platform":{"description":"Platform (enum: userconversations_platform_enum_param)","enum":["INSTAGRAM","MESSENGER"],"type":"string"},"tags":{"description":"Tags","items":{"type":"string"},"type":"array"},"user_id":{"description":"ID of the User","pattern":"^[0-9]+$","type":"string"}},"required":["id"],"type":"object"}`),
	)

	s.AddTool(tool, UserListConversationsHandler)
	return nil
}

// UserListConversationsHandler handles the user_list_conversations tool
func UserListConversationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args user_list_conversationsArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s/conversations", args.ID)
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
	if args.Folder != "" {
		params["folder"] = args.Folder
	}
	if args.Platform != "" {
		params["platform"] = args.Platform
	}
	if len(args.Tags) > 0 {
		params["tags"] = strings.Join(args.Tags, ",")
	}
	if args.UserId != "" {
		params["user_id"] = args.UserId
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

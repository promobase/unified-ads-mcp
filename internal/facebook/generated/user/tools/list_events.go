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

// user_list_eventsArgs defines the typed arguments for user_list_events
type user_list_eventsArgs struct {
	ID              string   `json:"id" jsonschema:"required,description=User ID,pattern=^[0-9]+$"`
	Fields          []string `json:"fields,omitempty" jsonschema:"description=Fields to return"`
	Limit           int      `json:"limit,omitempty" jsonschema:"description=Maximum number of results,minimum=1,maximum=100"`
	After           string   `json:"after,omitempty" jsonschema:"description=Cursor for pagination (next page)"`
	Before          string   `json:"before,omitempty" jsonschema:"description=Cursor for pagination (previous page)"`
	IncludeCanceled bool     `json:"include_canceled,omitempty" jsonschema:"description=Include Canceled"`
	Type            string   `json:"type,omitempty" jsonschema:"description=Type"`
}

// RegisterUserListEventsHandler registers the user_list_events tool
func RegisterUserListEventsHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"user_list_events",
		"List events for this User Returns Event.",
		json.RawMessage(`{"additionalProperties":false,"properties":{"after":{"description":"Cursor for pagination (next page)","type":"string"},"before":{"description":"Cursor for pagination (previous page)","type":"string"},"fields":{"description":"Fields to return","items":{"type":"string"},"type":"array"},"id":{"description":"User ID","pattern":"^[0-9]+$","type":"string"},"include_canceled":{"description":"Include Canceled","type":"boolean"},"limit":{"description":"Maximum number of results","maximum":100,"minimum":1,"type":"integer"},"type":{"description":"Type (enum: userevents_type_enum_param)","enum":["attending","created","declined","maybe","not_replied"],"type":"string"}},"required":["id"],"type":"object"}`),
	)

	s.AddTool(tool, UserListEventsHandler)
	return nil
}

// UserListEventsHandler handles the user_list_events tool
func UserListEventsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args user_list_eventsArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s/events", args.ID)
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
	params["include_canceled"] = fmt.Sprintf("%v", args.IncludeCanceled)
	if args.Type != "" {
		params["type"] = args.Type
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

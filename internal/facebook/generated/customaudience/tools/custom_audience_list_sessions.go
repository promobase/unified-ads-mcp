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

// custom_audience_list_sessionsArgs defines the typed arguments for custom_audience_list_sessions
type custom_audience_list_sessionsArgs struct {
	ID        string   `json:"id" jsonschema:"required,description=CustomAudience ID,pattern=^[0-9]+$"`
	Fields    []string `json:"fields,omitempty" jsonschema:"description=Fields to return"`
	Limit     int      `json:"limit,omitempty" jsonschema:"description=Maximum number of results,minimum=1,maximum=100"`
	After     string   `json:"after,omitempty" jsonschema:"description=Cursor for pagination (next page)"`
	Before    string   `json:"before,omitempty" jsonschema:"description=Cursor for pagination (previous page)"`
	SessionId int      `json:"session_id,omitempty" jsonschema:"description=ID of the Session,pattern=^[0-9]+$"`
}

// RegisterCustomAudienceListSessionsHandler registers the custom_audience_list_sessions tool
func RegisterCustomAudienceListSessionsHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"custom_audience_list_sessions",
		"List sessions for this CustomAudience Returns CustomAudienceSession.",
		json.RawMessage(`{"additionalProperties":false,"properties":{"after":{"description":"Cursor for pagination (next page)","type":"string"},"before":{"description":"Cursor for pagination (previous page)","type":"string"},"fields":{"description":"Fields to return","items":{"type":"string"},"type":"array"},"id":{"description":"CustomAudience ID","pattern":"^[0-9]+$","type":"string"},"limit":{"description":"Maximum number of results","maximum":100,"minimum":1,"type":"integer"},"session_id":{"description":"ID of the Session","pattern":"^[0-9]+$","type":"integer"}},"required":["id"],"type":"object"}`),
	)

	s.AddTool(tool, CustomAudienceListSessionsHandler)
	return nil
}

// CustomAudienceListSessionsHandler handles the custom_audience_list_sessions tool
func CustomAudienceListSessionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args custom_audience_list_sessionsArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s/sessions", args.ID)
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
	if args.SessionId > 0 {
		params["session_id"] = fmt.Sprintf("%d", args.SessionId)
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

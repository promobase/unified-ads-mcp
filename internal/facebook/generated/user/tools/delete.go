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

// user_deleteArgs defines the typed arguments for user_delete
type user_deleteArgs struct {
	ID string `json:"id" jsonschema:"required,description=User ID,pattern=^[0-9]+$"`
}

// RegisterUserDeleteHandler registers the user_delete tool
func RegisterUserDeleteHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"user_delete",
		"Delete a User",
		json.RawMessage(`{"additionalProperties":false,"properties":{"id":{"description":"User ID","pattern":"^[0-9]+$","type":"string"}},"required":["id"],"type":"object"}`),
	)

	s.AddTool(tool, UserDeleteHandler)
	return nil
}

// UserDeleteHandler handles the user_delete tool
func UserDeleteHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args user_deleteArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s", args.ID)
	// Prepare query parameters
	params := make(map[string]string)
	// ID is part of path, not query params

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

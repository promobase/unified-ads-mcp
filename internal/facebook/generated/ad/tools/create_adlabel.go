// Code generated by codegen. DO NOT EDIT.

package tools

import (
	"context"
	"encoding/json"

	"unified-ads-mcp/internal/facebook/generated/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ad_create_adlabelArgs defines the typed arguments for ad_create_adlabel
type ad_create_adlabelArgs struct {
	Adlabels         []*common.AdLabel `json:"adlabels" jsonschema:"description=Adlabels,required"`
	ExecutionOptions []string          `json:"execution_options,omitempty" jsonschema:"description=Execution Options"`
}

// RegisterAdCreateAdlabelHandler registers the ad_create_adlabel tool
func RegisterAdCreateAdlabelHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"ad_create_adlabel",
		"Associate adlabels with this Ad Returns Ad. Required: adlabels",
		json.RawMessage(`{"additionalProperties":false,"properties":{"adlabels":{"description":"Adlabels","items":{"additionalProperties":true,"type":"object"},"type":"array"},"execution_options":{"description":"Execution Options","items":{"type":"string"},"type":"array"}},"required":["adlabels"],"type":"object"}`),
	)

	s.AddTool(tool, AdCreateAdlabelHandler)
	return nil
}

// AdCreateAdlabelHandler handles the ad_create_adlabel tool
func AdCreateAdlabelHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args ad_create_adlabelArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := "/adlabels"
	// Prepare request body
	body := make(map[string]interface{})
	if len(args.Adlabels) > 0 {
		body["adlabels"] = args.Adlabels
	}
	if len(args.ExecutionOptions) > 0 {
		body["execution_options"] = args.ExecutionOptions
	}

	result, err := common.MakeGraphAPIRequest(ctx, "POST", endpoint, nil, body)

	if err != nil {
		return common.HandleAPIError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(result)),
		},
	}, nil
}

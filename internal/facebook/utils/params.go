package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ParseRequiredString extracts a required string parameter
func ParseRequiredString(request mcp.CallToolRequest, paramName string, args map[string]interface{}) error {
	value, err := request.RequireString(paramName)
	if err != nil {
		return fmt.Errorf("missing required parameter %s: %w", paramName, err)
	}
	args[paramName] = value
	return nil
}

// ParseOptionalString extracts an optional string parameter
func ParseOptionalString(request mcp.CallToolRequest, paramName string, args map[string]interface{}) {
	if val := request.GetString(paramName, ""); val != "" {
		args[paramName] = val
	}
}

// ParseOptionalInt extracts an optional integer parameter
func ParseOptionalInt(request mcp.CallToolRequest, paramName string, args map[string]interface{}) {
	if val := request.GetInt(paramName, 0); val != 0 {
		args[paramName] = val
	}
}

// ParseOptionalFloat extracts an optional float parameter
func ParseOptionalFloat(request mcp.CallToolRequest, paramName string, args map[string]interface{}) {
	if val := request.GetFloat(paramName, 0); val != 0 {
		args[paramName] = val
	}
}

// ParseOptionalBool extracts an optional boolean parameter
func ParseOptionalBool(request mcp.CallToolRequest, paramName string, args map[string]interface{}) {
	if val := request.GetBool(paramName, false); val {
		args[paramName] = val
	}
}

// ParseFieldsArray parses a fields array parameter and converts to comma-separated string
func ParseFieldsArray(request mcp.CallToolRequest, args map[string]interface{}) {
	if val := request.GetString("fields", ""); val != "" {
		// Try to parse as JSON array first
		var fields []string
		if err := json.Unmarshal([]byte(val), &fields); err == nil && len(fields) > 0 {
			args["fields"] = strings.Join(fields, ",")
		} else {
			// If not JSON, assume it's already a string (could be comma-separated)
			args["fields"] = val
		}
	}
}

// ParseParamsObject parses a params object and expands its properties into args
func ParseParamsObject(request mcp.CallToolRequest, args map[string]interface{}) {
	if val := request.GetString("params", ""); val != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(val), &params); err == nil {
			for key, value := range params {
				args[key] = value
			}
		}
	}
}

// ParseOptionalArray parses an optional array parameter
func ParseOptionalArray(request mcp.CallToolRequest, paramName string, args map[string]interface{}) {
	if val := request.GetString(paramName, ""); val != "" {
		var arr []interface{}
		if err := json.Unmarshal([]byte(val), &arr); err == nil {
			args[paramName] = arr
		}
	}
}

// ParseOptionalObject parses an optional object parameter
func ParseOptionalObject(request mcp.CallToolRequest, paramName string, args map[string]interface{}) {
	if val := request.GetString(paramName, ""); val != "" {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(val), &obj); err == nil {
			args[paramName] = obj
		}
	}
}

// ParseRequiredObject parses a required object parameter
func ParseRequiredObject(request mcp.CallToolRequest, paramName string, args map[string]interface{}) error {
	val, err := request.RequireString(paramName)
	if err != nil {
		return fmt.Errorf("missing required parameter %s: %w", paramName, err)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(val), &obj); err != nil {
		return fmt.Errorf("invalid %s object: %w", paramName, err)
	}

	args[paramName] = obj
	return nil
}

// ParseRequiredArray parses a required array parameter
func ParseRequiredArray(request mcp.CallToolRequest, paramName string, args map[string]interface{}) error {
	val, err := request.RequireString(paramName)
	if err != nil {
		return fmt.Errorf("missing required parameter %s: %w", paramName, err)
	}

	var arr []interface{}
	if err := json.Unmarshal([]byte(val), &arr); err != nil {
		return fmt.Errorf("invalid %s array: %w", paramName, err)
	}

	args[paramName] = arr
	return nil
}

// ParseRequiredFieldsArray parses a required fields array and converts to comma-separated string
func ParseRequiredFieldsArray(request mcp.CallToolRequest, args map[string]interface{}) error {
	val, err := request.RequireString("fields")
	if err != nil {
		return fmt.Errorf("missing required parameter fields: %w", err)
	}

	var fields []string
	if err := json.Unmarshal([]byte(val), &fields); err != nil {
		return fmt.Errorf("invalid fields array: %w", err)
	}

	args["fields"] = strings.Join(fields, ",")
	return nil
}

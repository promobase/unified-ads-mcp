package generators

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/mark3labs/mcp-go/mcp"
)

// SchemaGenerator generates JSON schemas from Go structs and converts them to MCP tool definitions
type SchemaGenerator struct {
	reflector *jsonschema.Reflector
	enumTypes map[string][]string // enum type name -> possible values
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator() *SchemaGenerator {
	reflector := &jsonschema.Reflector{
		// Expand references inline for easier processing
		ExpandedStruct: true,
		// Use JSON field names
		KeyNamer: func(s string) string {
			return s
		},
	}

	return &SchemaGenerator{
		reflector: reflector,
		enumTypes: make(map[string][]string),
	}
}

// LoadEnumDefinitions loads enum definitions from the enum_types.json
func (g *SchemaGenerator) LoadEnumDefinitions(enumTypes map[string][]string) {
	g.enumTypes = enumTypes
}

// GenerateSchemaFromStruct generates a JSON schema from a Go struct
func (g *SchemaGenerator) GenerateSchemaFromStruct(structType reflect.Type) (*jsonschema.Schema, error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", structType.Kind())
	}

	// Create instance for reflection
	instance := reflect.New(structType).Interface()
	schema := g.reflector.Reflect(instance)

	// Post-process to handle Facebook-specific types
	g.postProcessSchema(schema, structType)

	return schema, nil
}

// GenerateSchemaFromInterface generates a JSON schema from an interface{} value
func (g *SchemaGenerator) GenerateSchemaFromInterface(v interface{}) (*jsonschema.Schema, error) {
	schema := g.reflector.Reflect(v)

	// Post-process to handle Facebook-specific types
	if structType := reflect.TypeOf(v); structType != nil {
		if structType.Kind() == reflect.Ptr {
			structType = structType.Elem()
		}
		g.postProcessSchema(schema, structType)
	}

	return schema, nil
}

// postProcessSchema handles Facebook-specific type conversions
func (g *SchemaGenerator) postProcessSchema(schema *jsonschema.Schema, structType reflect.Type) {
	if schema.Properties == nil {
		return
	}

	// Process each property
	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		propName := pair.Key
		propSchema := pair.Value

		// Find the struct field
		field, found := structType.FieldByName(g.toCamelCase(propName))
		if !found {
			// Try exact match
			for i := 0; i < structType.NumField(); i++ {
				f := structType.Field(i)
				jsonTag := f.Tag.Get("json")
				if jsonTag == propName || strings.Split(jsonTag, ",")[0] == propName {
					field = f
					found = true
					break
				}
			}
		}

		if found {
			// Check if this is an enum type
			if enumValues, ok := g.enumTypes[field.Type.Name()]; ok {
				propSchema.Enum = make([]interface{}, len(enumValues))
				for i, v := range enumValues {
					propSchema.Enum[i] = v
				}
			}

			// Handle Facebook ID types
			if strings.Contains(field.Type.Name(), "ID") || strings.HasSuffix(propName, "_id") {
				propSchema.Pattern = "^[0-9]+$"
				if propSchema.Description == "" {
					propSchema.Description = "Facebook numeric ID"
				}
			}
		}
	}
}

// ConvertSchemaToMCPParams converts a JSON schema to MCP tool parameters
func (g *SchemaGenerator) ConvertSchemaToMCPParams(schema *jsonschema.Schema) []mcp.ToolOption {
	var params []mcp.ToolOption

	// Get required fields
	requiredMap := make(map[string]bool)
	for _, req := range schema.Required {
		requiredMap[req] = true
	}

	// Process each property
	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		propName := pair.Key
		propSchema := pair.Value

		param := g.createMCPParam(propName, propSchema, requiredMap[propName])
		if param != nil {
			params = append(params, param)
		}
	}

	return params
}

// createMCPParam creates an MCP parameter from a schema property
func (g *SchemaGenerator) createMCPParam(name string, schema *jsonschema.Schema, required bool) mcp.ToolOption {
	var options []mcp.PropertyOption

	// Add description if present
	if schema.Description != "" {
		options = append(options, mcp.Description(schema.Description))
	}

	// Add required if true
	if required {
		options = append(options, mcp.Required())
	}

	// Handle references - treat them as objects
	if schema.Ref != "" {
		// This is a reference to another schema, which means it's a complex object
		return mcp.WithObject(name, options...)
	}

	// Handle different schema types
	switch schema.Type {
	case "string":
		// Check for enums
		if len(schema.Enum) > 0 {
			enumValues := make([]string, len(schema.Enum))
			for i, v := range schema.Enum {
				enumValues[i] = fmt.Sprintf("%v", v)
			}
			enumDesc := fmt.Sprintf("%s (enum: %s)", schema.Description, strings.Join(enumValues, ", "))
			if len(options) > 0 {
				options[0] = mcp.Description(enumDesc)
			} else {
				options = append(options, mcp.Description(enumDesc))
			}
		}
		return mcp.WithString(name, options...)

	case "number", "integer":
		return mcp.WithNumber(name, options...)

	case "boolean":
		return mcp.WithBoolean(name, options...)

	case "array":
		// For arrays, we just use WithArray
		return mcp.WithArray(name, options...)

	case "object":
		// For complex objects
		return mcp.WithObject(name, options...)

	default:
		// Default to string for unknown types
		return mcp.WithString(name, options...)
	}
}

// GenerateToolFromSchema creates an MCP tool from a struct with schema validation
func (g *SchemaGenerator) GenerateToolFromSchema(toolName string, description string, handlerStruct interface{}) (mcp.Tool, error) {
	// Generate schema from struct
	schema, err := g.GenerateSchemaFromInterface(handlerStruct)
	if err != nil {
		return mcp.Tool{}, fmt.Errorf("failed to generate schema: %w", err)
	}

	// Get tool options from schema
	options := []mcp.ToolOption{mcp.WithDescription(description)}
	options = append(options, g.ConvertSchemaToMCPParams(schema)...)

	// Create tool with all options
	tool := mcp.NewTool(toolName, options...)

	return tool, nil
}

// Helper function to convert snake_case to CamelCase
func (g *SchemaGenerator) toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// ValidateAgainstSchema validates data against a generated schema
func (g *SchemaGenerator) ValidateAgainstSchema(schema *jsonschema.Schema, data interface{}) error {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Unmarshal to generic interface
	var genericData interface{}
	if err := json.Unmarshal(jsonData, &genericData); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Validate against schema
	// Note: This is a simplified validation. In production, you'd use a full JSON Schema validator
	return g.basicValidation(schema, genericData)
}

// basicValidation performs basic validation against schema
func (g *SchemaGenerator) basicValidation(schema *jsonschema.Schema, data interface{}) error {
	// Check required fields
	if dataMap, ok := data.(map[string]interface{}); ok {
		for _, required := range schema.Required {
			if _, exists := dataMap[required]; !exists {
				return fmt.Errorf("required field '%s' is missing", required)
			}
		}
	}

	// Additional validation can be added here
	return nil
}

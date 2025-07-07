package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	// APIVersion is the Facebook Graph API version to use
	APIVersion = "v23.0"
)

// templateFS embeds all template files from the templates directory
// Templates are used to generate Go code files for the Facebook Business API
//go:embed templates/*
var templateFS embed.FS

// APISpec represents the structure of each JSON spec file
type APISpec struct {
	APIs   []APIEndpoint `json:"apis"`
	Fields []Field       `json:"fields"`
}

// APIEndpoint represents an API endpoint definition
type APIEndpoint struct {
	Method   string      `json:"method"`
	Endpoint string      `json:"endpoint"`
	Return   string      `json:"return"`
	Params   []Parameter `json:"params"`
}

// Parameter represents a parameter for an API endpoint
type Parameter struct {
	Name        string   `json:"name"`
	Required    bool     `json:"required"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	EnumValues  []string `json:"enum_values,omitempty"`
}

// Field represents a field definition
type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// EnumType represents an enum definition
type EnumType struct {
	Name         string   `json:"name"`
	Node         string   `json:"node"`
	FieldOrParam string   `json:"field_or_param"`
	Values       []string `json:"values"`
}

// CodegenContext holds all the data needed for code generation
type CodegenContext struct {
	Specs map[string]*APISpec
	Enums []EnumType
	Tools []MCPTool
	Types map[string]*TypeDefinition
}

// MCPTool represents a generated MCP tool
type MCPTool struct {
	Name            string
	Description     string
	Method          string
	Endpoint        string
	Return          string
	Parameters      []MCPParameter
	NodeType        string
	AvailableFields []string     // For GET endpoints, the fields that can be requested
	APIParams       []Parameter  // Original API parameters for params object generation
}

// MCPParameter represents a parameter for an MCP tool
type MCPParameter struct {
	Name        string
	Type        string
	Required    bool
	Description string
	EnumValues  []string
	APIParams   []Parameter // For params object, holds the nested parameters
}

// TypeDefinition represents a Go type definition
type TypeDefinition struct {
	Name   string
	Fields []TypeField
}

// TypeField represents a field in a Go struct
type TypeField struct {
	Name     string
	Type     string
	JSONTag  string
	Optional bool
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <specs_directory>")
	}

	specsDir := os.Args[1]

	ctx, err := loadSpecs(specsDir)
	if err != nil {
		log.Fatalf("Failed to load specs: %v", err)
	}

	if err := generateCode(ctx); err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	fmt.Println("Code generation completed successfully!")
}

// loadSpecs loads all API specs and enums from the specs directory
func loadSpecs(specsDir string) (*CodegenContext, error) {
	ctx := &CodegenContext{
		Specs: make(map[string]*APISpec),
		Types: make(map[string]*TypeDefinition),
	}

	// Load enum types first
	enumPath := filepath.Join(specsDir, "enum_types.json")
	if err := loadEnums(ctx, enumPath); err != nil {
		return nil, fmt.Errorf("failed to load enums: %w", err)
	}

	// Load all spec files
	err := filepath.WalkDir(specsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") || strings.HasSuffix(path, "enum_types.json") {
			return nil
		}

		return loadSpecFile(ctx, path)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk specs directory: %w", err)
	}

	// Generate tools and types
	generateTools(ctx)
	generateTypes(ctx)

	return ctx, nil
}

// loadEnums loads enum definitions from enum_types.json
func loadEnums(ctx *CodegenContext, enumPath string) error {
	data, err := os.ReadFile(enumPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &ctx.Enums)
}

// loadSpecFile loads a single API spec file
func loadSpecFile(ctx *CodegenContext, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var spec APISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Extract node name from filename
	filename := filepath.Base(path)
	nodeName := strings.TrimSuffix(filename, ".json")

	ctx.Specs[nodeName] = &spec
	return nil
}

// generateTools creates MCP tools from API endpoints
func generateTools(ctx *CodegenContext) {
	for nodeName, spec := range ctx.Specs {
		for _, api := range spec.APIs {
			tool := MCPTool{
				Name:        fmt.Sprintf("%s_%s_%s", strings.ToLower(nodeName), strings.ToLower(api.Method), normalizeEndpoint(api.Endpoint)),
				Description: fmt.Sprintf("%s %s for %s", api.Method, api.Endpoint, nodeName),
				Method:      api.Method,
				Endpoint:    api.Endpoint,
				Return:      api.Return,
				NodeType:    nodeName,
			}

			// Add required ID parameters based on node type
			switch nodeName {
			case "AdAccount":
				accountIdParam := MCPParameter{
					Name:        "account_id",
					Type:        "string",
					Required:    true,
					Description: "Facebook Ad Account ID (without 'act_' prefix)",
				}
				tool.Parameters = append(tool.Parameters, accountIdParam)
			case "Ad":
				adIdParam := MCPParameter{
					Name:        "ad_id",
					Type:        "string",
					Required:    true,
					Description: "Facebook Ad ID",
				}
				tool.Parameters = append(tool.Parameters, adIdParam)
			case "AdCreative":
				adCreativeIdParam := MCPParameter{
					Name:        "ad_creative_id",
					Type:        "string",
					Required:    true,
					Description: "Facebook Ad Creative ID",
				}
				tool.Parameters = append(tool.Parameters, adCreativeIdParam)
			case "AdSet":
				adSetIdParam := MCPParameter{
					Name:        "ad_set_id",
					Type:        "string",
					Required:    true,
					Description: "Facebook Ad Set ID",
				}
				tool.Parameters = append(tool.Parameters, adSetIdParam)
			case "Campaign":
				campaignIdParam := MCPParameter{
					Name:        "campaign_id",
					Type:        "string",
					Required:    true,
					Description: "Facebook Campaign ID",
				}
				tool.Parameters = append(tool.Parameters, campaignIdParam)
			}

			// For endpoints with parameters, create a params object
			if len(api.Params) > 0 {
				// Build detailed description of params object
				paramDescriptions := []string{}
				requiredParams := []string{}
				
				// Enhance API params with descriptions and enum values
				enhancedParams := make([]Parameter, len(api.Params))
				for i, param := range api.Params {
					enhancedParams[i] = param
					
					// Build description for display
					paramDesc := fmt.Sprintf("%s (%s)", param.Name, convertTypeForDisplay(param.Type))
					
					// Add enum values if available
					if enumValues := findEnumValues(ctx, param.Type); len(enumValues) > 0 {
						enhancedParams[i].EnumValues = enumValues
						if len(enumValues) <= 5 {
							paramDesc += fmt.Sprintf(" [%s]", strings.Join(enumValues, ", "))
						} else {
							paramDesc += fmt.Sprintf(" [%s, ...]", strings.Join(enumValues[:5], ", "))
						}
					}
					
					// Set description
					enhancedParams[i].Description = fmt.Sprintf("%s parameter", param.Name)
					
					if param.Required {
						paramDesc += " [required]"
						requiredParams = append(requiredParams, param.Name)
					}
					paramDescriptions = append(paramDescriptions, paramDesc)
				}
				
				description := fmt.Sprintf("Parameters object containing: %s", strings.Join(paramDescriptions, ", "))
				
				// Create a params object that contains all API-specific parameters
				paramsParam := MCPParameter{
					Name:        "params",
					Type:        "object",
					Required:    len(requiredParams) > 0, // Required if any param is required
					Description: description,
					APIParams:   enhancedParams,
				}
				tool.Parameters = append(tool.Parameters, paramsParam)
				
				
				// Store the API params for later use in handler generation
				tool.APIParams = enhancedParams
			}

			// Add common parameters for GET endpoints
			if api.Method == "GET" {
				// Add fields parameter as an array
				fieldsParam := MCPParameter{
					Name:        "fields",
					Type:        "array",
					Required:    false,
					Description: fmt.Sprintf("Array of fields to return for %s objects", api.Return),
				}
				
				// Get available fields from the return type spec
				// For edge endpoints, api.Return contains the type being returned
				// For non-edge endpoints (like GET /{id}), we might need to use the node type itself
				var targetType string
				if api.Return != "" && api.Return != "Object" {
					targetType = api.Return
				} else if api.Endpoint == "" || api.Endpoint == "/" {
					// Non-edge endpoint, use the node type itself
					targetType = nodeName
				} else {
					// Default to the return type
					targetType = api.Return
				}
				
				// Look up the fields from the target type's spec
				if targetSpec, exists := ctx.Specs[targetType]; exists && len(targetSpec.Fields) > 0 {
					var fieldNames []string
					for _, field := range targetSpec.Fields {
						fieldNames = append(fieldNames, field.Name)
					}
					if len(fieldNames) > 0 {
						// Store available fields in the tool for type-safe generation
						tool.AvailableFields = fieldNames
						
						// Add field names to description for reference (show first 15 fields)
						displayFields := fieldNames
						if len(fieldNames) > 15 {
							displayFields = fieldNames[:15]
						}
						fieldsParam.Description = fmt.Sprintf("Array of fields to return for %s objects. Available fields: %s", 
							targetType, strings.Join(displayFields, ", "))
						if len(fieldNames) > 15 {
							fieldsParam.Description += fmt.Sprintf(" (and %d more)", len(fieldNames)-15)
						}
						
						tool.Parameters = append(tool.Parameters, fieldsParam)
					}
				} else {
					// If no fields found, still add the parameter but with a generic description
					fieldsParam.Description = "Array of fields to return"
					tool.Parameters = append(tool.Parameters, fieldsParam)
				}

				// Add pagination parameters
				limitParam := MCPParameter{
					Name:        "limit",
					Type:        "integer",
					Required:    false,
					Description: "Maximum number of results to return (default: 25, max: 500)",
				}
				tool.Parameters = append(tool.Parameters, limitParam)

				afterParam := MCPParameter{
					Name:        "after",
					Type:        "string",
					Required:    false,
					Description: "Cursor for pagination (use 'next' cursor from previous response)",
				}
				tool.Parameters = append(tool.Parameters, afterParam)

				beforeParam := MCPParameter{
					Name:        "before",
					Type:        "string",
					Required:    false,
					Description: "Cursor for pagination (use 'previous' cursor from previous response)",
				}
				tool.Parameters = append(tool.Parameters, beforeParam)
			}

			ctx.Tools = append(ctx.Tools, tool)
		}
	}
}

// generateTypes creates Go type definitions from field specs
func generateTypes(ctx *CodegenContext) {
	for nodeName, spec := range ctx.Specs {
		typeDef := &TypeDefinition{
			Name: nodeName,
		}

		for _, field := range spec.Fields {
			fieldType := convertType(field.Type)

			// Use pointers for potential recursive types (Facebook object types)
			if len(field.Type) > 0 && field.Type[0] >= 'A' && field.Type[0] <= 'Z' &&
				!strings.Contains(field.Type, "_") && field.Type != "Object" {
				fieldType = "*" + fieldType
			}

			typeField := TypeField{
				Name:    sanitizeFieldName(field.Name),
				Type:    fieldType,
				JSONTag: field.Name,
			}

			typeDef.Fields = append(typeDef.Fields, typeField)
		}

		ctx.Types[nodeName] = typeDef
	}
}

// Helper functions

func normalizeEndpoint(endpoint string) string {
	return strings.ReplaceAll(endpoint, "/", "_")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func convertType(fbType string) string {
	// Handle list types
	if strings.HasPrefix(fbType, "list<") && strings.HasSuffix(fbType, ">") {
		innerType := fbType[5 : len(fbType)-1]
		return "[]" + convertType(innerType)
	}

	// Handle basic types
	switch fbType {
	case "string":
		return "string"
	case "int", "unsigned int":
		return "int"
	case "float":
		return "float64"
	case "bool":
		return "bool"
	case "datetime":
		return "time.Time"
	case "Object", "map":
		return "map[string]interface{}"
	case "list":
		return "[]interface{}"
	default:
		// Check if it's an enum type or contains underscores (likely enum)
		if strings.Contains(fbType, "_") {
			return "string"
		}
		// For types that contain angle brackets, convert to interface{}
		if strings.Contains(fbType, "<") || strings.Contains(fbType, ">") {
			return "interface{}"
		}
		// Check if it's a known Facebook type (starts with capital letter)
		if len(fbType) > 0 && fbType[0] >= 'A' && fbType[0] <= 'Z' {
			// It's likely a custom Facebook type, keep as is
			return fbType
		}
		// Default to interface{} for unknown types
		return "interface{}"
	}
}
func convertTypeForMCP(fbType string) string {
	// Handle list types for MCP schema
	if strings.HasPrefix(fbType, "list<") && strings.HasSuffix(fbType, ">") {
		return "array"
	}

	// Handle basic types for MCP
	switch fbType {
	case "string":
		return "string"
	case "int", "unsigned int":
		return "integer"
	case "float":
		return "number"
	case "bool":
		return "boolean"
	case "datetime":
		return "string"
	case "Object":
		return "object"
	default:
		// Check if it's an enum type
		if strings.Contains(fbType, "_enum_param") {
			return "string"
		}
		// Default to string for unknown types
		return "string"
	}
}

func findEnumValues(ctx *CodegenContext, paramType string) []string {
	for _, enum := range ctx.Enums {
		if strings.Contains(paramType, enum.Name) {
			return enum.Values
		}
	}
	return nil
}

func convertTypeForDisplay(fbType string) string {
	// Handle list types
	if strings.HasPrefix(fbType, "list<") && strings.HasSuffix(fbType, ">") {
		innerType := fbType[5 : len(fbType)-1]
		return "array<" + convertTypeForDisplay(innerType) + ">"
	}

	// Handle basic types
	switch fbType {
	case "string":
		return "string"
	case "int", "unsigned int":
		return "integer"
	case "float":
		return "number"
	case "bool":
		return "boolean"
	case "datetime":
		return "datetime"
	case "Object":
		return "object"
	case "map":
		return "object"
	default:
		// Check if it's an enum type
		if strings.Contains(fbType, "_enum_param") {
			// Extract the base name for display
			parts := strings.Split(fbType, "_")
			if len(parts) > 0 {
				return "enum"
			}
		}
		// Keep original type for display
		return fbType
	}
}

func convertTypeToJSONSchema(fbType string) string {
	// Handle list types
	if strings.HasPrefix(fbType, "list<") && strings.HasSuffix(fbType, ">") {
		return "array"
	}

	// Handle basic types
	switch fbType {
	case "string":
		return "string"
	case "int", "unsigned int":
		return "integer"
	case "float":
		return "number"
	case "bool":
		return "boolean"
	case "datetime":
		return "string" // datetime represented as string in JSON Schema
	case "Object", "map":
		return "object"
	default:
		// Check if it's an enum type
		if strings.Contains(fbType, "_enum_param") {
			return "string"
		}
		// Default to string for unknown types
		return "string"
	}
}

func extractItemType(listType string) string {
	// Extract inner type from list<Type> format
	if strings.HasPrefix(listType, "list<") && strings.HasSuffix(listType, ">") {
		innerType := listType[5 : len(listType)-1]
		return convertTypeToJSONSchema(innerType)
	}
	return "string"
}

// getTemplateFuncs returns common template functions
func getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"capitalizeFirst": capitalizeFirst,
		"sanitizeVarName": sanitizeVarName,
		"sanitizeEnumValue": sanitizeEnumValue,
		"convertTypeToJSONSchema": convertTypeToJSONSchema,
		"extractItemType": extractItemType,
		"hasPrefix": strings.HasPrefix,
		"paramType": func(t string) string {
			switch t {
			case "string":
				return "String"
			case "integer":
				return "Number"
			case "number":
				return "Number"
			case "boolean":
				return "Boolean"
			case "array":
				return "Array"
			case "object":
				return "Object"
			default:
				return "String"
			}
		},
		"requireMethod": func(t string) string {
			switch t {
			case "string":
				return "RequireString"
			case "integer":
				return "RequireInt"
			case "number":
				return "RequireFloat"
			case "boolean":
				return "RequireBool"
			case "array":
				return "RequireString" // Arrays passed as JSON strings
			case "object":
				return "RequireString" // Objects passed as JSON strings
			default:
				return "RequireString"
			}
		},
		"varName": func(name string) string {
			return sanitizeVarName(name)
		},
	}
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func sanitizeFieldName(name string) string {
	// Replace dots and other invalid characters with underscores
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")

	// Handle Go keywords and reserved words
	goKeywords := map[string]string{
		"type":      "Type_",
		"func":      "Func_",
		"var":       "Var_",
		"const":     "Const_",
		"package":   "Package_",
		"import":    "Import_",
		"interface": "Interface_",
		"struct":    "Struct_",
		"map":       "Map_",
		"chan":      "Chan_",
		"select":    "Select_",
		"case":      "Case_",
		"default":   "Default_",
		"if":        "If_",
		"else":      "Else_",
		"for":       "For_",
		"range":     "Range_",
		"switch":    "Switch_",
		"return":    "Return_",
		"break":     "Break_",
		"continue":  "Continue_",
		"goto":      "Goto_",
		"defer":     "Defer_",
		"go":        "Go_",
	}

	if replacement, exists := goKeywords[name]; exists {
		return replacement
	}

	// Handle numeric field names
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		return "Field_" + name
	}

	return capitalizeFirst(name)
}

func sanitizeVarName(name string) string {
	// Replace dots and other invalid characters with underscores
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")

	// Handle Go keywords and reserved words, plus common package names
	goKeywords := map[string]string{
		"type":      "type_",
		"func":      "func_",
		"var":       "var_",
		"const":     "const_",
		"package":   "package_",
		"import":    "import_",
		"interface": "interface_",
		"struct":    "struct_",
		"map":       "map_",
		"chan":      "chan_",
		"select":    "select_",
		"case":      "case_",
		"default":   "default_",
		"if":        "if_",
		"else":      "else_",
		"for":       "for_",
		"range":     "range_",
		"switch":    "switch_",
		"return":    "return_",
		"break":     "break_",
		"continue":  "continue_",
		"goto":      "goto_",
		"defer":     "defer_",
		"go":        "go_",
		"url":       "url_",
		"fmt":       "fmt_",
		"http":      "http_",
		"json":      "json_",
		"strings":   "strings_",
		"time":      "time_",
	}

	if replacement, exists := goKeywords[name]; exists {
		return replacement
	}

	// Handle numeric field names
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		return "field_" + name
	}

	return name
}

func sanitizeEnumValue(value string) string {
	// Replace spaces and special characters with underscores
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, ".", "_")
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, "(", "_")
	value = strings.ReplaceAll(value, ")", "_")
	value = strings.ReplaceAll(value, "+", "Plus")
	value = strings.ReplaceAll(value, "&", "And")
	value = strings.ReplaceAll(value, ",", "_")
	value = strings.ReplaceAll(value, ":", "_")
	value = strings.ReplaceAll(value, ";", "_")
	value = strings.ReplaceAll(value, "'", "_")
	value = strings.ReplaceAll(value, "\"", "_")
	value = strings.ReplaceAll(value, "?", "_")
	value = strings.ReplaceAll(value, "!", "_")
	value = strings.ReplaceAll(value, "@", "At")
	value = strings.ReplaceAll(value, "#", "Hash")
	value = strings.ReplaceAll(value, "$", "Dollar")
	value = strings.ReplaceAll(value, "%", "Percent")
	value = strings.ReplaceAll(value, "^", "_")
	value = strings.ReplaceAll(value, "*", "_")
	value = strings.ReplaceAll(value, "=", "Equals")
	value = strings.ReplaceAll(value, "<", "LessThan")
	value = strings.ReplaceAll(value, ">", "GreaterThan")
	value = strings.ReplaceAll(value, "[", "_")
	value = strings.ReplaceAll(value, "]", "_")
	value = strings.ReplaceAll(value, "{", "_")
	value = strings.ReplaceAll(value, "}", "_")
	value = strings.ReplaceAll(value, "|", "_")
	value = strings.ReplaceAll(value, "\\", "_")
	value = strings.ReplaceAll(value, "`", "_")
	value = strings.ReplaceAll(value, "~", "_")

	// Remove any consecutive underscores
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}

	// Trim leading/trailing underscores
	value = strings.Trim(value, "_")

	// If the value starts with a number, prefix with "Num_"
	if len(value) > 0 && value[0] >= '0' && value[0] <= '9' {
		value = "Num_" + value
	}

	// If empty after sanitization, use "Empty"
	if value == "" {
		value = "Empty"
	}

	return value
}

// generateCode generates the actual Go code files
func generateCode(ctx *CodegenContext) error {
	// Create generated directory structure
	generatedDir := "../generated"
	dirs := []string{
		generatedDir,
		filepath.Join(generatedDir, "types"),
		filepath.Join(generatedDir, "client"),
		filepath.Join(generatedDir, "tools"),
		filepath.Join(generatedDir, "enums"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate types (one file per type)
	if err := generateTypesFiles(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}

	// Generate client methods (grouped by node type)
	if err := generateClientFiles(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate client: %w", err)
	}

	// Generate MCP tools (split into multiple files)
	if err := generateMCPToolsFiles(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate MCP tools: %w", err)
	}

	// Generate enums
	if err := generateEnumsFile(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate enums: %w", err)
	}

	// Generate constants
	if err := generateConstantsFile(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate constants: %w", err)
	}

	// Generate index files
	if err := generateIndexFiles(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate index files: %w", err)
	}

	// Generate main tools file for backward compatibility
	if err := generateLegacyToolsFile(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate legacy tools file: %w", err)
	}

	return nil
}

func generateTypesFiles(ctx *CodegenContext, outputDir string) error {
	typesDir := filepath.Join(outputDir, "types")

	// Generate individual type files
	for typeName, typeDef := range ctx.Types {
		if err := generateSingleTypeFile(typesDir, typeName, typeDef); err != nil {
			return fmt.Errorf("failed to generate type %s: %w", typeName, err)
		}
	}

	return nil
}

func generateSingleTypeFile(typesDir string, typeName string, typeDef *TypeDefinition) error {
	// Check if any field uses time.Time
	needsTimeImport := false
	for _, field := range typeDef.Fields {
		if field.Type == "time.Time" || strings.Contains(field.Type, "time.Time") {
			needsTimeImport = true
			break
		}
	}

	tmplData, err := templateFS.ReadFile("templates/types.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read types template: %w", err)
	}

	t, err := template.New("type").Funcs(getTemplateFuncs()).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse types template: %w", err)
	}

	filename := fmt.Sprintf("%s.go", strings.ToLower(typeName))
	file, err := os.Create(filepath.Join(typesDir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a context with the NeedsTimeImport flag
	data := struct {
		*TypeDefinition
		NeedsTimeImport bool
	}{
		TypeDefinition:  typeDef,
		NeedsTimeImport: needsTimeImport,
	}

	return t.Execute(file, data)
}

// generateLegacyToolsFile generates the main tools file for backward compatibility
func generateLegacyToolsFile(ctx *CodegenContext, outputDir string) error {
	// Generate main tools file
	toolsTmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package facebook

import (
	"context"

	"unified-ads-mcp/internal/facebook/generated/tools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetMCPTools returns all available MCP tools for Facebook Business API
func GetMCPTools(accessToken string) []mcp.Tool {
	// Deprecated: accessToken parameter is ignored, use context instead
	return tools.GetAllTools()
}

// GetFilteredMCPTools returns filtered MCP tools based on enabled objects
func GetFilteredMCPTools(enabledObjects map[string]bool) []mcp.Tool {
	return tools.GetFilteredTools(enabledObjects)
}

// GetContextAwareHandlers returns handlers that get auth from context
func GetContextAwareHandlers(accessToken string) map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Deprecated: accessToken parameter is ignored, use context instead
	return tools.GetHandlers()
}

// RegisterMCPTools registers all Facebook Business API tools with the MCP server
func RegisterMCPTools(s *server.MCPServer, accessToken string) error {
	// Deprecated: accessToken parameter is ignored, use context instead
	return tools.RegisterTools(s)
}

// Legacy compatibility functions that match the old tools package signatures

// GetAllTools with accessToken parameter (for backward compatibility)
func GetAllTools(accessToken string) []mcp.Tool {
	return tools.GetAllTools()
}

// RegisterTools with accessToken parameter (for backward compatibility)  
func RegisterTools(s *server.MCPServer, accessToken string) error {
	return tools.RegisterTools(s)
}
`

	toolsFile, err := os.Create(filepath.Join("..", "generated_tools.go"))
	if err != nil {
		return err
	}
	defer toolsFile.Close()

	_, err = toolsFile.WriteString(toolsTmpl)
	return err
}

func generateClientFiles(ctx *CodegenContext, outputDir string) error {
	clientDir := filepath.Join(outputDir, "client")

	// Group tools by node type
	toolsByNode := make(map[string][]MCPTool)
	for _, tool := range ctx.Tools {
		toolsByNode[tool.NodeType] = append(toolsByNode[tool.NodeType], tool)
	}

	// Generate client method files grouped by node type
	for nodeName, tools := range toolsByNode {
		if err := generateNodeClientFile(clientDir, nodeName, tools); err != nil {
			return fmt.Errorf("failed to generate client for %s: %w", nodeName, err)
		}
	}

	return nil
}

func generateNodeClientFile(clientDir string, nodeName string, tools []MCPTool) error {
	tmplData, err := templateFS.ReadFile("templates/client.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read client template: %w", err)
	}

	t, err := template.New("client").Funcs(getTemplateFuncs()).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse client template: %w", err)
	}

	filename := fmt.Sprintf("%s_client.go", strings.ToLower(nodeName))
	file, err := os.Create(filepath.Join(clientDir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, map[string]interface{}{
		"NodeName":   nodeName,
		"Tools":      tools,
		"APIVersion": APIVersion,
	})
}

func generateMCPToolsFiles(ctx *CodegenContext, outputDir string) error {
	toolsDir := filepath.Join(outputDir, "tools")

	// Group tools by node type
	toolsByNode := make(map[string][]MCPTool)
	for _, tool := range ctx.Tools {
		toolsByNode[tool.NodeType] = append(toolsByNode[tool.NodeType], tool)
	}

	// Generate tool definition files grouped by node type
	for nodeName, tools := range toolsByNode {
		if err := generateNodeToolsFile(toolsDir, nodeName, tools, ctx); err != nil {
			return fmt.Errorf("failed to generate tools for %s: %w", nodeName, err)
		}
	}

	// Generate main tools registry file
	if err := generateToolsRegistryFile(toolsDir, ctx); err != nil {
		return fmt.Errorf("failed to generate tools registry: %w", err)
	}

	return nil
}

func generateNodeToolsFile(toolsDir string, nodeName string, tools []MCPTool, ctx *CodegenContext) error {
	tmplData, err := templateFS.ReadFile("templates/tools.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read tools template: %w", err)
	}

	t, err := template.New("tools").Funcs(getTemplateFuncs()).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse tools template: %w", err)
	}

	filename := fmt.Sprintf("%s_tools.go", strings.ToLower(nodeName))
	file, err := os.Create(filepath.Join(toolsDir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, map[string]interface{}{
		"NodeName":   nodeName,
		"Tools":      tools,
		"APIVersion": APIVersion,
	})
}

// OLD generateNodeToolsFile - remove after verification
func generateNodeToolsFileOLD(toolsDir string, nodeName string, tools []MCPTool, ctx *CodegenContext) error {
	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"unified-ads-mcp/internal/facebook/generated/client"
	"unified-ads-mcp/internal/shared"
)

// Get{{.NodeName}}Tools returns MCP tools for {{.NodeName}}
func Get{{.NodeName}}Tools() []mcp.Tool {
	var tools []mcp.Tool

{{range $tool := .Tools}}
	// {{$tool.Name}} tool{{if $tool.AvailableFields}}
	// Available fields for {{$tool.Return}}: {{range $i, $field := $tool.AvailableFields}}{{if $i}}, {{end}}{{$field}}{{end}}{{end}}{{if $tool.APIParams}}
	// Params object accepts: {{range $i, $param := $tool.APIParams}}{{if $i}}, {{end}}{{$param.Name}} ({{$param.Type}}){{end}}{{end}}
	{{sanitizeVarName $tool.Name}}Tool := mcp.NewTool("{{$tool.Name}}",
		mcp.WithDescription("{{$tool.Description}}"),
{{range $param := $tool.Parameters}}{{if eq $param.Name "params"}}{{if gt (len $param.APIParams) 0}}		mcp.WithObject("{{$param.Name}}",
{{if $param.Required}}			mcp.Required(),
{{end}}			mcp.Properties(map[string]any{
{{range $apiParam := $param.APIParams}}				"{{$apiParam.Name}}": map[string]any{
					"type": "{{convertTypeToJSONSchema $apiParam.Type}}",
					"description": "{{$apiParam.Description}}",
{{if $apiParam.Required}}					"required": true,
{{end}}{{if $apiParam.EnumValues}}					"enum": []string{ {{range $i, $val := $apiParam.EnumValues}}{{if $i}}, {{end}}"{{$val}}"{{end}} },
{{end}}{{if hasPrefix $apiParam.Type "list<"}}					"items": map[string]any{"type": "{{extractItemType $apiParam.Type}}"},
{{end}}				},
{{end}}			}),
			mcp.Description("{{$param.Description}}"),
		),
{{else}}		mcp.WithObject("{{$param.Name}}",
{{if $param.Required}}			mcp.Required(),
{{end}}			mcp.Description("{{$param.Description}}"),
		),
{{end}}{{else}}		mcp.With{{paramType $param.Type}}("{{$param.Name}}",
{{if $param.Required}}			mcp.Required(),
{{end}}			mcp.Description("{{$param.Description}}"),
{{if $param.EnumValues}}			mcp.Enum({{range $i, $val := $param.EnumValues}}{{if $i}}, {{end}}"{{$val}}"{{end}}),
{{end}}		),
{{end}}{{end}}	)
	tools = append(tools, {{sanitizeVarName $tool.Name}}Tool)
{{end}}

	return tools
}

// {{.NodeName}} handlers

{{range $tool := .Tools}}
// Handle{{capitalizeFirst $tool.Name}} handles the {{$tool.Name}} tool with context-based auth
func Handle{{capitalizeFirst $tool.Name}}(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get access token from context
	accessToken, ok := shared.FacebookAccessTokenFromContext(ctx)
	if !ok {
		return mcp.NewToolResultError("Facebook access token not found in context"), nil
	}

	// Create client
	client := client.New{{$.NodeName}}Client(accessToken)

	// Build arguments map
	args := make(map[string]interface{})

{{range $param := $tool.Parameters}}{{if $param.Required}}	// Required: {{$param.Name}}
	{{sanitizeVarName $param.Name}}, err := request.{{requireMethod $param.Type}}("{{$param.Name}}")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("missing required parameter {{$param.Name}}: %v", err)), nil
	}
	{{if eq $param.Type "object"}}{{if eq $param.Name "params"}}// Parse required params object and extract parameters
	var paramsObj map[string]interface{}
	if err := json.Unmarshal([]byte({{sanitizeVarName $param.Name}}), &paramsObj); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params object: %v", err)), nil
	}
	for key, value := range paramsObj {
		args[key] = value
	}{{else}}// Parse as JSON object
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte({{sanitizeVarName $param.Name}}), &obj); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid {{$param.Name}} object: %v", err)), nil
	}
	args["{{$param.Name}}"] = obj{{end}}{{else if eq $param.Type "array"}}{{if eq $param.Name "fields"}}// Parse required fields array
	var fields []string
	if err := json.Unmarshal([]byte({{sanitizeVarName $param.Name}}), &fields); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid fields array: %v", err)), nil
	}
	args["{{$param.Name}}"] = strings.Join(fields, ","){{else}}// Parse as JSON array
	var arr []interface{}
	if err := json.Unmarshal([]byte({{sanitizeVarName $param.Name}}), &arr); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid {{$param.Name}} array: %v", err)), nil
	}
	args["{{$param.Name}}"] = arr{{end}}{{else}}args["{{$param.Name}}"] = {{sanitizeVarName $param.Name}}{{end}}
{{else}}	// Optional: {{$param.Name}}
{{if eq $param.Type "string"}}	if val := request.GetString("{{$param.Name}}", ""); val != "" {
		args["{{$param.Name}}"] = val
	}
{{else if eq $param.Type "integer"}}	if val := request.GetInt("{{$param.Name}}", 0); val != 0 {
		args["{{$param.Name}}"] = val
	}
{{else if eq $param.Type "number"}}	if val := request.GetFloat("{{$param.Name}}", 0); val != 0 {
		args["{{$param.Name}}"] = val
	}
{{else if eq $param.Type "boolean"}}	if val := request.GetBool("{{$param.Name}}", false); val {
		args["{{$param.Name}}"] = val
	}
{{else if eq $param.Type "array"}}	// Array parameter - expecting JSON string
	if val := request.GetString("{{$param.Name}}", ""); val != "" {
		{{if eq $param.Name "fields"}}// Parse array of fields and convert to comma-separated string
		var fields []string
		if err := json.Unmarshal([]byte(val), &fields); err == nil && len(fields) > 0 {
			args["{{$param.Name}}"] = strings.Join(fields, ",")
		}{{else}}// Parse as JSON array
		var arr []interface{}
		if err := json.Unmarshal([]byte(val), &arr); err == nil {
			args["{{$param.Name}}"] = arr
		}{{end}}
	}
{{else if eq $param.Type "object"}}	// Object parameter - expecting JSON string
	if val := request.GetString("{{$param.Name}}", ""); val != "" {
		{{if eq $param.Name "params"}}// Parse params object and extract individual parameters
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(val), &params); err == nil {
			for key, value := range params {
				args[key] = value
			}
		}{{else}}// Parse as JSON object
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(val), &obj); err == nil {
			args["{{$param.Name}}"] = obj
		}{{end}}
	}
{{else}}	// {{$param.Type}} type - using string
	if val := request.GetString("{{$param.Name}}", ""); val != "" {
		args["{{$param.Name}}"] = val
	}
{{end}}{{end}}
{{end}}

	// Call the client method
	result, err := client.{{capitalizeFirst $tool.Name}}(args)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to execute {{$tool.Name}}: %v", err)), nil
	}

	// Return the result as JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

{{end}}`

	t, err := template.New("tools").Funcs(template.FuncMap{
		"capitalizeFirst": capitalizeFirst,
		"sanitizeVarName": sanitizeVarName,
		"paramType": func(t string) string {
			switch t {
			case "string":
				return "String"
			case "integer":
				return "Number"
			case "number":
				return "Number"
			case "boolean":
				return "Boolean"
			case "array":
				return "Array"
			case "object":
				return "Object"
			default:
				return "String"
			}
		},
		"requireMethod": func(t string) string {
			switch t {
			case "string":
				return "RequireString"
			case "integer":
				return "RequireInt"
			case "number":
				return "RequireFloat"
			case "boolean":
				return "RequireBool"
			case "array":
				return "RequireString" // Arrays passed as JSON strings
			case "object":
				return "RequireString" // Objects passed as JSON strings
			default:
				return "RequireString"
			}
		},
		"convertTypeToJSONSchema": convertTypeToJSONSchema,
		"extractItemType": extractItemType,
		"hasPrefix": strings.HasPrefix,
		"varName": func(name string) string {
			return sanitizeVarName(name)
		},
	}).Parse(tmpl)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s_tools.go", strings.ToLower(nodeName))
	file, err := os.Create(filepath.Join(toolsDir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, map[string]interface{}{
		"NodeName": nodeName,
		"Tools":    tools,
	})
}

func generateToolsRegistryFile(toolsDir string, ctx *CodegenContext) error {
	tmplData, err := templateFS.ReadFile("templates/tools_registry.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read tools registry template: %w", err)
	}

	// Get unique node names
	nodeTypes := []string{}
	nodeTypesMap := make(map[string]bool)
	for _, tool := range ctx.Tools {
		if !nodeTypesMap[tool.NodeType] {
			nodeTypesMap[tool.NodeType] = true
			nodeTypes = append(nodeTypes, tool.NodeType)
		}
	}

	// Prepare handler list
	handlers := []struct {
		ToolName    string
		HandlerFunc string
	}{}
	for _, tool := range ctx.Tools {
		handlers = append(handlers, struct {
			ToolName    string
			HandlerFunc string
		}{
			ToolName:    tool.Name,
			HandlerFunc: fmt.Sprintf("Handle%s", capitalizeFirst(tool.Name)),
		})
	}

	t, err := template.New("registry").Funcs(getTemplateFuncs()).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse tools registry template: %w", err)
	}

	file, err := os.Create(filepath.Join(toolsDir, "registry.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, map[string]interface{}{
		"NodeTypes": nodeTypes,
		"Handlers":  handlers,
	})
}

func generateEnumsFile(ctx *CodegenContext, outputDir string) error {
	enumDir := filepath.Join(outputDir, "enums")

	tmplData, err := templateFS.ReadFile("templates/enums.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read enums template: %w", err)
	}

	t, err := template.New("enums").Funcs(getTemplateFuncs()).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse enums template: %w", err)
	}

	file, err := os.Create(filepath.Join(enumDir, "enums.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, ctx)
}

func generateIndexFiles(ctx *CodegenContext, outputDir string) error {
	// Generate package documentation files for each subdirectory
	packages := map[string]string{
		"types":  "Package types contains generated type definitions for Facebook Business API objects.",
		"client": "Package client contains generated client methods for Facebook Business API operations.",
		"tools":  "Package tools contains generated MCP tool definitions and handlers for Facebook Business API.",
		"enums":  "Package enums contains generated enum type definitions from Facebook Business API specifications.",
	}

	for pkg, description := range packages {
		docContent := fmt.Sprintf(`// Code generated by Facebook Business API codegen. DO NOT EDIT.

// %s
package %s
`, description, pkg)

		docFile, err := os.Create(filepath.Join(outputDir, pkg, "doc.go"))
		if err != nil {
			return fmt.Errorf("failed to create doc.go for package %s: %w", pkg, err)
		}
		if _, err := docFile.WriteString(docContent); err != nil {
			docFile.Close()
			return fmt.Errorf("failed to write doc.go for package %s: %w", pkg, err)
		}
		docFile.Close()
	}

	// Generate main index file
	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

// Package generated contains all generated code for Facebook Business API
package generated

// Re-export all types, tools, and client methods from subdirectories
`

	file, err := os.Create(filepath.Join(outputDir, "doc.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(tmpl)
	return err
}

func generateConstantsFile(ctx *CodegenContext, outputDir string) error {
	// Create constants directory
	constantsDir := filepath.Join(outputDir, "constants")
	if err := os.MkdirAll(constantsDir, 0755); err != nil {
		return fmt.Errorf("failed to create constants directory: %w", err)
	}

	// Generate field constants for each type
	for typeName, spec := range ctx.Specs {
		if len(spec.Fields) > 0 {
			if err := generateFieldConstantsFile(constantsDir, typeName, spec); err != nil {
				return fmt.Errorf("failed to generate field constants for %s: %w", typeName, err)
			}
		}
	}

	// Generate main constants file
	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package generated

// APIVersion is the Facebook Graph API version used by this SDK
const APIVersion = "` + APIVersion + `"
`

	file, err := os.Create(filepath.Join(outputDir, "constants.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(tmpl)
	return err
}

func generateFieldConstantsFile(constantsDir string, typeName string, spec *APISpec) error {
	tmplData, err := templateFS.ReadFile("templates/field_constants.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read field constants template: %w", err)
	}

	type fieldData struct {
		OriginalName string
		SafeName     string
	}

	data := struct {
		TypeName string
		Fields   []fieldData
	}{
		TypeName: typeName,
	}

	// Convert field names to Go constant names
	for _, field := range spec.Fields {
		constName := ""
		// Replace dots with underscores first
		sanitizedName := strings.ReplaceAll(field.Name, ".", "_")
		
		// Convert field name to PascalCase for constant name
		parts := strings.Split(sanitizedName, "_")
		for _, part := range parts {
			if len(part) > 0 {
				constName += strings.ToUpper(part[:1]) + part[1:]
			}
		}
		
		// Handle special cases
		if constName == "" {
			constName = "Field" + sanitizedName
		}
		
		// If starts with number, prefix with Field
		if len(constName) > 0 && constName[0] >= '0' && constName[0] <= '9' {
			constName = "Field" + constName
		}
		
		// Handle Go keywords
		switch constName {
		case "Type", "Func", "Var", "Const", "Package", "Import", "Interface", 
		     "Struct", "Map", "Chan", "Select", "Case", "Default", "If", "Else",
		     "For", "Range", "Switch", "Return", "Break", "Continue", "Goto",
		     "Defer", "Go":
			constName = "Field" + constName
		}
		
		data.Fields = append(data.Fields, fieldData{
			OriginalName: field.Name,
			SafeName:     constName,
		})
	}

	t, err := template.New("fieldConstants").Funcs(getTemplateFuncs()).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse field constants template: %w", err)
	}

	filename := fmt.Sprintf("%s_fields.go", strings.ToLower(typeName))
	file, err := os.Create(filepath.Join(constantsDir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

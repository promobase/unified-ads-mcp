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
//
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
	AvailableFields []string    // For GET endpoints, the fields that can be requested
	APIParams       []Parameter // Original API parameters for params object generation
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

	if err := generateCodeNew(ctx); err != nil {
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
				var targetType string
				if api.Return != "" && api.Return != "Object" {
					targetType = api.Return
				} else if api.Endpoint == "" || api.Endpoint == "/" {
					targetType = nodeName
				} else {
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

// generateCodeNew generates the new structure with individual endpoint files
func generateCodeNew(ctx *CodegenContext) error {
	// Create generated directory structure
	generatedDir := "../generated"

	// Create base directories
	dirs := []string{
		generatedDir,
		filepath.Join(generatedDir, "types"),
		filepath.Join(generatedDir, "enums"),
		filepath.Join(generatedDir, "constants"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Group tools by node type
	toolsByNode := make(map[string][]MCPTool)
	for _, tool := range ctx.Tools {
		toolsByNode[tool.NodeType] = append(toolsByNode[tool.NodeType], tool)
	}

	// Generate endpoint files grouped by node type
	for nodeName, tools := range toolsByNode {
		// Create node directory
		nodeDir := filepath.Join(generatedDir, strings.ToLower(nodeName))
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			return fmt.Errorf("failed to create node directory %s: %w", nodeDir, err)
		}

		// Generate individual endpoint files
		for _, tool := range tools {
			if err := generateEndpointFile(nodeDir, nodeName, tool, ctx); err != nil {
				return fmt.Errorf("failed to generate endpoint %s: %w", tool.Name, err)
			}
		}

		// Generate node registry file
		if err := generateNodeRegistryFile(nodeDir, nodeName, tools); err != nil {
			return fmt.Errorf("failed to generate registry for %s: %w", nodeName, err)
		}
	}

	// Generate types (one file per type)
	if err := generateTypesFiles(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}

	// Generate enums
	if err := generateEnumsFile(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate enums: %w", err)
	}

	// Generate constants
	if err := generateConstantsFile(ctx, generatedDir); err != nil {
		return fmt.Errorf("failed to generate constants: %w", err)
	}

	// Generate main registry file
	if err := generateMainRegistryFile(generatedDir, ctx); err != nil {
		return fmt.Errorf("failed to generate main registry: %w", err)
	}

	// Generate backward compatibility file
	if err := generateBackwardCompatibilityFile(generatedDir); err != nil {
		return fmt.Errorf("failed to generate backward compatibility file: %w", err)
	}

	return nil
}

// generateEndpointFile generates a single endpoint file
func generateEndpointFile(nodeDir string, nodeName string, tool MCPTool, ctx *CodegenContext) error {
	tmplData, err := templateFS.ReadFile("templates/endpoint.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read endpoint template: %w", err)
	}

	t, err := template.New("endpoint").Funcs(getTemplateFuncs()).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse endpoint template: %w", err)
	}

	// Create a safe filename from the tool name
	filename := fmt.Sprintf("%s.go", strings.ReplaceAll(tool.Name, "_", "_"))
	file, err := os.Create(filepath.Join(nodeDir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	// Determine if strings import is needed
	// We need strings.Join for fields array parameters
	needsStrings := false
	for _, param := range tool.Parameters {
		if param.Name == "fields" && param.Type == "array" {
			needsStrings = true
			break
		}
	}

	return t.Execute(file, map[string]interface{}{
		"PackageName":  strings.ToLower(nodeName),
		"NodeName":     nodeName,
		"Tool":         tool,
		"APIVersion":   APIVersion,
		"NeedsStrings": needsStrings,
	})
}

// generateNodeRegistryFile generates the registry file for a node
func generateNodeRegistryFile(nodeDir string, nodeName string, tools []MCPTool) error {
	tmplData, err := templateFS.ReadFile("templates/node_registry.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read node registry template: %w", err)
	}

	t, err := template.New("node_registry").Funcs(getTemplateFuncs()).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse node registry template: %w", err)
	}

	file, err := os.Create(filepath.Join(nodeDir, "registry.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	// Extract endpoint names for imports
	endpointNames := []string{}
	for _, tool := range tools {
		endpointNames = append(endpointNames, tool.Name)
	}

	return t.Execute(file, map[string]interface{}{
		"PackageName": strings.ToLower(nodeName),
		"NodeName":    nodeName,
		"Tools":       tools,
		"Endpoints":   endpointNames,
	})
}

// generateMainRegistryFile generates the main registry that imports all nodes
func generateMainRegistryFile(outputDir string, ctx *CodegenContext) error {
	// Get unique node names
	nodeTypes := []string{}
	nodeTypesMap := make(map[string]bool)
	for _, tool := range ctx.Tools {
		if !nodeTypesMap[tool.NodeType] {
			nodeTypesMap[tool.NodeType] = true
			nodeTypes = append(nodeTypes, tool.NodeType)
		}
	}

	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package generated

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
{{range .NodeTypes}}	"unified-ads-mcp/internal/facebook/generated/{{. | lower}}"
{{end}})

// GetAllTools returns all available MCP tools
func GetAllTools() []mcp.Tool {
	var tools []mcp.Tool

{{range .NodeTypes}}	tools = append(tools, {{. | lower}}.GetTools()...)
{{end}}

	return tools
}

// GetFilteredTools returns filtered MCP tools based on enabled objects
func GetFilteredTools(enabledObjects map[string]bool) []mcp.Tool {
	var tools []mcp.Tool

{{range .NodeTypes}}	if enabled, ok := enabledObjects["{{.}}"]; ok && enabled {
		tools = append(tools, {{. | lower}}.GetTools()...)
	}
{{end}}

	return tools
}

// GetHandlers returns all MCP handlers
func GetHandlers() map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	handlers := make(map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error))

{{range .NodeTypes}}	for name, handler := range {{. | lower}}.GetHandlers() {
		handlers[name] = handler
	}
{{end}}

	return handlers
}

// RegisterTools registers all tools with the MCP server
func RegisterTools(s *server.MCPServer) error {
	tools := GetAllTools()
	handlers := GetHandlers()

	for _, tool := range tools {
		s.AddTool(tool, handlers[tool.Name])
	}

	return nil
}
`

	file, err := os.Create(filepath.Join(outputDir, "registry.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	t, err := template.New("main_registry").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(tmpl)
	if err != nil {
		return err
	}

	return t.Execute(file, map[string]interface{}{
		"NodeTypes": nodeTypes,
	})
}

// generateBackwardCompatibilityFile generates the backward compatibility file
func generateBackwardCompatibilityFile(outputDir string) error {
	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package facebook

import (
	"context"

	"unified-ads-mcp/internal/facebook/generated"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetMCPTools returns all available MCP tools for Facebook Business API
// Deprecated: accessToken parameter is ignored, use context instead
func GetMCPTools(accessToken string) []mcp.Tool {
	return generated.GetAllTools()
}

// GetFilteredMCPTools returns filtered MCP tools based on enabled objects
func GetFilteredMCPTools(enabledObjects map[string]bool) []mcp.Tool {
	return generated.GetFilteredTools(enabledObjects)
}

// GetContextAwareHandlers returns handlers that get auth from context
// Deprecated: accessToken parameter is ignored, use context instead
func GetContextAwareHandlers(accessToken string) map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return generated.GetHandlers()
}

// RegisterMCPTools registers all Facebook Business API tools with the MCP server
// Deprecated: accessToken parameter is ignored, use context instead
func RegisterMCPTools(s *server.MCPServer, accessToken string) error {
	return generated.RegisterTools(s)
}
`

	file, err := os.Create(filepath.Join("..", "generated_tools.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(tmpl)
	return err
}

// Helper functions (reuse from original)
func normalizeEndpoint(endpoint string) string {
	return strings.ReplaceAll(endpoint, "/", "_")
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

func findEnumValues(ctx *CodegenContext, paramType string) []string {
	for _, enum := range ctx.Enums {
		if strings.Contains(paramType, enum.Name) {
			return enum.Values
		}
	}
	return nil
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

// getTemplateFuncs returns common template functions
func getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"capitalizeFirst":         capitalizeFirst,
		"sanitizeVarName":         sanitizeVarName,
		"sanitizeEnumValue":       sanitizeEnumValue,
		"convertTypeToJSONSchema": convertTypeToJSONSchema,
		"extractItemType":         extractItemType,
		"hasPrefix":               strings.HasPrefix,
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

// Reuse existing helper functions for types, enums, and constants generation
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

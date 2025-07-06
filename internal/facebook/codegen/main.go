package main

import (
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
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
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
	Name        string
	Description string
	Method      string
	Endpoint    string
	Return      string
	Parameters  []MCPParameter
	NodeType    string
}

// MCPParameter represents a parameter for an MCP tool
type MCPParameter struct {
	Name        string
	Type        string
	Required    bool
	Description string
	EnumValues  []string
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

			// Convert parameters
			for _, param := range api.Params {
				mcpParam := MCPParameter{
					Name:        param.Name,
					Type:        convertTypeForMCP(param.Type),
					Required:    param.Required,
					Description: fmt.Sprintf("%s parameter for %s", param.Name, api.Endpoint),
				}

				// Check if this parameter has enum values
				if enumValues := findEnumValues(ctx, param.Type); len(enumValues) > 0 {
					mcpParam.EnumValues = enumValues
				}

				tool.Parameters = append(tool.Parameters, mcpParam)
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
	if err := generateConstantsFile(generatedDir); err != nil {
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

	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package types
{{if .NeedsTimeImport}}
import "time"
{{end}}
// {{.Name}} represents a Facebook {{.Name}} object
type {{.Name}} struct {
{{range .Fields}}	{{.Name}} {{.Type}} ` + "`json:\"{{.JSONTag}}\"`" + `
{{end}}}
`

	t, err := template.New("type").Parse(tmpl)
	if err != nil {
		return err
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
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"unified-ads-mcp/internal/facebook/generated/tools"
)

// GetMCPTools returns all available MCP tools for Facebook Business API
func GetMCPTools(accessToken string) []mcp.Tool {
	return tools.GetAllTools(accessToken)
}

// RegisterMCPTools registers all Facebook Business API tools with the MCP server
func RegisterMCPTools(s *server.MCPServer, accessToken string) error {
	return tools.RegisterTools(s, accessToken)
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
	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// {{.NodeName}}Client provides methods for {{.NodeName}} operations
type {{.NodeName}}Client struct {
	accessToken string
}

// New{{.NodeName}}Client creates a new {{.NodeName}} client
func New{{.NodeName}}Client(accessToken string) *{{.NodeName}}Client {
	return &{{.NodeName}}Client{
		accessToken: accessToken,
	}
}

{{range .Tools}}
// {{.Name}} {{.Description}}
func (c *{{$.NodeName}}Client) {{capitalizeFirst .Name}}(args map[string]interface{}) (interface{}, error) {
	// Extract parameters
{{range .Parameters}}{{if .Required}}	{{sanitizeVarName .Name}}, ok := args["{{.Name}}"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: {{.Name}}")
	}
	_ = {{sanitizeVarName .Name}} // Suppress unused variable warning
{{end}}{{end}}

	// Build request URL and parameters
	baseURL := fmt.Sprintf("https://graph.facebook.com/%s/%s", "{{$.APIVersion}}", "{{.Endpoint}}")
	urlParams := url.Values{}
	urlParams.Set("access_token", c.accessToken)

{{range .Parameters}}	if val, ok := args["{{.Name}}"]; ok {
		urlParams.Set("{{.Name}}", fmt.Sprintf("%v", val))
	}
{{end}}

	// Make HTTP request
	var resp *http.Response
	var err error

	switch "{{.Method}}" {
	case "GET":
		resp, err = http.Get(baseURL + "?" + urlParams.Encode())
	case "POST":
		resp, err = http.PostForm(baseURL, urlParams)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: {{.Method}}")
	}

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

{{end}}`

	t, err := template.New("client").Funcs(template.FuncMap{
		"capitalizeFirst": capitalizeFirst,
		"sanitizeVarName": sanitizeVarName,
	}).Parse(tmpl)
	if err != nil {
		return err
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
		if err := generateNodeToolsFile(toolsDir, nodeName, tools); err != nil {
			return fmt.Errorf("failed to generate tools for %s: %w", nodeName, err)
		}
	}

	// Generate main tools registry file
	if err := generateToolsRegistryFile(toolsDir, ctx); err != nil {
		return fmt.Errorf("failed to generate tools registry: %w", err)
	}

	return nil
}

func generateNodeToolsFile(toolsDir string, nodeName string, tools []MCPTool) error {
	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"unified-ads-mcp/internal/facebook/generated/client"
	"unified-ads-mcp/internal/shared"
)

// Get{{.NodeName}}Tools returns MCP tools for {{.NodeName}}
func Get{{.NodeName}}Tools(accessToken string) []mcp.Tool {
	var tools []mcp.Tool

{{range .Tools}}
	// {{.Name}} tool
	{{sanitizeVarName .Name}}Tool := mcp.NewTool("{{.Name}}",
		mcp.WithDescription("{{.Description}}"),
		mcp.WithString("access_token",
			mcp.Required(),
			mcp.Description("Facebook access token for authentication"),
		),
{{range .Parameters}}		mcp.With{{paramType .Type}}("{{.Name}}",
{{if .Required}}			mcp.Required(),
{{end}}			mcp.Description("{{.Description}}"),
{{if .EnumValues}}			mcp.Enum({{range $i, $val := .EnumValues}}{{if $i}}, {{end}}"{{$val}}"{{end}}),
{{end}}		),
{{end}}	)
	tools = append(tools, {{sanitizeVarName .Name}}Tool)
{{end}}

	return tools
}

// Get{{.NodeName}}ToolsWithoutAuth returns MCP tools for {{.NodeName}} without access_token parameter
func Get{{.NodeName}}ToolsWithoutAuth() []mcp.Tool {
	var tools []mcp.Tool

{{range .Tools}}
	// {{.Name}} tool
	{{sanitizeVarName .Name}}Tool := mcp.NewTool("{{.Name}}",
		mcp.WithDescription("{{.Description}}"),
{{range .Parameters}}		mcp.With{{paramType .Type}}("{{.Name}}",
{{if .Required}}			mcp.Required(),
{{end}}			mcp.Description("{{.Description}}"),
{{if .EnumValues}}			mcp.Enum({{range $i, $val := .EnumValues}}{{if $i}}, {{end}}"{{$val}}"{{end}}),
{{end}}		),
{{end}}	)
	tools = append(tools, {{sanitizeVarName .Name}}Tool)
{{end}}

	return tools
}

// {{.NodeName}} handlers

{{range .Tools}}
// Handle{{capitalizeFirst .Name}} handles the {{.Name}} tool
func Handle{{capitalizeFirst .Name}}(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get access token
	accessToken, err := request.RequireString("access_token")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: access_token"), nil
	}

	// Create client
	client := client.New{{$.NodeName}}Client(accessToken)

	// Build arguments map
	args := make(map[string]interface{})

{{range .Parameters}}{{if .Required}}	// Required: {{.Name}}
	{{sanitizeVarName .Name}}, err := request.{{requireMethod .Type}}("{{.Name}}")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("missing required parameter {{.Name}}: %v", err)), nil
	}
	args["{{.Name}}"] = {{sanitizeVarName .Name}}
{{else}}	// Optional: {{.Name}}
{{if eq .Type "string"}}	if val := request.GetString("{{.Name}}", ""); val != "" {
		args["{{.Name}}"] = val
	}
{{else if eq .Type "integer"}}	if val := request.GetInt("{{.Name}}", 0); val != 0 {
		args["{{.Name}}"] = val
	}
{{else if eq .Type "number"}}	if val := request.GetFloat("{{.Name}}", 0); val != 0 {
		args["{{.Name}}"] = val
	}
{{else if eq .Type "boolean"}}	if val := request.GetBool("{{.Name}}", false); val {
		args["{{.Name}}"] = val
	}
{{else}}	// {{.Type}} type - using string
	if val := request.GetString("{{.Name}}", ""); val != "" {
		args["{{.Name}}"] = val
	}
{{end}}{{end}}
{{end}}

	// Call the client method
	result, err := client.{{capitalizeFirst .Name}}(args)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to execute {{.Name}}: %v", err)), nil
	}

	// Return the result as JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

{{end}}
// Context-aware handlers

{{range .Tools}}
// HandleContext{{capitalizeFirst .Name}} handles the {{.Name}} tool with context-based auth
func HandleContext{{capitalizeFirst .Name}}(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get access token from context
	accessToken, ok := shared.FacebookAccessTokenFromContext(ctx)
	if !ok {
		return mcp.NewToolResultError("Facebook access token not found in context"), nil
	}

	// Create client
	client := client.New{{$.NodeName}}Client(accessToken)

	// Build arguments map
	args := make(map[string]interface{})

{{range .Parameters}}{{if .Required}}	// Required: {{.Name}}
	{{sanitizeVarName .Name}}, err := request.{{requireMethod .Type}}("{{.Name}}")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("missing required parameter {{.Name}}: %v", err)), nil
	}
	args["{{.Name}}"] = {{sanitizeVarName .Name}}
{{else}}	// Optional: {{.Name}}
{{if eq .Type "string"}}	if val := request.GetString("{{.Name}}", ""); val != "" {
		args["{{.Name}}"] = val
	}
{{else if eq .Type "integer"}}	if val := request.GetInt("{{.Name}}", 0); val != 0 {
		args["{{.Name}}"] = val
	}
{{else if eq .Type "number"}}	if val := request.GetFloat("{{.Name}}", 0); val != 0 {
		args["{{.Name}}"] = val
	}
{{else if eq .Type "boolean"}}	if val := request.GetBool("{{.Name}}", false); val {
		args["{{.Name}}"] = val
	}
{{else}}	// {{.Type}} type - using string
	if val := request.GetString("{{.Name}}", ""); val != "" {
		args["{{.Name}}"] = val
	}
{{end}}{{end}}
{{end}}

	// Call the client method
	result, err := client.{{capitalizeFirst .Name}}(args)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to execute {{.Name}}: %v", err)), nil
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
				return "String" // Arrays passed as JSON strings
			case "object":
				return "String" // Objects passed as JSON strings
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
	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetAllTools returns all available MCP tools for Facebook Business API
func GetAllTools(accessToken string) []mcp.Tool {
	var allTools []mcp.Tool

{{range $nodeName, $_ := .NodeNames}}	allTools = append(allTools, Get{{$nodeName}}Tools(accessToken)...)
{{end}}

	return allTools
}

// GetAllToolsWithoutAuth returns all available MCP tools without access_token parameter
func GetAllToolsWithoutAuth() []mcp.Tool {
	var allTools []mcp.Tool

{{range $nodeName, $_ := .NodeNames}}	allTools = append(allTools, Get{{$nodeName}}ToolsWithoutAuth()...)
{{end}}

	return allTools
}

// GetFilteredTools returns tools filtered by object type
func GetFilteredTools(accessToken string, enabledObjects map[string]bool) []mcp.Tool {
	var filteredTools []mcp.Tool

{{range $nodeName, $_ := .NodeNames}}	if enabled, ok := enabledObjects[strings.ToLower("{{$nodeName}}")]; ok && enabled {
		filteredTools = append(filteredTools, Get{{$nodeName}}Tools(accessToken)...)
	}
{{end}}

	return filteredTools
}

// GetFilteredToolsWithoutAuth returns tools filtered by object type without auth
func GetFilteredToolsWithoutAuth(enabledObjects map[string]bool) []mcp.Tool {
	var filteredTools []mcp.Tool

{{range $nodeName, $_ := .NodeNames}}	if enabled, ok := enabledObjects[strings.ToLower("{{$nodeName}}")]; ok && enabled {
		filteredTools = append(filteredTools, Get{{$nodeName}}ToolsWithoutAuth()...)
	}
{{end}}

	return filteredTools
}

// RegisterTools registers all Facebook Business API tools with the MCP server
func RegisterTools(s *server.MCPServer, accessToken string) error {
	// Get all tools
	tools := GetAllTools(accessToken)

	// Create a map of handlers
	handlers := GetAllHandlers()

	// Register each tool with its handler
	for i := range tools {
		handler, ok := handlers[tools[i].Name]
		if !ok {
			continue // Skip if no handler found
		}
		s.AddTool(tools[i], handler)
	}

	return nil
}

// RegisterFilteredTools registers filtered Facebook Business API tools with the MCP server
func RegisterFilteredTools(s *server.MCPServer, accessToken string, enabledObjects map[string]bool) error {
	// Get filtered tools
	tools := GetFilteredTools(accessToken, enabledObjects)

	// Create a map of handlers
	handlers := GetAllHandlers()

	// Register each tool with its handler
	for i := range tools {
		handler, ok := handlers[tools[i].Name]
		if !ok {
			continue // Skip if no handler found
		}
		s.AddTool(tools[i], handler)
	}

	return nil
}

// RegisterFilteredToolsWithContextAuth registers filtered tools with context-aware handlers
func RegisterFilteredToolsWithContextAuth(s *server.MCPServer, enabledObjects map[string]bool) error {
	// Get filtered tools without auth param
	tools := GetFilteredToolsWithoutAuth(enabledObjects)

	// Create a map of context-aware handlers
	handlers := GetContextAwareHandlers()

	// Register each tool with its context-aware handler
	for i := range tools {
		handler, ok := handlers[tools[i].Name]
		if !ok {
			continue // Skip if no handler found
		}
		s.AddTool(tools[i], handler)
	}

	return nil
}

// GetAllHandlers returns a map of tool name to handler function
func GetAllHandlers() map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	handlers := make(map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error))

{{range .Tools}}	handlers["{{.Name}}"] = Handle{{capitalizeFirst .Name}}
{{end}}

	return handlers
}

// GetContextAwareHandlers returns a map of tool name to context-aware handler function
func GetContextAwareHandlers() map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	handlers := make(map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error))

{{range .Tools}}	handlers["{{.Name}}"] = HandleContext{{capitalizeFirst .Name}}
{{end}}

	return handlers
}
`

	// Get unique node names
	nodeNames := make(map[string]bool)
	for _, tool := range ctx.Tools {
		nodeNames[tool.NodeType] = true
	}

	t, err := template.New("registry").Funcs(template.FuncMap{
		"capitalizeFirst": capitalizeFirst,
	}).Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(toolsDir, "registry.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, map[string]interface{}{
		"NodeNames": nodeNames,
		"Tools":     ctx.Tools,
	})
}

func generateEnumsFile(ctx *CodegenContext, outputDir string) error {
	enumDir := filepath.Join(outputDir, "enums")

	tmpl := `// Code generated by Facebook Business API codegen. DO NOT EDIT.

package enums

// Enum definitions from Facebook Business API

{{range .Enums}}{{$enumName := .Name}}
// {{.Name}} enum values for {{.Node}}.{{.FieldOrParam}}
type {{.Name}} string

const (
{{range $i, $val := .Values}}	{{$enumName}}_{{sanitizeEnumValue $val}} {{$enumName}} = "{{$val}}"
{{end}})

// Valid{{.Name}} returns all valid values for {{.Name}}
func Valid{{.Name}}() []{{.Name}} {
	return []{{.Name}}{
{{range .Values}}		{{$enumName}}_{{sanitizeEnumValue .}},
{{end}}	}
}

{{end}}`

	t, err := template.New("enums").Funcs(template.FuncMap{
		"sanitizeEnumValue": sanitizeEnumValue,
	}).Parse(tmpl)
	if err != nil {
		return err
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

func generateConstantsFile(outputDir string) error {
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

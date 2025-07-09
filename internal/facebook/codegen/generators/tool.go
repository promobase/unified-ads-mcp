package generators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type API struct {
	Method   string  `json:"method"`
	Endpoint string  `json:"endpoint"`
	Return   string  `json:"return"`
	Params   []Param `json:"params"`
}

type Param struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

type ToolSpec struct {
	APIs []API `json:"apis"`
}

type ToolGenerator struct {
	specs         map[string]*ToolSpec // Object name -> spec
	enumTypes     map[string]bool      // Track which types are enums
	outputPath    string
	fieldSpecs    map[string]*APISpec // Object name -> field spec for struct generation
	complexTypes  map[string]bool     // Track which types need struct definitions
	enumTypeNames map[string]string   // Map from enum type to generated Go type name
	useSchema     bool                // Use schema-based generation
	schemaGen     *SchemaGenerator    // Schema generator for enhanced type handling
}

type SchemaParam struct {
	Name        string
	Type        string
	Required    bool
	Description string
	// For typed template
	SchemaType string // string, number, boolean, array, object
	ItemsType  string // For arrays
	JSONName   string // JSON field name
}

type TypedParam struct {
	GoName        string // Go struct field name (capitalized)
	GoType        string // Go type (string, int, bool, []string, etc.)
	JSONName      string // JSON field name
	JSONTag       string // Additional JSON tags like omitempty
	JSONSchemaTag string // JSONSchema tag for validation and documentation
}

type ToolData struct {
	ObjectName      string
	Method          string
	Endpoint        string
	Return          string
	ToolName        string
	HandlerName     string
	Description     string
	InputSchema     string // JSON string (deprecated, use RawSchema)
	RawSchema       string // Raw JSON schema for mcp.NewToolWithRawSchema
	NeedsID         bool
	Params          []Param
	HasComplexLogic bool          // Whether handler needs custom logic
	HasQueryParams  bool          // Whether POST/PUT has query params
	SchemaParams    []SchemaParam // For new template format
	TypedParams     []TypedParam  // For typed handlers
}

type ObjectGroup struct {
	ObjectName string
	Tools      []ToolData
}

func NewToolGenerator(outputPath string) *ToolGenerator {
	return &ToolGenerator{
		specs:         make(map[string]*ToolSpec),
		enumTypes:     make(map[string]bool),
		outputPath:    outputPath,
		fieldSpecs:    make(map[string]*APISpec),
		complexTypes:  make(map[string]bool),
		enumTypeNames: make(map[string]string),
		useSchema:     false,
		schemaGen:     nil,
	}
}

// EnableSchemaGeneration enables schema-based tool generation
func (g *ToolGenerator) EnableSchemaGeneration() {
	g.useSchema = true
	g.schemaGen = NewSchemaGenerator()

	// Load enum definitions if available
	if len(g.enumTypes) > 0 {
		enumDefs := make(map[string][]string)
		// TODO: Convert enumTypes to enumDefs format
		g.schemaGen.LoadEnumDefinitions(enumDefs)
	}
}

func (g *ToolGenerator) LoadEnumTypes(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read enum types: %w", err)
	}

	var enums []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &enums); err != nil {
		return fmt.Errorf("failed to parse enum types: %w", err)
	}

	for _, enum := range enums {
		g.enumTypes[enum.Name] = true
	}

	log.Printf("Loaded %d enum types for tool generation", len(g.enumTypes))
	return nil
}

func (g *ToolGenerator) LoadAPISpecs(specsDir string) error {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return fmt.Errorf("failed to read specs directory: %w", err)
	}

	// First pass: load all specs
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") && entry.Name() != "enum_types.json" {
			specPath := filepath.Join(specsDir, entry.Name())
			data, err := os.ReadFile(specPath)
			if err != nil {
				log.Printf("Warning: failed to read %s: %v", specPath, err)
				continue
			}

			// Parse both APIs and fields
			var fullSpec struct {
				APIs   []API   `json:"apis"`
				Fields []Field `json:"fields"`
			}
			if err := json.Unmarshal(data, &fullSpec); err != nil {
				log.Printf("Warning: failed to parse %s: %v", specPath, err)
				continue
			}

			objectName := strings.TrimSuffix(entry.Name(), ".json")
			g.specs[objectName] = &ToolSpec{APIs: fullSpec.APIs}

			// Store field specs for complex type detection
			if len(fullSpec.Fields) > 0 {
				g.fieldSpecs[objectName] = &APISpec{Fields: fullSpec.Fields}
			}
		}
	}

	// Second pass: identify complex types from field references
	g.identifyComplexTypes()

	log.Printf("Loaded %d API specs with endpoints", len(g.specs))
	log.Printf("Identified %d complex types", len(g.complexTypes))
	return nil
}

func (g *ToolGenerator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(g.outputPath, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Define core objects to generate
	coreObjects := map[string]bool{
		"AdAccount":  true,
		"Campaign":   true,
		"AdSet":      true,
		"AdCreative": true,
		"Ad":         true,
		"User":       true,
	}

	// Generate separate file for each core object
	for objectName := range coreObjects {
		if spec, exists := g.specs[objectName]; exists {
			if err := g.generateToolsForObject(objectName, spec); err != nil {
				return fmt.Errorf("failed to generate tools for %s: %w", objectName, err)
			}
		}
	}

	// Generate common utilities file
	if err := g.generateCommonFile(); err != nil {
		return fmt.Errorf("failed to generate common file: %w", err)
	}

	// Generate default fields file
	if err := g.generateDefaultFieldsFile(); err != nil {
		return fmt.Errorf("failed to generate default fields file: %w", err)
	}

	// Generate register_tools.go that imports all the tool files
	if err := g.generateRegisterFile(coreObjects); err != nil {
		return fmt.Errorf("failed to generate register file: %w", err)
	}

	// Run go fmt on all generated files
	if err := g.formatGeneratedFiles(); err != nil {
		return fmt.Errorf("failed to format generated files: %w", err)
	}

	return nil
}

func (g *ToolGenerator) generateToolsForObject(objectName string, spec *ToolSpec) error {
	// Prepare tools data for this object
	var tools []ToolData

	// Process APIs for this object
	for _, api := range spec.APIs {
		toolName := g.generateToolName(objectName, api)
		// Convert snake_case tool name to CamelCase for handler function name
		handlerName := g.toCamelCase(toolName) + "Handler"

		// Generate input schema based on method
		var inputSchemaJSON []byte
		var rawSchemaJSON string
		var schemaParams []SchemaParam

		// Always try raw schema generation first for better type safety
		if rawSchema, err := g.generateRawSchema(objectName, api); err == nil {
			rawSchemaJSON = rawSchema
			log.Printf("Using raw schema generation for %s", toolName)
		} else {
			log.Printf("Raw schema generation failed for %s: %v", toolName, err)
		}

		// Try enhanced schema generation if enabled
		if g.useSchema && g.schemaGen != nil && g.canUseSchemaGeneration(objectName, api) {
			// Try schema-based generation for better type handling
			schema, params, err := g.generateSchemaForAPI(objectName, api)
			if err == nil {
				inputSchemaJSON = schema
				schemaParams = params
				log.Printf("Using enhanced schema-based generation for %s", toolName)
			} else {
				log.Printf("Enhanced schema generation failed for %s: %v", toolName, err)
			}
		}

		// Fall back to original generation if enhanced schema not used or failed
		if inputSchemaJSON == nil {
			inputSchema := g.generateInputSchema(objectName, api)
			inputSchemaJSON, _ = json.Marshal(inputSchema)
			schemaParams = g.convertToSchemaParams(api.Params, g.needsObjectID(objectName, api.Endpoint), api.Method)
		}

		// Determine if handler needs custom logic
		hasComplexLogic := g.hasComplexLogic(api)
		hasQueryParams := g.hasQueryParams(api)
		needsID := g.needsObjectID(objectName, api.Endpoint)

		// Convert to typed params
		typedParams := g.convertToTypedParams(api.Params, api.Method)

		tools = append(tools, ToolData{
			ObjectName:      objectName,
			Method:          api.Method,
			Endpoint:        api.Endpoint,
			Return:          api.Return,
			ToolName:        toolName,
			HandlerName:     handlerName,
			Description:     strings.ReplaceAll(g.generateDescription(objectName, api), `"`, `\"`),
			InputSchema:     string(inputSchemaJSON),
			RawSchema:       rawSchemaJSON,
			NeedsID:         needsID,
			Params:          api.Params,
			HasComplexLogic: hasComplexLogic,
			HasQueryParams:  hasQueryParams,
			TypedParams:     typedParams,
			SchemaParams:    schemaParams,
		})
	}

	// Prepare template data
	data := struct {
		ObjectName string
		Tools      []ToolData
		TotalTools int
	}{
		ObjectName: objectName,
		Tools:      tools,
		TotalTools: len(tools),
	}

	// Load and execute template
	var tmplContent []byte
	var err error

	// Try raw schema template first (best type safety)
	tmplContent, err = os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "object_tools_raw_schema.go.tmpl"))
	if err == nil {
		log.Printf("Using raw schema template for %s", objectName)
	} else {
		log.Printf("Raw schema template not found, falling back: %v", err)

		// Use schema template when schema generation is enabled
		if g.useSchema {
			tmplContent, err = os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "object_tools_with_schema.go.tmpl"))
			if err != nil {
				log.Printf("Schema template not found, falling back to typed template: %v", err)
			}
		}

		// Fallback to typed template if schema template not available or not enabled
		if tmplContent == nil {
			tmplContent, err = os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "object_tools_typed.go.tmpl"))
			if err != nil {
				// Fallback to refactored template if typed doesn't exist
				tmplContent, err = os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "object_tools_refactored.go.tmpl"))
				if err != nil {
					// Fallback to original template
					tmplContent, err = os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "object_tools.go.tmpl"))
					if err != nil {
						return fmt.Errorf("failed to read template: %w", err)
					}
				}
			}
		}
	}

	tmpl, err := template.New("object_tools").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write to file named after the object (lowercase)
	filename := fmt.Sprintf("%s_tools.go", strings.ToLower(objectName))
	return os.WriteFile(filepath.Join(g.outputPath, filename), buf.Bytes(), 0o644)
}

func (g *ToolGenerator) generateCommonFile() error {
	// Load and execute template
	// Use the template with utility functions
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "tools_common_with_utils.go.tmpl"))
	if err != nil {
		// Fallback to original template
		tmplContent, err = os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "tools_common_with_utils.go.tmpl"))
	}
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Write the file as-is (no template variables needed)
	return os.WriteFile(filepath.Join(g.outputPath, "tools_common.go"), tmplContent, 0o644)
}

func (g *ToolGenerator) generateDefaultFieldsFile() error {
	// Load and execute template
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "default_fields.go.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Write the file as-is (no template variables needed)
	return os.WriteFile(filepath.Join(g.outputPath, "default_fields.go"), tmplContent, 0o644)
}

func (g *ToolGenerator) generateRegisterFile(coreObjects map[string]bool) error {
	// Prepare data for register file
	var objectNames []string
	for name := range coreObjects {
		if _, exists := g.specs[name]; exists {
			objectNames = append(objectNames, name)
		}
	}
	sort.Strings(objectNames)

	data := struct {
		Objects []string
	}{
		Objects: objectNames,
	}

	// Load and execute template
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "register_tools.go.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("register").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return os.WriteFile(filepath.Join(g.outputPath, "register_tools.go"), buf.Bytes(), 0o644)
}

func (g *ToolGenerator) generateTools() error {
	// Prepare tools data for template

	var tools []ToolData
	totalAPIs := 0

	// Sort object names for consistent output
	var objectNames []string
	for name := range g.specs {
		objectNames = append(objectNames, name)
	}
	sort.Strings(objectNames)

	// Process each object's APIs
	for _, objectName := range objectNames {
		spec := g.specs[objectName]
		for _, api := range spec.APIs {
			totalAPIs++

			toolName := g.generateToolName(objectName, api)
			handlerName := toolName + "Handler"

			// Generate input schema based on method
			inputSchema := g.generateInputSchema(objectName, api)
			// Marshal without indentation for embedding in Go code
			inputSchemaJSON, _ := json.Marshal(inputSchema)

			// Determine if handler needs custom logic
			hasComplexLogic := g.hasComplexLogic(api)
			hasQueryParams := g.hasQueryParams(api)

			// Convert params to SchemaParams
			schemaParams := g.convertToSchemaParams(api.Params, g.needsObjectID(objectName, api.Endpoint), api.Method)
			typedParams := g.convertToTypedParams(api.Params, api.Method)

			// Generate raw schema for this tool
			rawSchema, _ := g.generateRawSchema(objectName, api)

			tools = append(tools, ToolData{
				ObjectName:      objectName,
				Method:          api.Method,
				Endpoint:        api.Endpoint,
				Return:          api.Return,
				ToolName:        toolName,
				HandlerName:     handlerName,
				Description:     strings.ReplaceAll(g.generateDescription(objectName, api), `"`, `\"`),
				InputSchema:     string(inputSchemaJSON),
				RawSchema:       rawSchema,
				NeedsID:         g.needsObjectID(objectName, api.Endpoint),
				Params:          api.Params,
				HasComplexLogic: hasComplexLogic,
				HasQueryParams:  hasQueryParams,
				SchemaParams:    schemaParams,
				TypedParams:     typedParams,
			})
		}
	}

	// Group tools by object
	objectGroups := make(map[string][]ToolData)
	for _, tool := range tools {
		objectGroups[tool.ObjectName] = append(objectGroups[tool.ObjectName], tool)
	}

	// Convert to sorted slice of groups
	var groups []ObjectGroup
	for _, name := range objectNames {
		if tools, ok := objectGroups[name]; ok {
			groups = append(groups, ObjectGroup{
				ObjectName: name,
				Tools:      tools,
			})
		}
	}

	// Prepare template data
	data := struct {
		Tools        []ToolData
		ObjectGroups []ObjectGroup
		TotalTools   int
	}{
		Tools:        tools,
		ObjectGroups: groups,
		TotalTools:   totalAPIs,
	}

	// Load and execute template
	// Use refactored template that leverages utility functions
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "tools_refactored.go.tmpl"))
	if err != nil {
		// Fallback to original template if refactored doesn't exist
		tmplContent, err = os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "tools.go.tmpl"))
		if err != nil {
			return fmt.Errorf("failed to read template: %w", err)
		}
	}

	tmpl, err := template.New("tools").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return os.WriteFile(filepath.Join(g.outputPath, "tools.go"), buf.Bytes(), 0o644)
}

func (g *ToolGenerator) generateToolName(objectName string, api API) string {
	// Generate more intuitive tool names based on action and endpoint

	// Handle empty endpoint (CRUD on object itself)
	if api.Endpoint == "" {
		switch api.Method {
		case "GET":
			return fmt.Sprintf("get_%s", g.toSnakeCase(objectName))
		case "POST":
			return fmt.Sprintf("update_%s", g.toSnakeCase(objectName))
		case "DELETE":
			return fmt.Sprintf("delete_%s", g.toSnakeCase(objectName))
		default:
			return fmt.Sprintf("%s_%s", strings.ToLower(api.Method), g.toSnakeCase(objectName))
		}
	}

	// For endpoints, create action-based names
	endpoint := g.toSnakeCase(api.Endpoint)
	objectSnake := g.toSnakeCase(objectName)

	// Special cases for common patterns
	switch {
	case api.Endpoint == "insights" && api.Method == "GET":
		return fmt.Sprintf("get_%s_insights", objectSnake)
	case api.Endpoint == "insights" && api.Method == "POST":
		return fmt.Sprintf("create_%s_insights_report", objectSnake)
	case strings.HasSuffix(api.Endpoint, "s") && api.Method == "GET":
		// Plural endpoints usually list related objects
		return fmt.Sprintf("list_%s_%s", objectSnake, endpoint)
	case api.Method == "POST" && strings.HasSuffix(api.Endpoint, "s"):
		// POST to plural usually creates new items
		singular := strings.TrimSuffix(endpoint, "s")
		return fmt.Sprintf("create_%s_%s", objectSnake, singular)
	case api.Method == "DELETE":
		return fmt.Sprintf("remove_%s_from_%s", endpoint, objectSnake)
	case api.Method == "GET":
		return fmt.Sprintf("get_%s_%s", objectSnake, endpoint)
	case api.Method == "POST":
		return fmt.Sprintf("update_%s_%s", objectSnake, endpoint)
	default:
		return fmt.Sprintf("%s_%s_%s", strings.ToLower(api.Method), objectSnake, endpoint)
	}
}

func (g *ToolGenerator) generateDescription(objectName string, api API) string {
	// Create human-readable description based on endpoint and method
	action := g.getHumanReadableAction(api.Method, api.Endpoint, objectName)

	desc := action

	if api.Return != "" && api.Return != "Object" {
		desc += fmt.Sprintf(" Returns %s.", api.Return)
	}

	// Document required parameters only
	requiredParams := []string{}
	for _, param := range api.Params {
		if param.Required {
			paramDesc := param.Name
			if g.isEnumType(param.Type) {
				paramDesc += " (enum)"
			}
			requiredParams = append(requiredParams, paramDesc)
		}
	}

	if len(requiredParams) > 0 {
		desc += " Required: " + strings.Join(requiredParams, ", ")
	}

	return desc
}

func (g *ToolGenerator) getHumanReadableAction(method, endpoint, objectName string) string {
	// Handle empty endpoint (CRUD on object itself)
	if endpoint == "" {
		switch method {
		case "GET":
			return fmt.Sprintf("Get details of a specific %s", objectName)
		case "POST":
			return fmt.Sprintf("Update a %s", objectName)
		case "DELETE":
			return fmt.Sprintf("Delete a %s", objectName)
		default:
			return fmt.Sprintf("%s a %s", strings.Title(strings.ToLower(method)), objectName)
		}
	}

	// Handle specific endpoints with better descriptions
	endpointLower := strings.ToLower(endpoint)

	// Common patterns
	switch {
	case strings.HasSuffix(endpoint, "s") && method == "GET":
		// Plural endpoints usually list related objects
		return fmt.Sprintf("List %s for this %s", endpoint, objectName)

	case endpoint == "insights":
		if method == "GET" {
			return fmt.Sprintf("Get analytics insights for this %s", objectName)
		}
		return fmt.Sprintf("Generate an insights report for this %s", objectName)

	case endpoint == "copies":
		if method == "GET" {
			return fmt.Sprintf("List copies of this %s", objectName)
		}
		return fmt.Sprintf("Create a copy of this %s", objectName)

	case strings.Contains(endpointLower, "preview"):
		return fmt.Sprintf("Get preview of this %s", objectName)

	case endpoint == "leads":
		return fmt.Sprintf("Get lead information from this %s", objectName)

	case strings.Contains(endpointLower, "targeting"):
		return fmt.Sprintf("Get targeting information for this %s", objectName)

	case method == "POST" && strings.HasPrefix(endpoint, "ad"):
		return fmt.Sprintf("Associate %s with this %s", endpoint, objectName)

	default:
		// Generic description
		switch method {
		case "GET":
			return fmt.Sprintf("Get %s data for this %s", endpoint, objectName)
		case "POST":
			return fmt.Sprintf("Create or update %s for this %s", endpoint, objectName)
		case "DELETE":
			return fmt.Sprintf("Remove %s from this %s", endpoint, objectName)
		default:
			return fmt.Sprintf("%s %s for this %s", strings.Title(strings.ToLower(method)), endpoint, objectName)
		}
	}
}

func (g *ToolGenerator) generateInputSchema(objectName string, api API) map[string]interface{} {
	if api.Method == "GET" || api.Method == "DELETE" {
		return g.generateGETInputSchema(objectName, api)
	}
	return g.generatePOSTInputSchema(objectName, api)
}

func (g *ToolGenerator) generateGETInputSchema(objectName string, api API) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{}

	// Add ID field if needed
	if g.needsObjectID(objectName, api.Endpoint) {
		properties["id"] = map[string]interface{}{
			"type":        "string",
			"description": fmt.Sprintf("%s ID", objectName),
		}
		required = append(required, "id")
	}

	// Add all defined parameters
	for _, param := range api.Params {
		properties[param.Name] = g.generateParamSchema(param)
		if param.Required {
			required = append(required, param.Name)
		}
	}

	// Add common optional parameters for GET requests
	properties["fields"] = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
		},
		"description": "Fields to return",
	}
	properties["limit"] = map[string]interface{}{
		"type":        "integer",
		"description": "Maximum number of results",
	}
	properties["after"] = map[string]interface{}{
		"type":        "string",
		"description": "Cursor for pagination (next page)",
	}
	properties["before"] = map[string]interface{}{
		"type":        "string",
		"description": "Cursor for pagination (previous page)",
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": true, // Allow other params for flexibility
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func (g *ToolGenerator) generatePOSTInputSchema(objectName string, api API) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{}

	// Add ID field if this is updating/acting on existing object
	if g.needsObjectID(objectName, api.Endpoint) {
		properties["id"] = map[string]interface{}{
			"type":        "string",
			"description": fmt.Sprintf("%s ID", objectName),
		}
		required = append(required, "id")
	}

	// All POST parameters from the spec
	for _, param := range api.Params {
		properties[param.Name] = g.generateParamSchema(param)
		if param.Required {
			required = append(required, param.Name)
		}
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false, // POST bodies should be well-defined
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func (g *ToolGenerator) generateParamSchema(param Param) map[string]interface{} {
	schema := make(map[string]interface{})

	// Add description
	desc := param.Name
	if g.isEnumType(param.Type) {
		desc = fmt.Sprintf("%s (enum: %s)", param.Name, param.Type)
	}
	schema["description"] = desc

	// Handle different parameter types
	switch {
	case strings.HasPrefix(param.Type, "list<"):
		innerType := strings.TrimSuffix(strings.TrimPrefix(param.Type, "list<"), ">")
		schema["type"] = "array"
		if innerType == "Object" || strings.Contains(innerType, "map") {
			schema["items"] = map[string]interface{}{
				"type":                 "object",
				"additionalProperties": true,
			}
		} else {
			schema["items"] = map[string]interface{}{
				"type": g.mapBasicType(innerType),
			}
		}

	case strings.HasPrefix(param.Type, "map<"):
		schema["type"] = "object"
		schema["additionalProperties"] = true

	case param.Type == "Object" || param.Type == "object":
		schema["type"] = "object"
		schema["additionalProperties"] = true

	case g.isEnumType(param.Type):
		schema["type"] = "string"

	default:
		schema["type"] = g.mapBasicType(param.Type)
	}

	return schema
}

func (g *ToolGenerator) mapBasicType(fbType string) string {
	switch fbType {
	case "string":
		return "string"
	case "int", "integer", "unsigned int":
		return "integer"
	case "float", "double", "number":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "datetime", "timestamp":
		return "string" // with format: date-time
	default:
		return "string"
	}
}

func (g *ToolGenerator) isEnumType(typeName string) bool {
	return g.enumTypes[typeName]
}

func (g *ToolGenerator) needsObjectID(objectName string, endpoint string) bool {
	// Most endpoints need an object ID
	// Exceptions are typically creation endpoints or root queries
	exceptions := []string{
		"search",
		"me",
		"root",
	}

	for _, exc := range exceptions {
		if strings.Contains(endpoint, exc) {
			return false
		}
	}

	// If the object name is in the endpoint, it probably doesn't need an ID
	if strings.Contains(strings.ToLower(endpoint), strings.ToLower(objectName)) {
		return false
	}

	return true
}

func (g *ToolGenerator) formatGeneratedFiles() error {
	// Get all .go files in the output directory
	files, err := filepath.Glob(filepath.Join(g.outputPath, "*.go"))
	if err != nil {
		return fmt.Errorf("failed to glob files: %w", err)
	}

	// Run go fmt on each file
	for _, file := range files {
		cmd := exec.Command("go", "fmt", file)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("go fmt failed on %s: %w\nOutput: %s", file, err, string(output))
		}

		if len(output) > 0 {
			log.Printf("Formatted %s: %s", file, string(output))
		}
	}

	return nil
}

// toSnakeCase converts CamelCase to snake_case
func (g *ToolGenerator) hasComplexLogic(api API) bool {
	// Handlers that need custom logic:
	// 1. Multiple endpoints in one (e.g., handling both collection and item)
	// 2. Special parameter handling beyond standard patterns
	// 3. Complex return value processing

	// For now, consider all handlers as potentially needing custom logic
	// This can be refined based on actual API patterns
	return false // We'll start with assuming most can use standard handlers
}

func (g *ToolGenerator) hasQueryParams(api API) bool {
	// Check if POST/PUT methods have parameters that should go in query string
	if api.Method != "POST" && api.Method != "PUT" {
		return false
	}

	// Some Facebook APIs require certain params in query even for POST
	// This would need to be determined from API documentation
	return false
}

func (g *ToolGenerator) mapParamType(fbType string) string {
	// Map Facebook types to JSON schema types
	switch strings.ToLower(fbType) {
	case "string", "enum", "datetime", "url":
		return "string"
	case "int", "integer", "unsigned int":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "list", "array":
		return "array"
	case "object", "map":
		return "object"
	default:
		// Check if it's an enum type
		if g.isEnumType(fbType) {
			return "string"
		}
		return "string" // Default to string
	}
}

func (g *ToolGenerator) toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Add underscore before uppercase letter if previous char is lowercase
			if prev := s[i-1]; prev >= 'a' && prev <= 'z' {
				result = append(result, '_')
			}
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// toCamelCase converts snake_case to CamelCase
func (g *ToolGenerator) toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// convertToTypedParams converts API params to TypedParam for Go structs
func (g *ToolGenerator) convertToTypedParams(params []Param, method string) []TypedParam {
	var typedParams []TypedParam
	seen := make(map[string]bool)

	// Add common GET parameters if GET method
	if method == "GET" {
		typedParams = append(typedParams, TypedParam{
			GoName:        "Fields",
			GoType:        "[]string",
			JSONName:      "fields",
			JSONTag:       "omitempty",
			JSONSchemaTag: "description=Fields to return",
		})
		seen["fields"] = true

		typedParams = append(typedParams, TypedParam{
			GoName:        "Limit",
			GoType:        "int",
			JSONName:      "limit",
			JSONTag:       "omitempty",
			JSONSchemaTag: "description=Maximum number of results,minimum=1,maximum=100",
		})
		seen["limit"] = true

		typedParams = append(typedParams, TypedParam{
			GoName:        "After",
			GoType:        "string",
			JSONName:      "after",
			JSONTag:       "omitempty",
			JSONSchemaTag: "description=Cursor for pagination (next page)",
		})
		seen["after"] = true

		typedParams = append(typedParams, TypedParam{
			GoName:        "Before",
			GoType:        "string",
			JSONName:      "before",
			JSONTag:       "omitempty",
			JSONSchemaTag: "description=Cursor for pagination (previous page)",
		})
		seen["before"] = true
	}

	// Convert API params, skipping duplicates
	for _, param := range params {
		// Skip if already added as common param
		if seen[param.Name] {
			continue
		}

		goType := g.mapParamToGoType(param.Type, param.Name)
		goName := g.toCamelCase(param.Name)

		typedParam := TypedParam{
			GoName:        goName,
			GoType:        goType,
			JSONName:      param.Name,
			JSONSchemaTag: g.generateJSONSchemaTag(param, goType),
		}

		// Add omitempty for optional params
		if !param.Required {
			typedParam.JSONTag = "omitempty"
		}

		typedParams = append(typedParams, typedParam)
		seen[param.Name] = true
	}

	return typedParams
}

// generateJSONSchemaTag generates a jsonschema tag for a parameter
func (g *ToolGenerator) generateJSONSchemaTag(param Param, goType string) string {
	var tags []string

	// Add description
	desc := g.generateParamDescription(param.Name, param.Type)
	if desc != "" {
		tags = append(tags, fmt.Sprintf("description=%s", desc))
	}

	// Add required if needed
	if param.Required {
		tags = append(tags, "required")
	}

	// Add type-specific validations
	switch {
	case strings.Contains(param.Name, "_id") || param.Name == "id":
		// Facebook IDs are numeric strings
		tags = append(tags, `pattern=^[0-9]+$`)

	case strings.Contains(goType, "int") || param.Type == "integer":
		// Add reasonable bounds for common integer fields
		if strings.Contains(param.Name, "age") {
			tags = append(tags, "minimum=13", "maximum=100")
		} else if strings.Contains(param.Name, "budget") || strings.Contains(param.Name, "amount") {
			tags = append(tags, "minimum=1")
		}

	case param.Type == "string" && g.isEnumType(param.Type):
		// Add enum values if available
		if enumValues := g.getEnumValues(param.Type); len(enumValues) > 0 {
			for _, v := range enumValues {
				tags = append(tags, fmt.Sprintf("enum=%s", v))
			}
		}

	case param.Name == "status":
		// Common status values
		tags = append(tags, "enum=ACTIVE", "enum=PAUSED", "enum=DELETED", "enum=ARCHIVED")

	case strings.Contains(param.Name, "url"):
		// URL fields
		tags = append(tags, "format=uri")

	case param.Type == "datetime" || param.Type == "timestamp":
		// Date/time fields
		tags = append(tags, "format=date-time")
	}

	if len(tags) > 0 {
		return strings.Join(tags, ",")
	}
	return ""
}

// generateParamDescription generates a human-readable description for a parameter
func (g *ToolGenerator) generateParamDescription(paramName, paramType string) string {
	// Convert snake_case to human readable
	words := strings.Split(paramName, "_")
	for i, word := range words {
		if word != "" {
			// Handle common abbreviations
			switch strings.ToLower(word) {
			case "id":
				words[i] = "ID"
			case "url":
				words[i] = "URL"
			case "api":
				words[i] = "API"
			default:
				words[i] = strings.Title(word)
			}
		}
	}

	desc := strings.Join(words, " ")

	// Add context based on field patterns
	switch {
	case paramName == "id":
		return "ID"
	case strings.HasSuffix(paramName, "_id"):
		return fmt.Sprintf("ID of the %s", strings.TrimSuffix(desc, " ID"))
	case paramName == "name":
		return "Name"
	case paramName == "status":
		return "Status"
	case strings.Contains(paramName, "created"):
		return "When created"
	case strings.Contains(paramName, "updated"):
		return "When last updated"
	default:
		return desc
	}
}

// getEnumValues returns possible enum values for a type
func (g *ToolGenerator) getEnumValues(typeName string) []string {
	// This would be populated from enum_types.json
	// For now, return empty - the actual implementation would look up the values
	return []string{}
}

// mapParamToGoType maps Facebook parameter types to Go types
func (g *ToolGenerator) mapParamToGoType(fbType string, paramName string) string {
	// Special case for known complex fields that use Object in API specs
	if paramName == "adlabels" && fbType == "list<Object>" {
		return "[]*AdLabel"
	}
	if paramName == "targeting" && fbType == "Targeting" {
		return "*Targeting"
	}
	if paramName == "targeting_spec" && fbType == "Targeting" {
		return "*Targeting"
	}
	if paramName == "promoted_object" && (fbType == "Object" || fbType == "AdPromotedObject") {
		return "*AdPromotedObject"
	}
	if paramName == "creative" && fbType == "AdCreative" {
		return "*AdCreative"
	}
	if paramName == "adset_spec" && fbType == "AdSet" {
		return "*AdSet"
	}
	if paramName == "creative_parameters" && fbType == "AdCreative" {
		return "*AdCreative"
	}
	switch {
	case strings.HasPrefix(fbType, "list<"):
		innerType := strings.TrimSuffix(strings.TrimPrefix(fbType, "list<"), ">")

		// Check if inner type is a known complex type (like AdLabel)
		if g.isComplexType(innerType) {
			return fmt.Sprintf("[]*%s", innerType)
		}

		// Check if inner type is an enum
		if g.isEnumType(innerType) {
			// For now, use string until we have proper enum type names
			return "[]string"
		}

		// Handle basic types
		switch innerType {
		case "string":
			return "[]string"
		case "int", "integer":
			return "[]int"
		case "Object", "object":
			return "[]map[string]interface{}"
		default:
			if strings.Contains(innerType, "map") {
				return "[]map[string]interface{}"
			}
			return "[]interface{}"
		}

	case strings.HasPrefix(fbType, "map"):
		return "map[string]interface{}"

	case fbType == "Object" || fbType == "object":
		return "map[string]interface{}"

	default:
		// Check if it's a known complex type
		if g.isComplexType(fbType) {
			return "*" + fbType
		}

		// Check if it's an enum
		if g.isEnumType(fbType) {
			// For now, return string for enums
			// TODO: return proper enum type name when available
			return "string"
		}

		// Basic types
		switch fbType {
		case "string":
			return "string"
		case "int", "integer", "unsigned int":
			return "int"
		case "float", "double":
			return "float64"
		case "bool", "boolean":
			return "bool"
		case "datetime", "timestamp":
			return "string" // or time.Time with custom unmarshaling
		default:
			return "interface{}"
		}
	}
}

// convertToSchemaParams creates SchemaParam entries for mcp.NewTool
func (g *ToolGenerator) convertToSchemaParams(params []Param, needsID bool, method string) []SchemaParam {
	var schemaParams []SchemaParam
	seen := make(map[string]bool)

	// Add common GET parameters if GET method
	if method == "GET" {
		schemaParams = append(schemaParams, SchemaParam{
			Name:        "fields",
			Type:        "array",
			Required:    false,
			Description: "Fields to return",
			SchemaType:  "array",
			ItemsType:   "string",
			JSONName:    "fields",
		})
		seen["fields"] = true

		schemaParams = append(schemaParams, SchemaParam{
			Name:        "limit",
			Type:        "number",
			Required:    false,
			Description: "Maximum number of results",
			SchemaType:  "number",
			JSONName:    "limit",
		})
		seen["limit"] = true

		schemaParams = append(schemaParams, SchemaParam{
			Name:        "after",
			Type:        "string",
			Required:    false,
			Description: "Cursor for pagination (next page)",
			SchemaType:  "string",
			JSONName:    "after",
		})
		seen["after"] = true

		schemaParams = append(schemaParams, SchemaParam{
			Name:        "before",
			Type:        "string",
			Required:    false,
			Description: "Cursor for pagination (previous page)",
			SchemaType:  "string",
			JSONName:    "before",
		})
		seen["before"] = true
	}

	// Convert API params, skipping duplicates
	for _, param := range params {
		// Skip if already added as common param
		if seen[param.Name] {
			continue
		}

		schemaType, itemsType := g.mapParamToSchemaType(param.Type)
		desc := param.Name
		if g.isEnumType(param.Type) {
			desc = fmt.Sprintf("%s (enum: %s)", param.Name, param.Type)
		}

		schemaParams = append(schemaParams, SchemaParam{
			Name:        param.Name,
			Type:        param.Type,
			Required:    param.Required,
			Description: desc,
			SchemaType:  schemaType,
			ItemsType:   itemsType,
			JSONName:    param.Name,
		})
		seen[param.Name] = true
	}

	return schemaParams
}

// mapParamToSchemaType maps Facebook types to JSON schema types
func (g *ToolGenerator) mapParamToSchemaType(fbType string) (schemaType string, itemsType string) {
	switch {
	case g.isComplexType(fbType):
		// Complex types (like Targeting, AdLabel, etc.) should be objects
		return "object", ""

	case strings.HasPrefix(fbType, "list<"):
		innerType := strings.TrimSuffix(strings.TrimPrefix(fbType, "list<"), ">")
		if innerType == "Object" || strings.Contains(innerType, "map") || g.isComplexType(innerType) {
			return "array", "object"
		}
		return "array", "string"

	case strings.HasPrefix(fbType, "map") || fbType == "Object" || fbType == "object":
		return "object", ""

	case fbType == "string" || g.isEnumType(fbType):
		return "string", ""

	case fbType == "int" || fbType == "integer" || fbType == "unsigned int":
		return "number", ""

	case fbType == "float" || fbType == "double":
		return "number", ""

	case fbType == "bool" || fbType == "boolean":
		return "boolean", ""

	default:
		return "string", ""
	}
}

// isComplexType determines if a type requires a struct definition
func (g *ToolGenerator) isComplexType(typeName string) bool {
	// Check if it's in our automatically detected complex types
	if g.complexTypes != nil {
		return g.complexTypes[typeName]
	}

	// Fallback to known complex types
	complexTypes := map[string]bool{
		"AdLabel":          true,
		"PromotedObject":   true,
		"AdPromotedObject": true,
		"Targeting":        true,
		"AdCreative":       true,
		// Add more as needed
	}

	return complexTypes[typeName]
}

// identifyComplexTypes analyzes field specs to identify which types are complex objects
func (g *ToolGenerator) identifyComplexTypes() {
	g.complexTypes = make(map[string]bool)

	// Any type that appears as a field type and has its own field spec is a complex type
	for objectName := range g.fieldSpecs {
		g.complexTypes[objectName] = true
	}

	// Also check field references to identify complex types
	for _, spec := range g.fieldSpecs {
		for _, field := range spec.Fields {
			fieldType := g.extractBaseType(field.Type)
			// If this field type has its own spec, it's a complex type
			if _, hasSpec := g.fieldSpecs[fieldType]; hasSpec {
				g.complexTypes[fieldType] = true
			}
		}
	}
}

// extractBaseType extracts the base type from complex type definitions
func (g *ToolGenerator) extractBaseType(fieldType string) string {
	// Handle list<Type>
	if strings.HasPrefix(fieldType, "list<") && strings.HasSuffix(fieldType, ">") {
		return strings.TrimSuffix(strings.TrimPrefix(fieldType, "list<"), ">")
	}

	// Handle map<string, Type>
	if strings.HasPrefix(fieldType, "map<") && strings.HasSuffix(fieldType, ">") {
		// Extract the value type from map<string, Type>
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(fieldType, "map<"), ">"), ",")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}

	return fieldType
}

// canUseSchemaGeneration checks if we can use schema-based generation for this API
func (g *ToolGenerator) canUseSchemaGeneration(objectName string, api API) bool {
	// Check if we have field specs for complex parameters
	for _, param := range api.Params {
		if g.isComplexType(param.Type) {
			if _, hasFieldSpec := g.fieldSpecs[param.Type]; !hasFieldSpec {
				return false
			}
		}
	}
	return true
}

// generateSchemaForAPI generates JSON schema for an API endpoint using reflection
func (g *ToolGenerator) generateSchemaForAPI(objectName string, api API) ([]byte, []SchemaParam, error) {
	if g.schemaGen == nil {
		return nil, nil, fmt.Errorf("schema generator not initialized")
	}

	// Build a struct representation of the API parameters
	structFields := make(map[string]interface{})

	// Add ID field if needed
	if g.needsObjectID(objectName, api.Endpoint) {
		structFields["id"] = struct {
			Type        string `json:"type"`
			Description string `json:"description"`
			Pattern     string `json:"pattern,omitempty"`
		}{
			Type:        "string",
			Description: fmt.Sprintf("%s ID", objectName),
			Pattern:     "^[0-9]+$",
		}
	}

	// Build schema properties from parameters
	properties := make(map[string]interface{})
	required := []string{}

	for _, param := range api.Params {
		propSchema := g.generateParamSchema(param)
		properties[param.Name] = propSchema
		if param.Required {
			required = append(required, param.Name)
		}
	}

	// Add common GET parameters if applicable
	if api.Method == "GET" {
		properties["fields"] = map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "Fields to return",
		}
		properties["limit"] = map[string]interface{}{
			"type":        "integer",
			"description": "Maximum number of results",
		}
	}

	// Create the schema
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	// Convert to JSON
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, nil, err
	}

	// Generate schema params
	schemaParams := g.convertToSchemaParams(api.Params, g.needsObjectID(objectName, api.Endpoint), api.Method)

	return schemaJSON, schemaParams, nil
}

// generateRawSchema generates a complete JSON schema for use with mcp.NewToolWithRawSchema
func (g *ToolGenerator) generateRawSchema(objectName string, api API) (string, error) {
	// Initialize cycle detection for this schema generation
	visited := make(map[string]bool)

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           make(map[string]interface{}),
		"required":             []string{},
		"additionalProperties": false,
	}

	properties := schema["properties"].(map[string]interface{})
	required := schema["required"].([]string)

	// Add ID field if needed
	if g.needsObjectID(objectName, api.Endpoint) {
		properties["id"] = map[string]interface{}{
			"type":        "string",
			"description": fmt.Sprintf("%s ID", objectName),
			"pattern":     "^[0-9]+$",
		}
		required = append(required, "id")
	}

	// Process each parameter with enhanced complex type handling
	for _, param := range api.Params {
		propSchema, err := g.generateComplexParamSchemaWithCycleDetection(param, visited)
		if err != nil {
			return "", fmt.Errorf("failed to generate schema for param %s: %w", param.Name, err)
		}
		properties[param.Name] = propSchema

		if param.Required {
			required = append(required, param.Name)
		}
	}

	// Add common GET parameters
	if api.Method == "GET" {
		g.addCommonGetParams(properties)
	}

	// Update required array
	schema["required"] = required

	// Marshal to JSON
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	return string(schemaJSON), nil
}

// generateComplexParamSchemaWithCycleDetection generates schema for a parameter with cycle detection
func (g *ToolGenerator) generateComplexParamSchemaWithCycleDetection(param Param, visited map[string]bool) (map[string]interface{}, error) {
	schema := make(map[string]interface{})

	// Add description
	desc := g.generateParamDescription(param.Name, param.Type)
	if g.isEnumType(param.Type) {
		desc = fmt.Sprintf("%s (enum: %s)", desc, param.Type)
	}
	schema["description"] = desc

	// Handle different parameter types
	switch {
	case strings.HasPrefix(param.Type, "list<"):
		return g.generateListSchemaWithCycleDetection(param, visited)

	case strings.HasPrefix(param.Type, "map<"):
		return g.generateMapSchemaWithCycleDetection(param, visited)

	case param.Type == "Object" || param.Type == "object":
		schema["type"] = "object"
		schema["additionalProperties"] = true

	case g.isComplexType(param.Type):
		return g.generateComplexTypeSchemaWithCycleDetection(param.Type, param.Name, visited)

	case g.isEnumType(param.Type):
		schema["type"] = "string"
		// Add enum values if available
		if enumValues := g.getEnumValues(param.Type); len(enumValues) > 0 {
			schema["enum"] = enumValues
		}

	default:
		schema["type"] = g.mapBasicType(param.Type)
		g.addTypeSpecificValidation(schema, param)
	}

	return schema, nil
}

// generateComplexParamSchema generates schema for a parameter with recursive complex type handling (legacy)
func (g *ToolGenerator) generateComplexParamSchema(param Param) (map[string]interface{}, error) {
	visited := make(map[string]bool)
	return g.generateComplexParamSchemaWithCycleDetection(param, visited)
}

// generateListSchemaWithCycleDetection generates schema for list types with cycle detection
func (g *ToolGenerator) generateListSchemaWithCycleDetection(param Param, visited map[string]bool) (map[string]interface{}, error) {
	innerType := strings.TrimSuffix(strings.TrimPrefix(param.Type, "list<"), ">")
	schema := map[string]interface{}{
		"type":        "array",
		"description": g.generateParamDescription(param.Name, param.Type),
	}

	// Generate items schema based on inner type
	if g.isComplexType(innerType) {
		itemSchema, err := g.generateComplexTypeSchemaWithCycleDetection(innerType, param.Name, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for list item type %s: %w", innerType, err)
		}
		schema["items"] = itemSchema
	} else if innerType == "Object" || strings.Contains(innerType, "map") {
		schema["items"] = map[string]interface{}{
			"type":                 "object",
			"additionalProperties": true,
		}
	} else {
		schema["items"] = map[string]interface{}{
			"type": g.mapBasicType(innerType),
		}
	}

	return schema, nil
}

// generateListSchema generates schema for list types (legacy)
func (g *ToolGenerator) generateListSchema(param Param) (map[string]interface{}, error) {
	visited := make(map[string]bool)
	return g.generateListSchemaWithCycleDetection(param, visited)
}

// generateMapSchemaWithCycleDetection generates schema for map types with cycle detection
func (g *ToolGenerator) generateMapSchemaWithCycleDetection(param Param, visited map[string]bool) (map[string]interface{}, error) {
	schema := map[string]interface{}{
		"type":                 "object",
		"description":          g.generateParamDescription(param.Name, param.Type),
		"additionalProperties": true,
	}

	// Extract value type from map<key, value>
	if strings.HasPrefix(param.Type, "map<") && strings.HasSuffix(param.Type, ">") {
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(param.Type, "map<"), ">"), ",")
		if len(parts) == 2 {
			valueType := strings.TrimSpace(parts[1])
			if g.isComplexType(valueType) {
				valueSchema, err := g.generateComplexTypeSchemaWithCycleDetection(valueType, param.Name, visited)
				if err != nil {
					return nil, fmt.Errorf("failed to generate schema for map value type %s: %w", valueType, err)
				}
				schema["additionalProperties"] = valueSchema
			} else {
				schema["additionalProperties"] = map[string]interface{}{
					"type": g.mapBasicType(valueType),
				}
			}
		}
	}

	return schema, nil
}

// generateMapSchema generates schema for map types (legacy)
func (g *ToolGenerator) generateMapSchema(param Param) (map[string]interface{}, error) {
	visited := make(map[string]bool)
	return g.generateMapSchemaWithCycleDetection(param, visited)
}

// generateComplexTypeSchemaWithCycleDetection generates schema for complex types with cycle detection
func (g *ToolGenerator) generateComplexTypeSchemaWithCycleDetection(typeName string, paramName string, visited map[string]bool) (map[string]interface{}, error) {
	// Check for circular reference
	if visited[typeName] {
		log.Printf("Circular reference detected for type %s, using reference", typeName)
		return map[string]interface{}{
			"type":                 "object",
			"description":          fmt.Sprintf("%s object (circular reference)", typeName),
			"additionalProperties": true,
		}, nil
	}

	// Mark as visited
	visited[typeName] = true
	defer func() {
		delete(visited, typeName)
	}()

	schema := map[string]interface{}{
		"type":        "object",
		"description": fmt.Sprintf("%s object", typeName),
	}

	// Check if we have field specs for this type
	if fieldSpec, exists := g.fieldSpecs[typeName]; exists {
		properties := make(map[string]interface{})
		required := []string{}

		for _, field := range fieldSpec.Fields {
			fieldSchema, err := g.generateFieldSchemaWithCycleDetection(field, visited)
			if err != nil {
				return nil, fmt.Errorf("failed to generate schema for field %s.%s: %w", typeName, field.Name, err)
			}
			properties[field.Name] = fieldSchema

			// Check if field is required based on common patterns
			if g.isFieldRequired(field.Name, typeName) {
				required = append(required, field.Name)
			}
		}

		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}
		schema["additionalProperties"] = false
	} else {
		// Fallback to generic object for unknown complex types
		log.Printf("Warning: No field spec found for complex type %s, using generic object", typeName)
		schema["additionalProperties"] = true
	}

	return schema, nil
}

// generateComplexTypeSchema generates schema for complex types using field definitions (legacy)
func (g *ToolGenerator) generateComplexTypeSchema(typeName string, paramName string) (map[string]interface{}, error) {
	visited := make(map[string]bool)
	return g.generateComplexTypeSchemaWithCycleDetection(typeName, paramName, visited)
}

// generateFieldSchemaWithCycleDetection generates schema for a field with cycle detection
func (g *ToolGenerator) generateFieldSchemaWithCycleDetection(field Field, visited map[string]bool) (map[string]interface{}, error) {
	schema := make(map[string]interface{})

	// Add description
	schema["description"] = g.generateParamDescription(field.Name, field.Type)

	// Handle different field types recursively
	switch {
	case strings.HasPrefix(field.Type, "list<"):
		innerType := strings.TrimSuffix(strings.TrimPrefix(field.Type, "list<"), ">")
		schema["type"] = "array"
		if g.isComplexType(innerType) {
			itemSchema, err := g.generateComplexTypeSchemaWithCycleDetection(innerType, field.Name, visited)
			if err != nil {
				return nil, err
			}
			schema["items"] = itemSchema
		} else {
			schema["items"] = map[string]interface{}{
				"type": g.mapBasicType(innerType),
			}
		}

	case strings.HasPrefix(field.Type, "map<"):
		schema["type"] = "object"
		schema["additionalProperties"] = true

	case field.Type == "Object" || field.Type == "object":
		schema["type"] = "object"
		schema["additionalProperties"] = true

	case g.isComplexType(field.Type):
		return g.generateComplexTypeSchemaWithCycleDetection(field.Type, field.Name, visited)

	case g.isEnumType(field.Type):
		schema["type"] = "string"
		if enumValues := g.getEnumValues(field.Type); len(enumValues) > 0 {
			schema["enum"] = enumValues
		}

	default:
		schema["type"] = g.mapBasicType(field.Type)
	}

	return schema, nil
}

// generateFieldSchema generates schema for a field from field specifications (legacy)
func (g *ToolGenerator) generateFieldSchema(field Field) (map[string]interface{}, error) {
	visited := make(map[string]bool)
	return g.generateFieldSchemaWithCycleDetection(field, visited)
}

// addCommonGetParams adds common GET parameters to the schema properties
func (g *ToolGenerator) addCommonGetParams(properties map[string]interface{}) {
	properties["fields"] = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
		},
		"description": "Fields to return",
	}
	properties["limit"] = map[string]interface{}{
		"type":        "integer",
		"description": "Maximum number of results",
		"minimum":     1,
		"maximum":     100,
	}
	properties["after"] = map[string]interface{}{
		"type":        "string",
		"description": "Cursor for pagination (next page)",
	}
	properties["before"] = map[string]interface{}{
		"type":        "string",
		"description": "Cursor for pagination (previous page)",
	}
}

// addTypeSpecificValidation adds type-specific validation to schema
func (g *ToolGenerator) addTypeSpecificValidation(schema map[string]interface{}, param Param) {
	switch {
	case strings.Contains(param.Name, "_id") || param.Name == "id":
		schema["pattern"] = "^[0-9]+$"

	case param.Type == "integer" || param.Type == "int":
		if strings.Contains(param.Name, "age") {
			schema["minimum"] = 13
			schema["maximum"] = 100
		} else if strings.Contains(param.Name, "budget") || strings.Contains(param.Name, "amount") {
			schema["minimum"] = 1
		}

	case param.Type == "string":
		if strings.Contains(param.Name, "url") {
			schema["format"] = "uri"
		} else if param.Type == "datetime" || param.Type == "timestamp" {
			schema["format"] = "date-time"
		}
	}
}

// isFieldRequired determines if a field should be marked as required based on common patterns
func (g *ToolGenerator) isFieldRequired(fieldName, typeName string) bool {
	// Common required fields
	requiredFields := map[string]bool{
		"id":   true,
		"name": true,
	}

	// Type-specific required fields
	switch typeName {
	case "AdSet":
		switch fieldName {
		case "campaign_id", "optimization_goal", "billing_event", "targeting":
			return true
		}
	case "Campaign":
		switch fieldName {
		case "objective", "status":
			return true
		}
	case "Ad":
		switch fieldName {
		case "adset_id", "creative", "status":
			return true
		}
	case "Targeting":
		switch fieldName {
		case "geo_locations":
			return true
		}
	}

	return requiredFields[fieldName]
}

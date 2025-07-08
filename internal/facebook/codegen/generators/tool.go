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
	specs      map[string]*ToolSpec // Object name -> spec
	enumTypes  map[string]bool
	outputPath string
}

func NewToolGenerator(outputPath string) *ToolGenerator {
	return &ToolGenerator{
		specs:      make(map[string]*ToolSpec),
		enumTypes:  make(map[string]bool),
		outputPath: outputPath,
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

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") && entry.Name() != "enum_types.json" {
			specPath := filepath.Join(specsDir, entry.Name())
			data, err := os.ReadFile(specPath)
			if err != nil {
				log.Printf("Warning: failed to read %s: %v", specPath, err)
				continue
			}

			var spec ToolSpec
			if err := json.Unmarshal(data, &spec); err != nil {
				log.Printf("Warning: failed to parse %s: %v", specPath, err)
				continue
			}

			objectName := strings.TrimSuffix(entry.Name(), ".json")
			g.specs[objectName] = &spec
		}
	}

	log.Printf("Loaded %d API specs with endpoints", len(g.specs))
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
	type ToolData struct {
		ObjectName  string
		Method      string
		Endpoint    string
		Return      string
		ToolName    string
		HandlerName string
		Description string
		InputSchema string // JSON string
		NeedsID     bool
		Params      []Param
	}

	var tools []ToolData

	// Process APIs for this object
	for _, api := range spec.APIs {
		toolName := g.generateToolName(objectName, api)
		handlerName := toolName + "Handler"

		// Generate input schema based on method
		inputSchema := g.generateInputSchema(objectName, api)
		inputSchemaJSON, _ := json.Marshal(inputSchema)

		tools = append(tools, ToolData{
			ObjectName:  objectName,
			Method:      api.Method,
			Endpoint:    api.Endpoint,
			Return:      api.Return,
			ToolName:    toolName,
			HandlerName: handlerName,
			Description: strings.ReplaceAll(g.generateDescription(objectName, api), `"`, `\"`),
			InputSchema: string(inputSchemaJSON),
			NeedsID:     g.needsObjectID(objectName, api.Endpoint),
			Params:      api.Params,
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
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "object_tools.go.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
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
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "tools_common.go.tmpl"))
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
	type ToolData struct {
		ObjectName  string
		Method      string
		Endpoint    string
		Return      string
		ToolName    string
		HandlerName string
		Description string
		InputSchema string // JSON string
		NeedsID     bool
		Params      []Param
	}

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

			tools = append(tools, ToolData{
				ObjectName:  objectName,
				Method:      api.Method,
				Endpoint:    api.Endpoint,
				Return:      api.Return,
				ToolName:    toolName,
				HandlerName: handlerName,
				Description: strings.ReplaceAll(g.generateDescription(objectName, api), `"`, `\"`),
				InputSchema: string(inputSchemaJSON),
				NeedsID:     g.needsObjectID(objectName, api.Endpoint),
				Params:      api.Params,
			})
		}
	}

	// Prepare template data
	data := struct {
		Tools      []ToolData
		TotalTools int
	}{
		Tools:      tools,
		TotalTools: totalAPIs,
	}

	// Load and execute template
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "tools.go.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
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
	// Format: ObjectName_METHOD_endpoint
	endpoint := strings.ReplaceAll(api.Endpoint, "/", "_")
	return fmt.Sprintf("%s_%s_%s", objectName, api.Method, endpoint)
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
	// Run go fmt on the output directory
	cmd := exec.Command("go", "fmt", g.outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go fmt failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		log.Printf("Formatted files: %s", string(output))
	}

	return nil
}

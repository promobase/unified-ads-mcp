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

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type APISpec struct {
	Fields []Field `json:"fields"`
}

type FieldGenerator struct {
	specs             map[string]*APISpec
	enumTypes         map[string]bool
	outputPath        string
	useSchemaTemplate bool // Flag to use template with jsonschema tags
}

func NewFieldGenerator(outputPath string) *FieldGenerator {
	return &FieldGenerator{
		specs:             make(map[string]*APISpec),
		enumTypes:         make(map[string]bool),
		outputPath:        outputPath,
		useSchemaTemplate: false,
	}
}

// EnableSchemaGeneration enables generation with jsonschema tags
func (g *FieldGenerator) EnableSchemaGeneration() {
	g.useSchemaTemplate = true
}

func (g *FieldGenerator) LoadEnumTypes(path string) error {
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

	log.Printf("Loaded %d enum types for field generation", len(g.enumTypes))
	return nil
}

func (g *FieldGenerator) LoadAPISpecs(specsDir string) error {
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

			var spec APISpec
			if err := json.Unmarshal(data, &spec); err != nil {
				log.Printf("Warning: failed to parse %s: %v", specPath, err)
				continue
			}

			objectName := strings.TrimSuffix(entry.Name(), ".json")
			g.specs[objectName] = &spec
		}
	}

	log.Printf("Loaded %d API specs with fields", len(g.specs))
	return nil
}

func (g *FieldGenerator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(g.outputPath, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate fields file
	if err := g.generateFields(); err != nil {
		return fmt.Errorf("failed to generate fields: %w", err)
	}

	// Format the generated file
	if err := g.formatGeneratedFile(); err != nil {
		return fmt.Errorf("failed to format generated file: %w", err)
	}

	return nil
}

func (g *FieldGenerator) generateFields() error {
	// Sort object names for consistent output
	var objectNames []string
	for name := range g.specs {
		objectNames = append(objectNames, name)
	}
	sort.Strings(objectNames)

	// Sort fields within each spec for consistent output
	for _, spec := range g.specs {
		if len(spec.Fields) > 0 {
			sort.Slice(spec.Fields, func(i, j int) bool {
				return spec.Fields[i].Name < spec.Fields[j].Name
			})
		}
	}

	// Create ordered slice for template
	type ObjectSpec struct {
		Name string
		Spec *APISpec
	}
	var orderedObjects []ObjectSpec
	for _, name := range objectNames {
		orderedObjects = append(orderedObjects, ObjectSpec{
			Name: name,
			Spec: g.specs[name],
		})
	}

	// Calculate total fields
	totalFields := 0
	for _, spec := range g.specs {
		totalFields += len(spec.Fields)
	}

	// Prepare template data
	data := struct {
		Objects      []ObjectSpec
		TotalObjects int
		TotalFields  int
	}{
		Objects:      orderedObjects,
		TotalObjects: len(objectNames),
		TotalFields:  totalFields,
	}

	// Load and execute template
	tmplFile := "fields.go.tmpl"
	if g.useSchemaTemplate {
		tmplFile = "fields_with_schema.go.tmpl"
	}
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", tmplFile))
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	funcMap := template.FuncMap{
		"toGoFieldName": toGoFieldName,
		"mapFieldType":  g.mapFieldType,
	}

	// Add schema generation function if enabled
	if g.useSchemaTemplate {
		funcMap["generateJSONSchemaTag"] = func(field Field, objectName string) string {
			return g.GenerateJSONSchemaTag(field, objectName)
		}
	}

	tmpl, err := template.New("fields").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return os.WriteFile(filepath.Join(g.outputPath, "fields.go"), buf.Bytes(), 0644)
}

func (g *FieldGenerator) mapFieldType(fieldType string) string {
	// Handle generic types
	if strings.HasPrefix(fieldType, "list<") && strings.HasSuffix(fieldType, ">") {
		innerType := strings.TrimSuffix(strings.TrimPrefix(fieldType, "list<"), ">")
		return "[]" + g.mapFieldType(innerType)
	}

	if strings.HasPrefix(fieldType, "map<") && strings.HasSuffix(fieldType, ">") {
		// Parse map type like "map<string, unsigned int>"
		inner := strings.TrimSuffix(strings.TrimPrefix(fieldType, "map<"), ">")
		parts := strings.SplitN(inner, ",", 2)
		if len(parts) == 2 {
			keyType := g.mapFieldType(strings.TrimSpace(parts[0]))
			valueType := g.mapFieldType(strings.TrimSpace(parts[1]))

			// Maps can only have comparable types as keys
			// If key type is not a basic type, use string
			if strings.Contains(keyType, "map") || strings.Contains(keyType, "[]") || strings.Contains(keyType, "interface{}") {
				keyType = "string"
			}

			return fmt.Sprintf("map[%s]%s", keyType, valueType)
		}
		return "map[string]interface{}"
	}

	// Check if it's an enum
	if g.isEnumType(fieldType) {
		return ToGoTypeName(fieldType)
	}

	// Check if it's a known object type
	if g.isKnownObjectType(fieldType) {
		return "*" + fieldType // Pointer to avoid circular dependencies
	}

	// Map basic types
	switch fieldType {
	case "string":
		return "string"
	case "int", "integer":
		return "int"
	case "unsigned int", "uint":
		return "uint"
	case "float", "double", "number":
		return "float64"
	case "bool", "boolean":
		return "bool"
	case "datetime", "timestamp":
		return "time.Time"
	case "Object", "object", "map", "Map":
		return "map[string]interface{}"
	default:
		// If we don't recognize the type, treat it as interface{}
		return "interface{}"
	}
}

func (g *FieldGenerator) isEnumType(typeName string) bool {
	return g.enumTypes[typeName]
}

func (g *FieldGenerator) isKnownObjectType(typeName string) bool {
	_, exists := g.specs[typeName]
	return exists
}

// toGoFieldName converts field names to Go field names
func toGoFieldName(s string) string {
	// Replace dots with underscores
	s = strings.ReplaceAll(s, ".", "_")

	// Handle fields that start with numbers
	if len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
		s = "X" + s // Prefix with X for fields starting with numbers
	}

	// Split by underscores
	parts := strings.Split(s, "_")

	// Capitalize each part
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.Title(part)
		}
	}

	// Handle common acronyms
	result := strings.Join(parts, "")
	result = strings.ReplaceAll(result, "Id", "ID")
	result = strings.ReplaceAll(result, "Url", "URL")
	result = strings.ReplaceAll(result, "Api", "API")
	result = strings.ReplaceAll(result, "Ios", "IOS")
	result = strings.ReplaceAll(result, "Http", "HTTP")
	result = strings.ReplaceAll(result, "Https", "HTTPS")

	// Ensure it starts with a capital letter
	if len(result) > 0 && result[0] >= 'a' && result[0] <= 'z' {
		result = strings.ToUpper(string(result[0])) + result[1:]
	}

	return result
}

func (g *FieldGenerator) formatGeneratedFile() error {
	// Run go fmt on the generated fields.go file
	fieldsFile := filepath.Join(g.outputPath, "fields.go")
	cmd := exec.Command("go", "fmt", fieldsFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go fmt failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		log.Printf("Formatted: %s", string(output))
	}

	return nil
}

// GenerateJSONSchemaTag generates jsonschema tag for a field
func (g *FieldGenerator) GenerateJSONSchemaTag(field Field, objectName string) string {
	var tags []string

	// Add description based on field name and object
	desc := g.generateFieldDescription(field.Name, objectName)
	if desc != "" {
		tags = append(tags, fmt.Sprintf("description=%s", desc))
	}

	// Handle required fields based on common patterns
	if g.isRequiredField(field.Name, objectName) {
		tags = append(tags, "required")
	}

	// Add type-specific validations
	switch {
	case strings.Contains(field.Name, "_id") || field.Name == "id":
		// Facebook IDs are numeric strings
		tags = append(tags, `pattern=^[0-9]+$`)

	case strings.Contains(field.Type, "int") || field.Type == "integer":
		// Add reasonable bounds for common integer fields
		if strings.Contains(field.Name, "age") {
			tags = append(tags, "minimum=13", "maximum=100")
		} else if strings.Contains(field.Name, "budget") || strings.Contains(field.Name, "amount") {
			tags = append(tags, "minimum=1")
		}

	case field.Type == "string" && g.isEnumType(field.Type):
		// Add enum values if available
		if enumValues := g.getEnumValues(field.Type); len(enumValues) > 0 {
			for _, v := range enumValues {
				tags = append(tags, fmt.Sprintf("enum=%s", v))
			}
		}

	case field.Name == "status":
		// Common status values
		tags = append(tags, "enum=ACTIVE", "enum=PAUSED", "enum=DELETED", "enum=ARCHIVED")

	case strings.Contains(field.Name, "url"):
		// URL fields
		tags = append(tags, "format=uri")

	case field.Type == "datetime" || field.Type == "timestamp":
		// Date/time fields
		tags = append(tags, "format=date-time")
	}

	if len(tags) > 0 {
		return fmt.Sprintf(`jsonschema:"%s"`, strings.Join(tags, ","))
	}
	return ""
}

// generateFieldDescription generates a human-readable description for a field
func (g *FieldGenerator) generateFieldDescription(fieldName, objectName string) string {
	// Convert snake_case to human readable
	words := strings.Split(fieldName, "_")
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
	case fieldName == "id":
		return fmt.Sprintf("%s ID", objectName)
	case strings.HasSuffix(fieldName, "_id"):
		return fmt.Sprintf("ID of the %s", strings.TrimSuffix(desc, " ID"))
	case fieldName == "name":
		return fmt.Sprintf("Name of the %s", objectName)
	case fieldName == "status":
		return fmt.Sprintf("Current status of the %s", objectName)
	case strings.Contains(fieldName, "created"):
		return fmt.Sprintf("When the %s was created", objectName)
	case strings.Contains(fieldName, "updated"):
		return fmt.Sprintf("When the %s was last updated", objectName)
	default:
		return desc
	}
}

// isRequiredField determines if a field should be marked as required
func (g *FieldGenerator) isRequiredField(fieldName, objectName string) bool {
	// Common required fields
	requiredFields := map[string]bool{
		"id":   true,
		"name": true,
	}

	// Object-specific required fields
	switch objectName {
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
	}

	return requiredFields[fieldName]
}

// getEnumValues returns possible enum values for a type
func (g *FieldGenerator) getEnumValues(typeName string) []string {
	// This would be populated from enum_types.json
	// For now, return empty - the actual implementation would look up the values
	return []string{}
}

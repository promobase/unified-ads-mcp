package generators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type EnumType struct {
	Name         string   `json:"name"`
	Node         string   `json:"node"`
	FieldOrParam string   `json:"field_or_param"`
	Values       []string `json:"values"`
}

type EnumGenerator struct {
	enumTypes  []EnumType
	outputPath string
}

func NewEnumGenerator(outputPath string) *EnumGenerator {
	return &EnumGenerator{
		outputPath: outputPath,
	}
}

func (g *EnumGenerator) LoadEnumTypes(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read enum types: %w", err)
	}

	if err := json.Unmarshal(data, &g.enumTypes); err != nil {
		return fmt.Errorf("failed to parse enum types: %w", err)
	}

	log.Printf("Loaded %d enum types", len(g.enumTypes))
	return nil
}

func (g *EnumGenerator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(g.outputPath, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate enums file
	if err := g.generateEnums(); err != nil {
		return fmt.Errorf("failed to generate enums: %w", err)
	}

	return nil
}

func (g *EnumGenerator) generateEnums() error {
	// Sort enums by name for consistent output
	sort.Slice(g.enumTypes, func(i, j int) bool {
		return g.enumTypes[i].Name < g.enumTypes[j].Name
	})

	// Group enums by their Go type name to handle conflicts
	enumsByTypeName := make(map[string][]EnumType)
	for _, enum := range g.enumTypes {
		typeName := ToGoTypeName(enum.Name)
		enumsByTypeName[typeName] = append(enumsByTypeName[typeName], enum)
	}

	// Sort type names for consistent output
	var typeNames []string
	for typeName := range enumsByTypeName {
		typeNames = append(typeNames, typeName)
	}
	sort.Strings(typeNames)

	// Create ordered slice for template
	type EnumGroup struct {
		TypeName string
		Enums    []EnumType
	}
	var orderedEnumGroups []EnumGroup
	
	for _, typeName := range typeNames {
		// Remove duplicate values within each enum
		enums := enumsByTypeName[typeName]
		if len(enums) > 0 {
			uniqueValues := make(map[string]bool)
			var filteredValues []string
			for _, value := range enums[0].Values {
				if !uniqueValues[value] {
					uniqueValues[value] = true
					filteredValues = append(filteredValues, value)
				}
			}
			enums[0].Values = filteredValues
			orderedEnumGroups = append(orderedEnumGroups, EnumGroup{
				TypeName: typeName,
				Enums:    enums,
			})
		}
	}

	// Prepare template data
	data := struct {
		EnumsByType []EnumGroup
		TotalEnums  int
		UniqueTypes int
	}{
		EnumsByType: orderedEnumGroups,
		TotalEnums:  len(g.enumTypes),
		UniqueTypes: len(enumsByTypeName),
	}

	// Load and execute template
	tmplContent, err := os.ReadFile(filepath.Join(filepath.Dir(g.outputPath), "codegen", "templates", "enums.go.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("enums").Funcs(template.FuncMap{
		"toGoConstName": ToGoConstName,
	}).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return os.WriteFile(filepath.Join(g.outputPath, "enums.go"), buf.Bytes(), 0644)
}

// ToGoTypeName converts enum names to Go type names
func ToGoTypeName(s string) string {
	// Handle special cases first
	if s == "" {
		return "UnknownEnum"
	}

	parts := strings.Split(s, "_")
	for i, part := range parts {
		parts[i] = strings.Title(strings.ToLower(part))
	}
	result := strings.Join(parts, "")

	// Ensure it starts with a letter
	if len(result) > 0 && (result[0] >= '0' && result[0] <= '9') {
		result = "Enum" + result
	}

	return result
}

// ToGoConstName converts enum values to Go constant names
func ToGoConstName(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "(", "_")
	s = strings.ReplaceAll(s, ")", "_")
	s = strings.ReplaceAll(s, "&", "_AND_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, ",", "_")
	s = strings.ReplaceAll(s, "'", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "+", "_PLUS_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, ";", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "!", "_")
	s = strings.ReplaceAll(s, "@", "_AT_")
	s = strings.ReplaceAll(s, "#", "_HASH_")
	s = strings.ReplaceAll(s, "$", "_DOLLAR_")
	s = strings.ReplaceAll(s, "%", "_PERCENT_")
	s = strings.ReplaceAll(s, "^", "_")
	s = strings.ReplaceAll(s, "*", "_STAR_")
	s = strings.ReplaceAll(s, "=", "_EQUALS_")
	s = strings.ReplaceAll(s, "{", "_")
	s = strings.ReplaceAll(s, "}", "_")
	s = strings.ReplaceAll(s, "[", "_")
	s = strings.ReplaceAll(s, "]", "_")
	s = strings.ReplaceAll(s, "|", "_")
	s = strings.ReplaceAll(s, "<", "_LT_")
	s = strings.ReplaceAll(s, ">", "_GT_")
	s = strings.ReplaceAll(s, "~", "_")
	s = strings.ReplaceAll(s, "`", "_")

	// Remove consecutive underscores
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}

	// Trim leading/trailing underscores
	s = strings.Trim(s, "_")

	// If empty or starts with digit, prefix with underscore
	if s == "" || (s[0] >= '0' && s[0] <= '9') {
		s = "_" + s
	}

	return s
}

package generators

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ConstantsGenerator struct {
	outputPath string
	specs      map[string]*APISpec
}

type ConstantField struct {
	FieldName    string
	ConstantName string
}

type ConstantData struct {
	PackageName string
	ObjectName  string
	Fields      []ConstantField
}

func NewConstantsGenerator(outputPath string) *ConstantsGenerator {
	return &ConstantsGenerator{
		outputPath: outputPath,
		specs:      make(map[string]*APISpec),
	}
}

// LoadAPISpec loads a single API specification file
func LoadAPISpec(path string) (*APISpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read API spec: %w", err)
	}

	var spec APISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse API spec: %w", err)
	}

	return &spec, nil
}

func (g *ConstantsGenerator) LoadAPISpecs(specsDir string) error {
	files, err := os.ReadDir(specsDir)
	if err != nil {
		return fmt.Errorf("failed to read specs directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") || file.Name() == "enum_types.json" {
			continue
		}

		specPath := filepath.Join(specsDir, file.Name())
		spec, err := LoadAPISpec(specPath)
		if err != nil {
			return fmt.Errorf("failed to load spec %s: %w", file.Name(), err)
		}

		objectName := strings.TrimSuffix(file.Name(), ".json")
		g.specs[objectName] = spec
	}

	return nil
}

func (g *ConstantsGenerator) Generate() error {
	// Create constants directory
	constantsDir := filepath.Join(g.outputPath, "constants")
	if err := os.MkdirAll(constantsDir, 0755); err != nil {
		return fmt.Errorf("failed to create constants directory: %w", err)
	}

	// Key objects to generate constants for
	keyObjects := []string{"Ad", "AdAccount", "AdCreative", "AdSet", "Campaign", "User"}

	for _, objectName := range keyObjects {
		if spec, exists := g.specs[objectName]; exists {
			if err := g.generateConstantsForObject(objectName, spec, constantsDir); err != nil {
				return fmt.Errorf("failed to generate constants for %s: %w", objectName, err)
			}
		}
	}

	return nil
}

func (g *ConstantsGenerator) generateConstantsForObject(objectName string, spec *APISpec, constantsDir string) error {
	if len(spec.Fields) == 0 {
		return nil
	}

	// Create package directory
	packageDir := filepath.Join(constantsDir, strings.ToLower(objectName))
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	// Prepare fields data
	var fields []ConstantField
	for _, field := range spec.Fields {
		constantName := toConstantName(field.Name)
		fields = append(fields, ConstantField{
			FieldName:    field.Name,
			ConstantName: constantName,
		})
	}

	data := ConstantData{
		PackageName: strings.ToLower(objectName),
		ObjectName:  objectName,
		Fields:      fields,
	}

	// Generate constants file
	outputFile := filepath.Join(packageDir, "constants.go")
	if err := g.generateFile(outputFile, data); err != nil {
		return fmt.Errorf("failed to generate constants file: %w", err)
	}

	return nil
}

func (g *ConstantsGenerator) generateFile(outputFile string, data ConstantData) error {
	// Load template
	tmpl, err := template.ParseFiles(filepath.Join("templates", "constants.go.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// toConstantName converts field name to constant name
// e.g. "account_id" -> "AccountID", "created_time" -> "CreatedTime"
func toConstantName(fieldName string) string {
	parts := strings.Split(fieldName, "_")
	var result strings.Builder
	
	for _, part := range parts {
		if part == "" {
			continue
		}
		
		// Handle common abbreviations
		switch strings.ToLower(part) {
		case "id":
			result.WriteString("ID")
		case "url":
			result.WriteString("URL")
		case "api":
			result.WriteString("API")
		case "utm":
			result.WriteString("UTM")
		case "cpm":
			result.WriteString("CPM")
		case "cpc":
			result.WriteString("CPC")
		case "ctr":
			result.WriteString("CTR")
		case "ios":
			result.WriteString("IOS")
		case "sms":
			result.WriteString("SMS")
		case "dpa":
			result.WriteString("DPA")
		case "roas":
			result.WriteString("ROAS")
		case "asc":
			result.WriteString("ASC")
		case "rf":
			result.WriteString("RF")
		case "ba":
			result.WriteString("BA")
		case "og":
			result.WriteString("OG")
		case "tco":
			result.WriteString("TCO")
		case "gdpr":
			result.WriteString("GDPR")
		case "ccpa":
			result.WriteString("CCPA")
		default:
			// Capitalize first letter
			if len(part) > 0 {
				result.WriteString(strings.ToUpper(part[:1]) + strings.ToLower(part[1:]))
			}
		}
	}
	
	return result.String()
}
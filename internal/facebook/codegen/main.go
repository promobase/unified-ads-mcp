package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"unified-ads-mcp/internal/facebook/codegen/generators"
)

type Config struct {
	specsDir   string
	outputPath string
	genType    string
	useSchema  bool // Use schema-based generation
}

func main() {
	var config Config

	// Define command line flags
	flag.StringVar(&config.specsDir, "specs", "", "Path to API specs directory (required)")
	flag.StringVar(&config.outputPath, "output", "", "Output directory (optional, defaults to ../generated relative to specs)")
	flag.StringVar(&config.genType, "type", "all", "Type of code to generate: all, enums, fields, tools")
	flag.BoolVar(&config.useSchema, "schema", false, "Use schema-based generation for enhanced type safety")

	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Facebook API Code Generator\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -specs <directory> [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate all code\n")
		fmt.Fprintf(os.Stderr, "  %s -specs ../api_specs/specs\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate with schema support (recommended)\n")
		fmt.Fprintf(os.Stderr, "  %s -specs ../api_specs/specs -schema\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate only fields with jsonschema tags\n")
		fmt.Fprintf(os.Stderr, "  %s -specs ../api_specs/specs -type fields -schema\n\n", os.Args[0])
	}

	flag.Parse()

	// Validate required arguments
	if config.specsDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Set default output path if not provided
	if config.outputPath == "" {
		config.outputPath = filepath.Join(filepath.Dir(config.specsDir), "..", "generated")
	}

	// Validate generation type
	validTypes := map[string]bool{
		"all":       true,
		"enums":     true,
		"fields":    true,
		"tools":     true,
		"constants": true,
	}
	if !validTypes[config.genType] {
		log.Fatalf("Invalid generation type: %s. Must be one of: all, enums, fields, tools, constants", config.genType)
	}

	// Run generation based on type
	switch config.genType {
	case "all":
		if err := generateAll(config); err != nil {
			log.Fatalf("Failed to generate all: %v", err)
		}
	case "enums":
		if err := generateEnums(config); err != nil {
			log.Fatalf("Failed to generate enums: %v", err)
		}
	case "fields":
		if err := generateFields(config); err != nil {
			log.Fatalf("Failed to generate fields: %v", err)
		}
	case "tools":
		if err := generateTools(config); err != nil {
			log.Fatalf("Failed to generate tools: %v", err)
		}
	case "constants":
		if err := generateConstants(config); err != nil {
			log.Fatalf("Failed to generate constants: %v", err)
		}
	}

	log.Printf("Code generation completed successfully. Output in: %s", config.outputPath)
	if config.useSchema {
		log.Printf("Schema-based generation was used for enhanced type safety")
	}
}

func generateAll(config Config) error {
	log.Println("Generating all code...")
	if config.useSchema {
		log.Println("Schema-based generation enabled for better type safety")
	}

	// Generate enums first
	if err := generateEnums(config); err != nil {
		return fmt.Errorf("failed to generate enums: %w", err)
	}

	// Generate fields
	if err := generateFields(config); err != nil {
		return fmt.Errorf("failed to generate fields: %w", err)
	}

	// Generate tools
	if err := generateTools(config); err != nil {
		return fmt.Errorf("failed to generate tools: %w", err)
	}

	// Generate constants
	if err := generateConstants(config); err != nil {
		return fmt.Errorf("failed to generate constants: %w", err)
	}

	return nil
}

func generateEnums(config Config) error {
	log.Println("Generating enum types...")

	enumPath := filepath.Join(config.specsDir, "enum_types.json")
	generator := generators.NewEnumGenerator(config.outputPath)

	if err := generator.LoadEnumTypes(enumPath); err != nil {
		return fmt.Errorf("failed to load enum types: %w", err)
	}

	if err := generator.Generate(); err != nil {
		return fmt.Errorf("failed to generate enums: %w", err)
	}

	return nil
}

func generateFields(config Config) error {
	log.Println("Generating field types...")
	if config.useSchema {
		log.Println("Using schema-based generation with jsonschema tags")
	}

	generator := generators.NewFieldGenerator(config.outputPath)

	// Enable schema generation if requested
	if config.useSchema {
		generator.EnableSchemaGeneration()
	}

	// Load enum types first
	enumPath := filepath.Join(config.specsDir, "enum_types.json")
	if err := generator.LoadEnumTypes(enumPath); err != nil {
		return fmt.Errorf("failed to load enum types: %w", err)
	}

	// Load API specs
	if err := generator.LoadAPISpecs(config.specsDir); err != nil {
		return fmt.Errorf("failed to load API specs: %w", err)
	}

	if err := generator.Generate(); err != nil {
		return fmt.Errorf("failed to generate fields: %w", err)
	}

	return nil
}

func generateTools(config Config) error {
	log.Println("Generating MCP tools...")
	if config.useSchema {
		log.Println("Using schema-based tool generation for proper type handling")
	}

	generator := generators.NewToolGenerator(config.outputPath)

	// Enable schema generation if requested
	if config.useSchema {
		generator.EnableSchemaGeneration()
	}

	// Load enum types first
	enumPath := filepath.Join(config.specsDir, "enum_types.json")
	if err := generator.LoadEnumTypes(enumPath); err != nil {
		return fmt.Errorf("failed to load enum types: %w", err)
	}

	// Load API specs
	if err := generator.LoadAPISpecs(config.specsDir); err != nil {
		return fmt.Errorf("failed to load API specs: %w", err)
	}

	if err := generator.Generate(); err != nil {
		return fmt.Errorf("failed to generate tools: %w", err)
	}

	return nil
}

func generateConstants(config Config) error {
	log.Println("Generating field constants...")

	generator := generators.NewConstantsGenerator(config.outputPath)

	// Load API specs
	if err := generator.LoadAPISpecs(config.specsDir); err != nil {
		return fmt.Errorf("failed to load API specs: %w", err)
	}

	if err := generator.Generate(); err != nil {
		return fmt.Errorf("failed to generate constants: %w", err)
	}

	return nil
}

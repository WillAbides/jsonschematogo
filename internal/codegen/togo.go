package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// GoStructGenerator generates Go structs from JSON Schema with extensions.
type GoStructGenerator struct {
	schemaCompiler *SchemaCompiler
	typeInferrer   *TypeInferrer
	codeRenderer   *CodeRenderer
	generatedTypes map[string]*GeneratedStruct // Cache to avoid duplicate types
}

// NewGoStructGenerator creates a new Go struct generator.
func NewGoStructGenerator() *GoStructGenerator {
	return &GoStructGenerator{
		schemaCompiler: NewSchemaCompiler(),
		typeInferrer:   NewTypeInferrer(),
		codeRenderer:   NewCodeRenderer(),
		generatedTypes: make(map[string]*GeneratedStruct),
	}
}

// AddSchema adds a schema that can be referenced by other schemas.
func (g *GoStructGenerator) AddSchema(uri string, rawSchema any) error {
	return g.schemaCompiler.AddSchema(uri, rawSchema)
}

// GenerateStructs generates Go structs for a schema and all its referenced schemas.
func (g *GoStructGenerator) GenerateStructs(
	uri string,
	rawSchema any,
	structName string,
) ([]*GeneratedStruct, error) {
	// Add the main schema only if it doesn't already exist
	if _, exists := g.schemaCompiler.GetRawSchema(uri); !exists {
		if err := g.AddSchema(uri, rawSchema); err != nil {
			return nil, fmt.Errorf("failed to add main schema: %w", err)
		}
	}

	// Generate the main struct and all referenced structs
	_, err := g.generateStructRecursive(uri, structName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate structs: %w", err)
	}

	// Convert map to slice with consistent ordering
	var uris []string
	for u := range g.generatedTypes {
		uris = append(uris, u)
	}
	sort.Strings(uris)

	var results []*GeneratedStruct
	for _, u := range uris {
		results = append(results, g.generatedTypes[u])
	}

	return results, nil
}

// generateStructRecursive recursively generates structs for a schema and its references.
func (g *GoStructGenerator) generateStructRecursive(uri, structName string) (*GeneratedStruct, error) {
	// Check if we've already generated this type
	if existing, exists := g.generatedTypes[uri]; exists {
		return existing, nil
	}

	rawSchema, exists := g.schemaCompiler.GetRawSchema(uri)
	if !exists {
		return nil, fmt.Errorf("schema not found for URI: %s", uri)
	}

	schema, err := g.schemaCompiler.CompileSchemaWithExtensions(uri, rawSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	// Use x-go-type if available, otherwise use the provided struct name
	typeName := schema.GetGoType()
	if typeName == "any" || typeName == "map[string]any" {
		typeName = structName
	}

	var fields []StructField

	// Process each property
	for propName, propSchema := range schema.Schema.Properties {
		field, err := g.typeInferrer.ProcessProperty(propName, propSchema, schema, g)
		if err != nil {
			return nil, fmt.Errorf("failed to process property %s: %w", propName, err)
		}
		fields = append(fields, *field)
	}

	// Generate the struct code
	code, err := g.codeRenderer.RenderStruct(typeName, fields)
	if err != nil {
		return nil, fmt.Errorf("failed to render struct: %w", err)
	}

	generated := &GeneratedStruct{
		Name:   typeName,
		Code:   code,
		Fields: fields,
	}

	// Cache the generated struct
	g.generatedTypes[uri] = generated

	return generated, nil
}

// findSchemaURI finds the original URI that was used to add a schema.
func (g *GoStructGenerator) findSchemaURI(cleanRefURI string) string {
	for uri := range g.schemaCompiler.GetAllRawSchemas() {
		if strings.HasSuffix(cleanRefURI, strings.TrimPrefix(uri, "./")) ||
			uri == cleanRefURI ||
			strings.HasSuffix(uri, strings.TrimPrefix(cleanRefURI, "file://")) {
			return uri
		}
	}
	return ""
}

// extractTypeNameFromURI extracts a type name from a URI.
func (g *GoStructGenerator) extractTypeNameFromURI(uri string) string {
	if uri == "" {
		return "Object"
	}

	parts := strings.Split(uri, "/")
	if len(parts) == 0 {
		return "Object"
	}

	name := parts[len(parts)-1]
	name = g.cleanTypeName(name)
	if name != "" {
		return capitalizeFirst(name)
	}

	return "Object"
}

// cleanTypeName removes file extensions and fragments from a type name.
func (g *GoStructGenerator) cleanTypeName(name string) string {
	// Remove file extensions
	name = strings.TrimSuffix(name, ".json")
	name = strings.TrimSuffix(name, ".yaml")

	// Remove fragment
	if strings.Contains(name, "#") {
		name = strings.Split(name, "#")[0]
	}

	return name
}

// extractInlineObjectSchema extracts the raw schema for an inline object property.
func (g *GoStructGenerator) extractInlineObjectSchema(propSchema *jsonschema.Schema) (any, error) {
	// For now, create a simple object schema based on the compiled schema
	// In a more complete implementation, this would extract from the original raw schema
	schema := map[string]any{
		"type":       "object",
		"properties": make(map[string]any),
	}

	// Add properties from the compiled schema
	if propSchema.Properties != nil {
		for propName, prop := range propSchema.Properties {
			propType := inferGoTypeFromCompiledSchema(prop)
			schema["properties"].(map[string]any)[propName] = map[string]any{
				"type": propType,
			}
		}
	}

	return schema, nil
}

// GenerateAllCode generates all struct code as a single string.
func (g *GoStructGenerator) GenerateAllCode(structs []*GeneratedStruct) string {
	return g.codeRenderer.GenerateAllCode(structs)
}

package codegen

import (
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// SchemaCompiler handles compilation of JSON schemas with extensions.
type SchemaCompiler struct {
	compiler   *jsonschema.Compiler
	rawSchemas map[string]any
	extHandler *ExtensionHandler
}

// NewSchemaCompiler creates a new schema compiler.
func NewSchemaCompiler() *SchemaCompiler {
	return &SchemaCompiler{
		compiler:   jsonschema.NewCompiler(),
		rawSchemas: make(map[string]any),
		extHandler: NewExtensionHandler(),
	}
}

// AddSchema adds a schema that can be referenced by other schemas.
func (c *SchemaCompiler) AddSchema(uri string, rawSchema any) error {
	c.rawSchemas[uri] = rawSchema
	return c.compiler.AddResource(uri, rawSchema)
}

// CompileSchemaWithExtensions compiles a schema while preserving custom extensions.
func (c *SchemaCompiler) CompileSchemaWithExtensions(uri string, rawSchema any) (*SchemaWithExtensions, error) {
	compiledSchema, err := c.compiler.Compile(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	extensions, properties, err := c.extHandler.ExtractExtensions(rawSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to extract extensions: %w", err)
	}

	return &SchemaWithExtensions{
		Schema:     compiledSchema,
		Extensions: extensions,
		Properties: properties,
	}, nil
}

// GetRawSchema retrieves a raw schema by URI.
func (c *SchemaCompiler) GetRawSchema(uri string) (any, bool) {
	schema, exists := c.rawSchemas[uri]
	return schema, exists
}

// GetAllRawSchemas returns all registered raw schemas.
func (c *SchemaCompiler) GetAllRawSchemas() map[string]any {
	return c.rawSchemas
}

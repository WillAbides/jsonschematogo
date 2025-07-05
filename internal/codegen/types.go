package codegen

import (
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// SchemaWithExtensions combines a compiled schema with custom extensions.
type SchemaWithExtensions struct {
	*jsonschema.Schema
	Extensions map[string]any
	Properties map[string]*PropertyWithExtensions
}

// PropertyWithExtensions represents a schema property with custom extensions.
type PropertyWithExtensions struct {
	*jsonschema.Schema
	Extensions map[string]any
}

// GeneratedStruct represents a generated Go struct.
type GeneratedStruct struct {
	Name   string
	Code   string
	Fields []StructField
}

// StructField represents a field in a Go struct.
type StructField struct {
	Name    string
	Type    string
	JSONTag string
	IsRef   bool
	RefType string
}

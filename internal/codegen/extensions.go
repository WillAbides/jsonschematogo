package codegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ExtensionHandler handles custom JSON Schema extensions.
type ExtensionHandler struct{}

// NewExtensionHandler creates a new extension handler.
func NewExtensionHandler() *ExtensionHandler {
	return &ExtensionHandler{}
}

// ExtractExtensions recursively extracts custom extensions from a schema.
func (h *ExtensionHandler) ExtractExtensions(schema any) (map[string]any, map[string]*PropertyWithExtensions, error) {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil, nil, nil
	}

	extensions := make(map[string]any)
	for key, value := range schemaMap {
		if h.isExtension(key) {
			extensions[key] = value
		}
	}

	properties := h.extractPropertyExtensions(schemaMap)
	return extensions, properties, nil
}

// extractPropertyExtensions extracts property-level extensions from a schema map.
func (h *ExtensionHandler) extractPropertyExtensions(schemaMap map[string]any) map[string]*PropertyWithExtensions {
	properties := make(map[string]*PropertyWithExtensions)
	propsInterface, exists := schemaMap["properties"]
	if !exists {
		return properties
	}
	propsMap, ok := propsInterface.(map[string]any)
	if !ok {
		return properties
	}
	for propName, propSchema := range propsMap {
		propMap, ok := propSchema.(map[string]any)
		if !ok {
			continue
		}
		propExtensions := make(map[string]any)
		for key, value := range propMap {
			if h.isExtension(key) {
				propExtensions[key] = value
			}
		}
		properties[propName] = &PropertyWithExtensions{
			Schema:     nil, // Parent schema's Properties map is used
			Extensions: propExtensions,
		}
	}
	return properties
}

// ExtractExtensionsFromRawJSON extracts extensions from the raw JSON data.
func (h *ExtensionHandler) ExtractExtensionsFromRawJSON(rawMap map[string]any) (map[string]any, map[string]*PropertyWithExtensions, error) {
	extensions := make(map[string]any)
	for key, value := range rawMap {
		if h.isExtension(key) {
			extensions[key] = value
		}
	}

	properties := h.extractPropertyExtensions(rawMap)
	return extensions, properties, nil
}

// IsExtension checks if a key is a custom extension (starts with "x-").
func (h *ExtensionHandler) isExtension(key string) bool {
	return strings.HasPrefix(key, "x-")
}

// GetGoType returns the Go type for a schema, checking for x-go-type extension.
func (s *SchemaWithExtensions) GetGoType() string {
	if goType, exists := s.Extensions["x-go-type"]; exists {
		if goTypeStr, ok := goType.(string); ok {
			return goTypeStr
		}
	}
	return s.inferGoTypeFromSchema()
}

// GetPropertyGoType returns the Go type for a property, checking for x-go-type extension.
func (s *SchemaWithExtensions) GetPropertyGoType(propName string) string {
	if prop, exists := s.Properties[propName]; exists {
		if goType, exists := prop.Extensions["x-go-type"]; exists {
			if goTypeStr, ok := goType.(string); ok {
				return goTypeStr
			}
		}
	}
	if prop, exists := s.Schema.Properties[propName]; exists {
		return inferGoTypeFromCompiledSchema(prop)
	}
	return "interface{}"
}

// inferGoTypeFromSchema infers Go type from the schema's standard properties.
func (s *SchemaWithExtensions) inferGoTypeFromSchema() string {
	if s.Schema.Types == nil || s.Schema.Types.IsEmpty() {
		return "any"
	}

	types := s.Schema.Types.ToStrings()
	if len(types) == 0 {
		return "any"
	}

	switch types[0] {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		return "[]any"
	case "object":
		return "map[string]any"
	default:
		return "any"
	}
}

// inferGoTypeFromCompiledSchema infers Go type from a compiled schema property.
func inferGoTypeFromCompiledSchema(schema *jsonschema.Schema) string {
	if schema.Types == nil || schema.Types.IsEmpty() {
		return "any"
	}

	types := schema.Types.ToStrings()
	if len(types) == 0 {
		return "any"
	}

	switch types[0] {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		return "[]any"
	case "object":
		return "map[string]any"
	default:
		return "any"
	}
}

// UnmarshalJSONSchema unmarshals JSON Schema with extension support.
func (s *SchemaWithExtensions) UnmarshalJSONSchema(data []byte) error {
	// Unmarshal the standard JSON Schema properties
	if err := json.Unmarshal(data, &s.Schema); err != nil {
		return fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	// Extract extensions from the raw JSON
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return fmt.Errorf("failed to unmarshal raw schema: %w", err)
	}

	// Extract custom extensions
	handler := NewExtensionHandler()
	extensions, properties, err := handler.ExtractExtensionsFromRawJSON(rawMap)
	if err != nil {
		return fmt.Errorf("failed to extract extensions: %w", err)
	}

	s.Extensions = extensions
	s.Properties = properties

	return nil
}

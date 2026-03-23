package schema

import (
	"cmp"
	"encoding/json"
	"iter"
	"maps"
	"slices"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Schema represents a JSON Schema in a language-agnostic way.
type Schema struct {
	schema *jsonschema.Schema
	rawMap map[string]any
}

// Type returns the schema type.
func (s *Schema) Type() string {
	if s.schema.Types != nil && !s.schema.Types.IsEmpty() {
		types := s.schema.Types.ToStrings()
		if len(types) > 0 {
			return types[0]
		}
	}
	return ""
}

// Ref returns the schema reference.
func (s *Schema) Ref() string {
	// Try raw map first to preserve original references
	if s.rawMap != nil {
		if ref, ok := s.rawMap["$ref"].(string); ok {
			return ref
		}
	}
	// Fall back to compiled schema
	if s.schema.Ref != nil {
		return s.schema.Ref.Location
	}
	return ""
}

func (s *Schema) RefSchema() *Schema {
	if s.schema.Ref != nil {
		schema := Schema{
			schema: s.schema.Ref,
		}
		getMapValue(s.rawMap, "$ref", &schema.rawMap)
		return &schema
	}
	return nil
}

// Required returns the list of required property names.
func (s *Schema) Required() []string {
	return s.schema.Required
}

// Items returns the schema for array items.
func (s *Schema) Items() *Schema {
	if s.schema.Items2020 != nil {
		var itemsRawMap map[string]any
		if s.rawMap != nil {
			if m, ok := s.rawMap["items"].(map[string]any); ok {
				itemsRawMap = m
			}
		}
		return &Schema{
			schema: s.schema.Items2020,
			rawMap: itemsRawMap,
		}
	}

	if s.schema.Items == nil {
		return nil
	}

	switch items := s.schema.Items.(type) {
	case *jsonschema.Schema:
		var itemsRawMap map[string]any
		if s.rawMap != nil {
			if m, ok := s.rawMap["items"].(map[string]any); ok {
				itemsRawMap = m
			}
		}
		return &Schema{
			schema: items,
			rawMap: itemsRawMap,
		}
	case []*jsonschema.Schema:
		if len(items) > 0 {
			var itemsRawMap map[string]any
			if s.rawMap != nil {
				if m, ok := s.rawMap["items"].(map[string]any); ok {
					itemsRawMap = m
				}
			}
			return &Schema{
				schema: items[0],
				rawMap: itemsRawMap,
			}
		}
	}
	return nil
}

func orderedMap[K cmp.Ordered, V any](m map[K]V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		keys := slices.Collect(maps.Keys(m))
		slices.Sort(keys)
		for _, key := range keys {
			if !yield(key, m[key]) {
				return
			}
		}
	}
}

func (s *Schema) OrderedProperties() iter.Seq2[string, *Schema] {
	return orderedMap(s.Properties())
}

// Properties returns the schema properties.
func (s *Schema) Properties() map[string]*Schema {
	if s.schema.Properties == nil {
		return nil
	}

	props := make(map[string]*Schema)
	var rawProps map[string]any
	if s.rawMap != nil {
		if m, ok := s.rawMap["properties"].(map[string]any); ok {
			rawProps = m
		}
	}

	for propName, propSchema := range s.schema.Properties {
		var propRawMap map[string]any
		if rawProps != nil {
			if m, ok := rawProps[propName].(map[string]any); ok {
				propRawMap = m
			}
		}
		props[propName] = &Schema{
			schema: propSchema,
			rawMap: propRawMap,
		}
	}
	return props
}

// Location returns the schema location.
func (s *Schema) Location() string {
	return s.schema.Location
}

// IsObject returns true if the schema represents an object type.
func (s *Schema) IsObject() bool {
	return s.Type() == "object"
}

type GoTypeImport struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type Extensions struct {
	GoType       *string       `json:"x-go-type"`
	GoTypeImport *GoTypeImport `json:"x-go-type-import"`
	GoName       *string       `json:"x-go-name"`
	GoTypeName   *string       `json:"x-go-type-name"`
}

func (s *Schema) Extensions() (*Extensions, error) {
	b, err := json.Marshal(s.rawMap)
	if err != nil {
		return nil, err
	}
	var ext Extensions
	err = json.Unmarshal(b, &ext)
	if err != nil {
		return nil, err
	}
	return &ext, nil
}

// GetImportExtension retrieves import information from x-go-type-import extension.
// Returns path and name if the extension exists and is properly formatted.
func (s *Schema) GetImportExtension() (path, name string, exists bool) {
	var importMap map[string]any
	if !getMapValue(s.rawMap, "x-go-type-import", &importMap) {
		return "", "", false
	}
	if !getMapValue(importMap, "path", &path) {
		return "", "", false
	}
	if !getMapValue(importMap, "name", &name) {
		return "", "", false
	}

	return path, name, true
}

// HasProperties returns true if the schema has properties (is an object with properties).
func (s *Schema) HasProperties() bool {
	return s.IsObject() && len(s.Properties()) > 0
}

// IsPropertyRequired returns true if a property is required.
func (s *Schema) IsPropertyRequired(propName string) bool {
	return slices.Contains(s.Required(), propName)
}

func getMapValue[K comparable, V any, T any](mp map[K]V, key K, target *T) bool {
	if mp == nil {
		return false
	}
	mapValue, exists := mp[key]
	if !exists {
		return false
	}
	targetValue, ok := any(mapValue).(T)
	if !ok {
		return false
	}
	*target = targetValue
	return true
}

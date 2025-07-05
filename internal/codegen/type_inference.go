package codegen

import (
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// TypeInferrer handles type inference and property processing.
type TypeInferrer struct{}

// NewTypeInferrer creates a new type inferrer.
func NewTypeInferrer() *TypeInferrer {
	return &TypeInferrer{}
}

// ProcessProperty processes a single property and handles references.
func (i *TypeInferrer) ProcessProperty(
	propName string,
	propSchema *jsonschema.Schema,
	parentSchema *SchemaWithExtensions,
	generator *GoStructGenerator,
) (*StructField, error) {
	goType := parentSchema.GetPropertyGoType(propName)
	jsonTag := fmt.Sprintf("`json:\"%s\"`", propName)

	field := &StructField{
		Name:    capitalizeFirst(propName),
		Type:    goType,
		JSONTag: jsonTag,
		IsRef:   false,
	}

	// Check if this property is a reference
	if propSchema.Ref != nil {
		return i.processReferenceProperty(propSchema, field, generator)
	}

	// Check if this is an inline object that should be a separate struct
	if i.isObjectType(propSchema) && i.shouldGenerateInlineStruct(propSchema) {
		return i.processInlineObjectProperty(propName, propSchema, field, parentSchema, generator)
	}

	return field, nil
}

// processReferenceProperty handles properties that reference other schemas.
func (i *TypeInferrer) processReferenceProperty(
	propSchema *jsonschema.Schema,
	field *StructField,
	generator *GoStructGenerator,
) (*StructField, error) {
	refURI := propSchema.Ref.Location
	cleanRefURI := strings.TrimSuffix(refURI, "#")

	// Find the original URI that was used to add the schema
	foundURI := generator.findSchemaURI(cleanRefURI)
	if foundURI == "" {
		foundURI = cleanRefURI
	}

	refTypeName := generator.extractTypeNameFromURI(foundURI)

	// Generate the referenced struct
	_, err := generator.generateStructRecursive(foundURI, refTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate referenced struct for %s (resolved to %s): %w", refURI, foundURI, err)
	}

	field.Type = refTypeName
	field.IsRef = true
	field.RefType = foundURI

	return field, nil
}

// processInlineObjectProperty handles inline object properties that should become separate structs.
func (i *TypeInferrer) processInlineObjectProperty(
	propName string,
	propSchema *jsonschema.Schema,
	field *StructField,
	parentSchema *SchemaWithExtensions,
	generator *GoStructGenerator,
) (*StructField, error) {
	inlineTypeName := capitalizeFirst(propName) + "Object"
	inlineURI := fmt.Sprintf("%s#/properties/%s", parentSchema.Location, propName)

	// Extract the raw schema for the inline object
	rawProp, err := generator.extractInlineObjectSchema(propSchema)
	if err != nil {
		return field, nil // Fall back to the original field
	}

	generator.schemaCompiler.AddSchema(inlineURI, rawProp)

	_, err = generator.generateStructRecursive(inlineURI, inlineTypeName)
	if err != nil {
		return field, nil // Fall back to the original field
	}

	field.Type = inlineTypeName
	field.IsRef = true
	field.RefType = inlineURI

	return field, nil
}

// isObjectType checks if a schema represents an object type.
func (i *TypeInferrer) isObjectType(schema *jsonschema.Schema) bool {
	if schema.Types == nil || schema.Types.IsEmpty() {
		return false
	}

	types := schema.Types.ToStrings()
	for _, t := range types {
		if t == "object" {
			return true
		}
	}
	return false
}

// shouldGenerateInlineStruct determines if an inline object should get its own struct.
func (i *TypeInferrer) shouldGenerateInlineStruct(schema *jsonschema.Schema) bool {
	// Generate inline struct if it has properties
	return len(schema.Properties) > 0
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

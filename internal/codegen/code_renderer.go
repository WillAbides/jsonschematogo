package codegen

import (
	"fmt"
	"sort"
	"strings"
)

// CodeRenderer handles rendering of Go code from generated structs.
type CodeRenderer struct{}

// NewCodeRenderer creates a new code renderer.
func NewCodeRenderer() *CodeRenderer {
	return &CodeRenderer{}
}

// RenderStruct renders a Go struct using a template.
func (r *CodeRenderer) RenderStruct(typeName string, fields []StructField) (string, error) {
	var builder strings.Builder

	// Sort fields by name for consistent ordering
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	builder.WriteString(fmt.Sprintf("type %s struct {\n", typeName))
	for _, field := range fields {
		builder.WriteString(fmt.Sprintf("\t%s %s %s\n", field.Name, field.Type, field.JSONTag))
	}
	builder.WriteString("}")

	return builder.String(), nil
}

// GenerateAllCode generates all struct code as a single string.
func (r *CodeRenderer) GenerateAllCode(structs []*GeneratedStruct) string {
	var builder strings.Builder

	for i, s := range structs {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(s.Code)
	}

	return builder.String()
}

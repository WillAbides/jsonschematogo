package codegen

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/willabides/jsonschematogo/internal/golden"
	"gopkg.in/yaml.v3"
)

func TestObjectReferences(t *testing.T) {
	// Define a Person schema that will be referenced
	var personSchema any
	personSchemaPath := "testdata/schemas/company/person.yaml"
	personSchemaBytes, err := os.ReadFile(personSchemaPath)
	require.NoError(t, err)
	err = yaml.Unmarshal(personSchemaBytes, &personSchema)
	require.NoError(t, err)

	// Define a Company schema that references Person
	var companySchema any
	companySchemaPath := "testdata/schemas/company/company.yaml"
	companySchemaBytes, err := os.ReadFile(companySchemaPath)
	require.NoError(t, err)
	err = yaml.Unmarshal(companySchemaBytes, &companySchema)
	require.NoError(t, err)

	// Test the new generator with object references
	generator := NewGoStructGenerator()

	// Add the Person schema first (it will be referenced)
	err = generator.AddSchema("person.yaml", personSchema)
	require.NoError(t, err)

	// Generate structs for the Company schema (which references Person)
	structs, err := generator.GenerateStructs("company.yaml", companySchema, "Company")
	require.NoError(t, err)

	// Generate all code
	allCode := generator.GenerateAllCode(structs)
	fmt.Printf("\nGenerated Go code:\n%s\n", allCode)

	golden.AssertGolden(t, []byte(allCode), nil)
}

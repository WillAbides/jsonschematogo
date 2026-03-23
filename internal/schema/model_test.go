package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchema_IsPrimitive(t *testing.T) {
	schema, err := LoadSchema("../codegen/testdata/schemas/primitives.yaml")
	require.NoError(t, err)
	assert.NotEmpty(t, schema.Type())
}

func TestSchema_IsObject(t *testing.T) {
	schema, err := LoadSchema("../codegen/testdata/schemas/company/company.yaml")
	require.NoError(t, err)
	assert.True(t, schema.IsObject())
}

func TestSchema_HasProperties(t *testing.T) {
	schema, err := LoadSchema("../codegen/testdata/schemas/company/company.yaml")
	require.NoError(t, err)
	assert.True(t, schema.HasProperties())
}

func TestSchema_IsPropertyRequired(t *testing.T) {
	schema, err := LoadSchema("../codegen/testdata/schemas/company/company.yaml")
	require.NoError(t, err)
	assert.True(t, schema.IsPropertyRequired("name"))
	assert.False(t, schema.IsPropertyRequired("email"))
}

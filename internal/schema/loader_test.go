package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSchema(t *testing.T) {
	schema, err := LoadSchema("../codegen/testdata/schemas/company/company.yaml")
	require.NoError(t, err)
	assert.Equal(t, "object", schema.Type())

	props := schema.Properties()
	assert.NotNil(t, props)
	assert.Len(t, props, 4)

	// Check required properties
	assert.True(t, schema.IsPropertyRequired("ceo"))
	assert.True(t, schema.IsPropertyRequired("employees"))
	assert.True(t, schema.IsPropertyRequired("name"))
	assert.True(t, schema.IsPropertyRequired("founded"))

	// Check specific properties
	ceo, ok := props["ceo"]
	switch {
	case !ok:
		assert.Fail(t, "ceo property missing")
	case ceo.Ref() != "person.yaml":
		assert.Equal(t, "person.yaml", ceo.Ref())
	}

	employees, ok := props["employees"]
	assert.True(t, ok, "employees property missing")
	assert.Equal(t, "array", employees.Type())
	items := employees.Items()
	assert.NotNil(t, items)
	assert.Equal(t, "person.yaml", items.Ref())
}

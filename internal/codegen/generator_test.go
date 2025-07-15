package codegen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/willabides/jsonschematogo/internal/codegen/run/testrun"
	"github.com/willabides/jsonschematogo/internal/testutil"
	"gopkg.in/yaml.v3"
)

func normalizeRunResult(result testrun.Result) testrun.Result {
	stderr := strings.ReplaceAll(result.Stderr, filepath.ToSlash(testutil.RepoRoot())+"/", "")
	return testrun.Result{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   stderr,
	}
}

func testCodegen(t *testing.T, files ...string) {
	goOutputFile := t.TempDir() + "/output.go"
	args := []string{"--output", goOutputFile}
	args = append(args, files...)
	runResult := testrun.Run(args...)
	runResultYaml, err := yaml.Marshal(normalizeRunResult(runResult))
	require.NoError(t, err)
	testutil.AssertGolden(t, runResultYaml, &testutil.AssertGoldenOptions{
		GoldenfileSuffix: "/run_result.yaml",
	})
	goOutput, err := os.ReadFile(goOutputFile)
	require.NoError(t, err)
	testutil.AssertGolden(t, goOutput, &testutil.AssertGoldenOptions{
		GoldenfileSuffix: "/output.go",
	})
}

func TestCodegen(t *testing.T) {
	for _, test := range []struct {
		name string
		file string
	}{
		{
			name: "Primitives",
			file: "testdata/schemas/primitives.yaml",
		},
		{
			name: "ArrayOfPrimitives",
			file: "testdata/schemas/array_of_primitives.yaml",
		},
		{
			name: "MapType",
			file: "testdata/schemas/map_type.yaml",
		},
		{
			name: "InlineObject",
			file: "testdata/schemas/inline_object.yaml",
		},
		{
			name: "EnumType",
			file: "testdata/schemas/enum_type.yaml",
		},
		{
			name: "XGoTypeNested",
			file: "testdata/schemas/x_go_type_nested.yaml",
		},
		{
			name: "XGoTypeArrays",
			file: "testdata/schemas/x_go_type_arrays.yaml",
		},
		{
			name: "XGoTypePrimitives",
			file: "testdata/schemas/x_go_type_primitives.yaml",
		},
		{
			name: "XGoTypeImport",
			file: "testdata/schemas/x_go_type_import.yaml",
		},
		{
			name: "Company",
			file: "testdata/schemas/company/company.yaml",
		},
		{
			name: "OptionalProperties",
			file: "testdata/schemas/optional_properties.yaml",
		},
		{
			name: "EmptyArray",
			file: "testdata/schemas/empty_array.yaml",
		},
		{
			name: "NestedEmptyObject",
			file: "testdata/schemas/nested_empty_object.yaml",
		},
		{
			name: "ComplexNesting",
			file: "testdata/schemas/complex_nesting.yaml",
		},
		{
			name: "FieldNamingEdgeCases",
			file: "testdata/schemas/field_naming_edge_cases.yaml",
		},
		{
			name: "MultipleEnums",
			file: "testdata/schemas/multiple_enums.yaml",
		},
		{
			name: "MixedTypeArray",
			file: "testdata/schemas/mixed_type_array.yaml",
		},
		{
			name: "EmptyObjectSchema",
			file: "testdata/schemas/empty_object_schema.yaml",
		},
		{
			name: "ArrayWithRefItems",
			file: "testdata/schemas/array_with_ref_items.yaml",
		},
		{
			name: "ObjectWithNoProperties",
			file: "testdata/schemas/object_with_no_properties.yaml",
		},
		{
			name: "ComplexXGoTypeImport",
			file: "testdata/schemas/complex_x_go_type_import.yaml",
		},
		{
			name: "ExtensionEdgeCases",
			file: "testdata/schemas/extension_edge_cases.yaml",
		},
		{
			name: "SchemaDraft2019",
			file: "testdata/schemas/schema_draft_2019.yaml",
		},
		{
			name: "SchemaDraft2020",
			file: "testdata/schemas/schema_draft_2020.yaml",
		},
		{
			name: "SchemaDraft07",
			file: "testdata/schemas/schema_draft_07.yaml",
		},
		{
			name: "NoSchemaDraft",
			file: "testdata/schemas/no_schema_draft.yaml",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			testCodegen(t, test.file)
		})
	}
}

func TestCodegenErrors(t *testing.T) {
	for _, test := range []struct {
		name        string
		file        string
		expectError bool
	}{
		{
			name:        "NonExistentFile",
			file:        "testdata/schemas/nonexistent.yaml",
			expectError: true,
		},
		{
			name:        "InvalidSchema",
			file:        "testdata/schemas/invalid_schema.yaml",
			expectError: true,
		},
		{
			name:        "SchemaDraft2019",
			file:        "testdata/schemas/schema_draft_2019.yaml",
			expectError: false,
		},
		{
			name:        "SchemaDraft2020",
			file:        "testdata/schemas/schema_draft_2020.yaml",
			expectError: false,
		},
		{
			name:        "SchemaDraft07",
			file:        "testdata/schemas/schema_draft_07.yaml",
			expectError: false,
		},
		{
			name:        "NoSchemaDraft",
			file:        "testdata/schemas/no_schema_draft.yaml",
			expectError: false,
		},
		{
			name:        "MalformedYAML",
			file:        "testdata/schemas/malformed_yaml.yaml",
			expectError: true,
		},
		{
			name:        "EmptyFile",
			file:        "testdata/schemas/empty_file.yaml",
			expectError: true,
		},
		{
			name:        "InvalidRef",
			file:        "testdata/schemas/invalid_ref.yaml",
			expectError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			testCodegenError(t, test.file, test.expectError)
		})
	}
}

func testCodegenError(t *testing.T, file string, expectError bool) {
	goOutputFile := t.TempDir() + "/output.go"
	args := []string{"--output", goOutputFile}
	args = append(args, file)
	runResult := testrun.Run(args...)

	runResultYaml, err := yaml.Marshal(normalizeRunResult(runResult))
	require.NoError(t, err)
	testutil.AssertGolden(t, runResultYaml, &testutil.AssertGoldenOptions{
		GoldenfileSuffix: "/run_result.yaml",
	})

	if expectError {
		require.NotEqual(t, 0, runResult.ExitCode, "Expected error but got success")
		require.NotEmpty(t, runResult.Stderr, "Expected error message in stderr")
	} else {
		require.Equal(t, 0, runResult.ExitCode, "Expected success but got error")
		require.Empty(t, runResult.Stderr, "Unexpected error message in stderr")
	}
}

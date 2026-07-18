package schema

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/willabides/jsonschematogo/internal/schemaloader"
)

// LoadSchema loads a JSON or YAML schema file and parses it into a *Schema model.
func LoadSchema(filename string) (*Schema, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}
	u, err := url.Parse(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	absPath, err := filepath.Abs(u.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	fileURL := "file://" + absPath

	schemaMap := map[string]any{}
	loader, err := schemaloader.New(
		func(url string, schema any) {
			schemaMap[url] = schema
		},
		&schemaloader.Options{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema loader: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(loader)

	compiled, err := compiler.Compile(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	rawSchema, ok := schemaMap[fileURL]
	if !ok {
		return nil, fmt.Errorf("schema not found in map after compilation for URL: %s", fileURL)
	}
	rawMap, ok := rawSchema.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected schema type: %T", rawSchema)
	}

	definitions, err := compileDefinitions(compiler, fileURL, rawMap)
	if err != nil {
		return nil, err
	}

	schema := fromJSONSchema(compiled, rawMap)
	schema.definitions = definitions
	return schema, nil
}

func fromJSONSchema(compiled *jsonschema.Schema, rawMap map[string]any) *Schema {
	return &Schema{
		schema: compiled,
		rawMap: rawMap,
	}
}

func compileDefinitions(
	compiler *jsonschema.Compiler,
	fileURL string,
	rawMap map[string]any,
) ([]NamedSchema, error) {
	definitions := []NamedSchema{}
	for _, keyword := range []string{"$defs", "definitions"} {
		keywordDefinitions, err := compileKeywordDefinitions(compiler, fileURL, rawMap, keyword)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, keywordDefinitions...)
	}
	return definitions, nil
}

func compileKeywordDefinitions(
	compiler *jsonschema.Compiler,
	fileURL string,
	rawMap map[string]any,
	keyword string,
) ([]NamedSchema, error) {
	rawDefinitions, definitionsExist := rawMap[keyword].(map[string]any)
	if !definitionsExist {
		return []NamedSchema{}, nil
	}

	definitions := make([]NamedSchema, 0, len(rawDefinitions))
	for name, rawDefinition := range rawDefinitions {
		var rawDefinitionMap map[string]any
		switch definition := rawDefinition.(type) {
		case map[string]any:
			rawDefinitionMap = definition
		case bool:
		default:
			continue
		}

		ref := "#/" + keyword + "/" + escapeJSONPointerToken(name)
		location := fileURL + ref
		compiled, err := compiler.Compile(location)
		if err != nil {
			return nil, fmt.Errorf("compile definition %q: %w", name, err)
		}
		definitions = append(definitions, NamedSchema{
			Name:   name,
			Ref:    ref,
			Schema: fromJSONSchema(compiled, rawDefinitionMap),
		})
	}
	return definitions, nil
}

func escapeJSONPointerToken(token string) string {
	token = strings.ReplaceAll(token, "~", "~0")
	return strings.ReplaceAll(token, "/", "~1")
}

// LoadAllSchemas loads all entry schemas and recursively loads all referenced schemas.
func LoadAllSchemas(entryFiles []string) (map[string]*Schema, error) {
	if len(entryFiles) == 0 {
		return nil, fmt.Errorf("no schema files provided")
	}

	schemas := make(map[string]*Schema)
	// Load all entry point schemas
	for _, file := range entryFiles {
		sch, err := LoadSchema(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load schema %s: %w", file, err)
		}
		schemas[file] = sch
	}
	return schemas, nil
}

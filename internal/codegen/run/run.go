package run

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/willabides/jsonschematogo/internal/codegen"
	"github.com/willabides/jsonschematogo/internal/schema"
)

const description = `jsonschematogo converts JSON Schema definitions into Go struct types with proper JSON tags and type mapping.`

type Cmd struct {
	Files    []string          `kong:"arg,help='JSON/YAML schema files to process'"`
	Output   string            `kong:"short=o,help='Output file path (defaults to stdout)'"`
	Package  string            `kong:"short=p,default='gen',help='Package name for generated Go code'"`
	BaseDir  string            `kong:"group=parsing,help='Base directory for resolving relative schema references'"`
	URLMap   map[string]string `kong:"placeholder='prefix=directory',group=parsing,help='URL mappings for schema references'"`
	CACert   string            `kong:"group=parsing,help='CA certificate file for HTTPS connections'"`
	Insecure bool              `kong:"group=parsing,help='Skip TLS verification for HTTPS connections'"`
	Version  kong.VersionFlag  `kong:"short=v,help='Output the version and exit'"`
}

func (cli *Cmd) Run(k *kong.Context) error {
	if len(cli.Files) == 0 {
		return fmt.Errorf("no schema files provided")
	}

	// Load all schemas and build the map for $ref resolution
	schemas, err := schema.LoadAllSchemas(cli.Files)
	if err != nil {
		return err
	}

	// Generate Go code for each entry point
	var output bytes.Buffer
	for _, file := range cli.Files {
		sch := schemas[file]
		opts := &codegen.Options{
			PackageName: cli.Package,
			Schemas:     schemas,
		}
		genErr := codegen.GenerateGoCode(&output, sch, opts)
		if genErr != nil {
			return fmt.Errorf("failed to generate code for %s: %w", file, genErr)
		}
	}

	if cli.Output == "" {
		_, writeErr := fmt.Fprintln(k.Stdout, output.String())
		return writeErr
	}

	err = os.WriteFile(cli.Output, output.Bytes(), 0o600)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}
	return nil
}

func Run(args []string, opts []kong.Option) (exitCode int) {
	done := false
	errForceDone := fmt.Errorf("force done")
	opts = append(
		opts,
		kong.Description(description),
		kong.ShortUsageOnError(),
		kong.Exit(func(i int) {
			exitCode = i
			done = true
		}),
		kong.ExplicitGroups([]kong.Group{
			{
				Key:   "parsing",
				Title: "Schema Parsing Options:",
			},
		}),
		kong.WithBeforeResolve(func() error {
			if done {
				return errForceDone
			}
			return nil
		}),
	)
	var cli Cmd
	parser := kong.Must(&cli, opts...)
	k, err := parser.Parse(args)
	if errors.Is(err, errForceDone) {
		// If we forced a done state, we don't want to run the command
		return exitCode
	}
	parser.FatalIfErrorf(err)
	if done {
		return exitCode
	}
	k.FatalIfErrorf(k.Run())
	return exitCode
}

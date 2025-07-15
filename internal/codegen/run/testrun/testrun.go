package testrun

import (
	"bytes"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/willabides/jsonschematogo/internal/codegen/run"
)

type Result struct {
	ExitCode int    `yaml:"exit_code"`
	Stdout   string `yaml:"stdout"`
	Stderr   string `yaml:"stderr"`
}

func (r Result) AssertSuccess(t *testing.T) {
	t.Helper()
	assert.Equal(t, 0, r.ExitCode)
	assert.Empty(t, r.Stderr, "stderr")
}

type Runner struct{}

func Run(args ...string) Result {
	var stdout, stderr bytes.Buffer
	opts := []kong.Option{
		kong.Name("jsonschematogo"),
		kong.Writers(&stdout, &stderr),
	}
	exitCode := run.Run(args, opts)
	return Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

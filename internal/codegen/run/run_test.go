package run_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/willabides/jsonschematogo/internal/codegen/run/testrun"
)

func TestRun(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		result := testrun.Run("../testdata/schemas/primitives.yaml")
		result.AssertSuccess(t)
	})

	t.Run("help", func(t *testing.T) {
		result := testrun.Run("-h")
		result.AssertSuccess(t)
	})

	t.Run("undefined arg", func(t *testing.T) {
		result := testrun.Run("--undefined-arg")
		assert.NotZero(t, result.ExitCode)
	})
}

func TestURL(t *testing.T) {
	u, err := url.Parse("../testdata/schemas/primitives.yaml")
	assert.NoError(t, err)
	fmt.Println(u)
}

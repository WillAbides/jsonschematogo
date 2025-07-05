package golden

import (
	"cmp"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
)

const updateGoldenKey = "UPDATE_GOLDEN"

type AssertGoldenOptions struct {
	// Path to the golden file, default: testdata/golden/<test_name><suffix>.
	// When set, GoldenfileSuffix is ignored
	Goldenfile string

	// Suffix for the golden file, default: ".txt"
	GoldenfileSuffix string
}

// AssertGolden asserts that got matches the contents of the golden file. If the golden file does not
// exist or if the environment variable UPDATE_GOLDEN is set and not empty, it writes got to the golden file first.
// Golden files are places in testdata/golden/<test_name>.<suffix>
func AssertGolden(t testing.TB, got []byte, opts *AssertGoldenOptions) bool {
	t.Helper()
	if opts == nil {
		opts = &AssertGoldenOptions{}
	}
	goldenfile := cmp.Or(opts.Goldenfile, defaultGoldenfileName(t, opts.GoldenfileSuffix))
	exists := true
	_, err := os.Stat(goldenfile)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		exists = false
	default:
		t.Error(err.Error())
		return false
	}
	if os.Getenv(updateGoldenKey) != "" || !exists {
		err = errors.Join(
			os.MkdirAll(filepath.Dir(goldenfile), 0o700),
			os.WriteFile(goldenfile, got, 0o600),
		)
		if err != nil {
			t.Error(err.Error())
			return false
		}
	}
	//nolint:gosec // this is a test helper
	want, err := os.ReadFile(goldenfile)
	if err != nil {
		t.Error(err.Error())
		return false
	}
	diff := gocmp.Diff(string(want), string(got))
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
		return false
	}
	return true
}

func defaultGoldenfileName(t testing.TB, suffix string) string {
	if suffix == "" {
		suffix = ".txt"
	}
	name := strings.ReplaceAll(strings.ReplaceAll(t.Name(), "/", "-"), " ", "_")
	return filepath.Join("testdata", "golden", name+suffix)
}

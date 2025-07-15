package testutil

import (
	"path/filepath"
	"runtime"
)

func RepoRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

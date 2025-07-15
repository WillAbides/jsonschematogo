package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/willabides/jsonschematogo/internal/codegen/run"
)

var version = "unknown"

func main() {
	exitCode := run.Run(os.Args[1:], []kong.Option{kong.Vars{"version": version}})
	os.Exit(exitCode)
}

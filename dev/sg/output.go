package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

var out = output.NewOutput(os.Stdout, output.OutputOpts{
	ForceColor: true,
	ForceTTY:   true,
})

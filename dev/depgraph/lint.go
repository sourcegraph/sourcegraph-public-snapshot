package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
	"github.com/sourcegraph/sourcegraph/dev/depgraph/lints"
)

// Handles a command of the following form:
//
// depgraph lint [pass...]
//
// Runs each of the default lints registered in the lint package and returns
// any errors that are seen with the current checkout's package import graph.
func lint(graph *graph.DependencyGraph) error {
	names := lints.DefaultLints
	if len(os.Args) > 2 {
		names = os.Args[2:]
	}

	return lints.Run(graph, names)
}

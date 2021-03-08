package lints

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// NoReachingIntoCommands returns an error for each shared package that imports a package
// from a command. This includes reaching into cmd/X from another cmd, or from shared code.
func NoReachingIntoCommands(graph *graph.DependencyGraph) error {
	violations := map[string][]string{}
	for _, pkg := range graph.Packages {
		for _, dependency := range graph.Dependencies[pkg] {
			if cmd := containingCommand(dependency); cmd != "" && cmd != containingCommand(pkg) {
				violations[dependency] = append(violations[dependency], pkg)
			}
		}
	}

	errors := make([]lintError, 0, len(violations))
	for imported, importers := range violations {
		errors = append(errors, makeReachingIntoCommandError(imported, importers))
	}

	return multi(errors)
}

func makeReachingIntoCommandError(imported string, importers []string) lintError {
	items := make([]string, 0, len(importers))
	for _, importer := range importers {
		items = append(items, fmt.Sprintf("\t- %s", importer))
	}

	return lintError{
		name:        "NoReachingIntoCommands",
		pkg:         imported,
		description: fmt.Sprintf("imported past command boundary by %d packages:\n%s", len(items), strings.Join(items, "\n")),
	}
}

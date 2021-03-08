package lints

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// NoReachingIntoCommands returns an error for each package that imports a package of a
// command that it itself is not a part of. This includes reaching into cmd/X from another
// cmd, or from shared code.
func NoReachingIntoCommands(graph *graph.DependencyGraph) error {
	violations := map[string][]string{}
	for _, pkg := range graph.Packages {
		for _, dependency := range graph.Dependencies[pkg] {
			if containingCommand(dependency) != containingCommand(pkg) && containingCommand(dependency) != "" {
				violations[dependency] = append(violations[dependency], pkg)
			}
		}
	}

	errors := make([]lintError, 0, len(violations))
	for imported, importers := range violations {
		items := make([]string, 0, len(importers))
		for _, importer := range importers {
			items = append(items, fmt.Sprintf("\t- %s", importer))
		}

		errors = append(errors, lintError{
			name:        "NoReachingIntoCommands",
			pkg:         imported,
			description: fmt.Sprintf("imported past command boundary by %d packages:\n%s", len(items), strings.Join(items, "\n")),
		})
	}

	return multi(errors)
}

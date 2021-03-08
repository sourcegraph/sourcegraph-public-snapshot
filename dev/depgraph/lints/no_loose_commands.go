package lints

import (
	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// NoLooseCommands returns an error for each main package not declared in a known command root.
func NoLooseCommands(graph *graph.DependencyGraph) error {
	return mapPackageErrors(graph, func(pkg string) (lintError, bool) {
		if !isMain(graph.PackageNames, pkg) || cmdPattern.MatchString(pkg) {
			return lintError{}, false
		}

		return lintError{
			name: "NoLooseCommands",
			pkg:  pkg,
		}, true
	})
}

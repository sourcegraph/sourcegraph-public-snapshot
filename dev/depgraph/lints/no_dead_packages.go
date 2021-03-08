package lints

import (
	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// NoDeadPackages returns an error for any package that is not importable from outside the
// repository and is not imported (transitively) by a main package.
func NoDeadPackages(graph *graph.DependencyGraph) error {
	return mapPackageErrors(graph, func(pkg string) (lintError, bool) {
		if isMain(graph.PackageNames, pkg) || isLibrary(pkg) {
			return lintError{}, false
		}

		for _, dependent := range allDependents(graph, pkg) {
			if isMain(graph.PackageNames, dependent) {
				return lintError{}, false
			}
		}

		return lintError{name: "NoDeadPackages", pkg: pkg}, true
	})
}

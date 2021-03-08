package lints

import "github.com/sourcegraph/sourcegraph/dev/depgraph/graph"

// NoUnusedSharedCommandCode returns an error for each non-private package within
// a command that is imported only by private packages within the same command.
func NoUnusedSharedCommandCode(graph *graph.DependencyGraph) error {
	return mapPackageErrors(graph, func(pkg string) (lintError, bool) {
		if isMain(graph.PackageNames, pkg) || containingCommand(pkg) == "" || isCommandPrivate(pkg) {
			// Not shared command code
			return lintError{}, false
		}

		for _, dependent := range graph.Dependents[pkg] {
			if !isCommandPrivate(dependent) {
				return lintError{}, false
			}
		}

		return lintError{name: "NoUnusedSharedCommandCode", pkg: pkg}, true
	})
}

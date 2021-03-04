package lints

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// NoSingleDependents returns an error for each package that is imported by only
// one other (heuristically unrelated) internal dependent. We skip imports from
// ancestors, descendants, and siblings as these are generally used to organize
// code into highly cohesive sub-units.
//
// Alternative ideas for implementations here:
//   - Trigger when the path distance exceeds a threshold
//   - Explicitly only concern ourselves with internal vs cmd edges
func NoSingleDependents(graph *graph.DependencyGraph) error {
	var errors []lintError
	for _, pkg := range graph.Packages {
		if len(graph.Dependents[pkg]) != 1 || strings.HasPrefix(pkg, "enterprise/lib") {
			// Caught by NoDeadPackages, multiple has multiple
			// dependents, or may have external/unknown imports
			continue
		}

		if singleImportAllowed(pkg, graph.Dependents[pkg][0]) {
			continue
		}

		errors = append(errors, lintError{
			name:        "NoSingleDependents",
			pkg:         pkg,
			description: fmt.Sprintf("imported only by %s", graph.Dependents[pkg][0]),
		})
	}

	return multi(errors)
}

// isParent returns true if child is a proper (path) suffix of parent.
func isParent(child, parent string) bool {
	return parent == "" || strings.HasPrefix(child, parent+"/")
}

// singleDependentsIgnorePathPrefixes lists the packages prefixes to ignore in NoSingleDependents.
var singleDependentsIgnorePathPrefixes = []string{
	"enterprise/lib", // external
}

// singleImportAllowed returns true if the given package can be imported solely by the
// given dependent in this repository.
func singleImportAllowed(pkg, dependent string) bool {
	for _, prefix := range singleDependentsIgnorePathPrefixes {
		if strings.HasPrefix(pkg, prefix) {
			return true
		}
	}

	if isParent(pkg, dependent) || isParent(dependent, pkg) || filepath.Dir(pkg) == filepath.Dir(dependent) {
		// ancestor, descendant, or sibling
		return true
	}

	return false
}

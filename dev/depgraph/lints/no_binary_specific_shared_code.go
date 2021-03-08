package lints

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// NoBinarySpecificSharedCode returns an error for each shared package that is used
// by a single command.
func NoBinarySpecificSharedCode(graph *graph.DependencyGraph) error {
	return mapPackageErrors(graph, func(pkg string) (lintError, bool) {
		if containingCommand(pkg) != "" || isLibrary(pkg) {
			// Not shared code
			return lintError{}, false
		}

		dependentCommands := map[string]struct{}{}
		for _, dependent := range graph.Dependents[pkg] {
			dependentCommands[containingCommand(dependent)] = struct{}{}
		}
		if len(dependentCommands) != 1 {
			// Not a single import
			return lintError{}, false
		}

		var importer string
		for cmd := range dependentCommands {
			importer = cmd
		}
		if importer == "" {
			// Only imported by other internal packages
			return lintError{}, false
		}

		return lintError{
			name:        "NoBinarySpecificSharedCode",
			pkg:         pkg,
			description: fmt.Sprintf("imported only by %s", importer),
		}, true
	})
}

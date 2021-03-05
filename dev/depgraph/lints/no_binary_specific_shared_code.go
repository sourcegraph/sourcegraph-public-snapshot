package lints

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// NoBinarySpecificSharedCode returns an error for each shared package that is
// used only by a single binary.
func NoBinarySpecificSharedCode(graph *graph.DependencyGraph) error {
	var errors []lintError
outer:
	for _, pkg := range graph.Packages {
		if strings.HasPrefix(pkg, "enterprise/lib") {
			// May have external/unknown imports
			continue
		}
		if containingCommand(pkg) != "" {
			// Not shared code
			continue
		}

		var firstImporter *string
		for _, p := range graph.Dependents[pkg] {
			if cmd := containingCommand(p); firstImporter == nil {
				firstImporter = &cmd
			} else if cmd != *firstImporter {
				continue outer
			}
		}

		if firstImporter == nil || *firstImporter == "" {
			// No or multiple distinct importers
			continue
		}

		errors = append(errors, lintError{
			name:        "NoSingleUseSharedCode",
			pkg:         pkg,
			description: fmt.Sprintf("imported only by %s", *firstImporter),
		})
	}

	return multi(errors)
}

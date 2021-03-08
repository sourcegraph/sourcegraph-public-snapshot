package lints

import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

var validCommandPattern = regexp.MustCompile(`^(?:enterprise/)?(?:dev|cmd)/([^/]+)$`)

// NoLooseCommands returns an error for each package
func NoLooseCommands(graph *graph.DependencyGraph) error {
	var errors []lintError
	for _, pkg := range graph.Packages {
		if isMain(graph.PackageNames, pkg) && !validCommandPattern.MatchString(pkg) {
			errors = append(errors, lintError{
				name: "NoLooseCommands",
				pkg:  pkg,
			})
		}
	}

	return multi(errors)
}

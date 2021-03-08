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
		for _, name := range graph.PackageNames[pkg] {
			if name == "main" && !validCommandPattern.MatchString(pkg) {
				errors = append(errors, lintError{
					name: "NoLooseCommands",
					pkg:  pkg,
				})
			}
		}
	}

	return multi(errors)
}

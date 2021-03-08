package lints

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// NoSingleDependents returns an error for each package that is not imported
// by any other package unless it is an entrypoint for a command or a known
// and explicitly listed dev tool.
func NoDeadPackages(graph *graph.DependencyGraph) error {
	var errors []lintError
	for _, pkg := range graph.Packages {
		if len(graph.Dependents[pkg]) == 0 && !isMain(graph.PackageNames, pkg) && !deadPackageAllowed(pkg) {
			errors = append(errors, lintError{name: "NoDeadPackages", pkg: pkg})
		}
	}

	return multi(errors)
}

// deadPackageIgnorePathPrefixes lists the packages prefixes to ignore in NoDeadPackages.
var deadPackageIgnorePathPrefixes = []string{
	"enterprise/lib",                       // external
	"client/browser/code-intel-extensions", // embedded repository
	"docker-images",                        // contains binaries
	"enterprise/dev/ci",                    // contains binaries
}

// deadPackageIgnorePackages lists the packages to ignore in NoDeadPackages.
var deadPackageIgnorePackages = []string{
	"",                            // root sg/sg/ package
	"monitoring",                  // known dev binary
	"internal/database/schemadoc", // known dev binary
}

// deadPackageAllowed returns true if the given package can be non-imported by another
// package in this repository. This gives us a way to exclude binary entrypoints,
// development tooling, and code that is explicitly meant to be imported externally.
func deadPackageAllowed(pkg string) bool {
	for _, prefix := range deadPackageIgnorePathPrefixes {
		if strings.HasPrefix(pkg, prefix) {
			return true
		}
	}

	for _, ignored := range deadPackageIgnorePackages {
		if pkg == ignored {
			return true
		}
	}

	return false
}

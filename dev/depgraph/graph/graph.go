package graph

import (
	"sort"
	"strings"
)

// DependencyGraph encodes the import relationships between packages within
// the sourcegraph/sourcegraph repository.
type DependencyGraph struct {
	// Packages is a de-duplicated and ordered list of all package.
	Packages []string

	// Dependencies is a map from package name to the set of packages it imports.
	Dependencies map[string][]string

	// Dependents is a map from package name to the set of packages that import it.
	Dependents map[string][]string
}

// Load returns a dependency graph constructed by running `go list` to get an initial
// set of packages, then running `go list` on each package to get a list of its imports.
func Load() (*DependencyGraph, error) {
	pkgs, err := listPackages()
	if err != nil {
		return nil, err
	}
	for i, pkg := range pkgs {
		pkgs[i] = trimPackage(pkg)
	}

	imports, err := importsOfPackages(pkgs)
	if err != nil {
		return nil, err
	}

	reverseImports := make(map[string][]string, len(imports))
	for pkg, dependencies := range imports {
		for _, dependency := range dependencies {
			reverseImports[dependency] = append(reverseImports[dependency], pkg)
		}
	}

	allPackages := make(map[string]struct{}, len(imports))
	for k := range imports {
		allPackages[k] = struct{}{}
	}
	for k := range reverseImports {
		allPackages[k] = struct{}{}
	}

	packages := make([]string, 0, len(allPackages))
	for pkg := range allPackages {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	return &DependencyGraph{
		Packages:     packages,
		Dependencies: imports,
		Dependents:   reverseImports,
	}, nil
}

// listPackages lists all packages under sourcegraph/sourcegraph. This assumes that the
// binary is being run from the root of the repository.
func listPackages() ([]string, error) {
	out, err := runGo("list", "./...")
	if err != nil {
		return nil, err
	}

	return strings.Split(out, "\n"), nil
}

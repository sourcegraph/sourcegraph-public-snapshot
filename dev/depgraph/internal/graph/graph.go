package graph

import (
	"sort"
)

// DependencyGraph encodes the import relationships between packages within
// the sourcegraph/sourcegraph repository.
type DependencyGraph struct {
	// Packages is a de-duplicated and ordered list of all package paths.
	Packages []string

	// PackageNames is a map from package paths to their declared names.
	PackageNames map[string][]string

	// Dependencies is a map from package path to the set of packages it imports.
	Dependencies map[string][]string

	// Dependents is a map from package path to the set of packages that import it.
	Dependents map[string][]string
}

// Load returns a dependency graph constructed by walking the source tree of the
// sg/sg repository and parsing the imports out of all file with a .go extension.
func Load(root string) (*DependencyGraph, error) {
	packageMap, err := listPackages(root)
	if err != nil {
		return nil, err
	}
	names, err := parseNames(root, packageMap)
	if err != nil {
		return nil, err
	}
	imports, err := parseImports(root, packageMap)
	if err != nil {
		return nil, err
	}
	reverseImports := reverseGraph(imports)

	allPackages := make(map[string]struct{}, len(names)+len(imports)+len(reverseImports))
	for pkg := range names {
		allPackages[pkg] = struct{}{}
	}
	for pkg := range imports {
		allPackages[pkg] = struct{}{}
	}
	for pkg := range reverseImports {
		allPackages[pkg] = struct{}{}
	}

	packages := make([]string, 0, len(allPackages))
	for pkg := range allPackages {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	return &DependencyGraph{
		Packages:     packages,
		PackageNames: names,
		Dependencies: imports,
		Dependents:   reverseImports,
	}, nil
}

// reverseGraph returns the given graph with all edges reversed.
func reverseGraph(graph map[string][]string) map[string][]string {
	reverseGraph := make(map[string][]string, len(graph))
	for pkg, dependencies := range graph {
		for _, dependency := range dependencies {
			reverseGraph[dependency] = append(reverseGraph[dependency], pkg)
		}
	}

	return reverseGraph
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

// Handles a command of the following form:
//
// depgraph trace {internal/honey} [-dependencies=false] [-dependents=false]
//
// Outputs a graph in dot format (convert with `dot -Tsvg {file.dot} -o file.svg`).
func trace(graph *graph.DependencyGraph) error {
	if len(os.Args) < 3 {
		fmt.Printf("No path supplied.\n")
		os.Exit(1)
	}
	pkg := os.Args[2]

	dependenciesFlag := flag.Bool("dependencies", true, "Whether to show dependencies (default true)")
	dependentsFlag := flag.Bool("dependents", true, "Whether to show dependents (default true)")
	if err := flag.CommandLine.Parse(os.Args[3:]); err != nil {
		return err
	}

	packages, dependencyEdges, dependentEdges := traceWalkGraph(
		graph,
		pkg,
		*dependenciesFlag,
		*dependentsFlag,
	)

	traceDotify(packages, dependencyEdges, dependentEdges)
	return nil
}

// traceWalkGraph traverses the given dependency graph in both directions and returns a
// set of packages and edges (separated by traversal direction) forming the dependency
// graph around the given blessed package.
func traceWalkGraph(graph *graph.DependencyGraph, pkg string, traverseDependencies, traverseDependents bool) (packages []string, dependencyEdges, dependentEdges map[string][]string) {
	if traverseDependencies {
		newPackages, newEdges := traceTraverse(pkg, graph.Dependencies)
		packages = append(packages, newPackages...)
		dependencyEdges = newEdges
	}
	if traverseDependents {
		newPackages, newEdges := traceTraverse(pkg, graph.Dependents)
		packages = append(packages, newPackages...)
		dependentEdges = newEdges
	}

	return
}

// traceTraverse returns a set of packages and edges forming the dependency graph around
// the given blessed package using the given relation to traverse the dependency graph in
// one direction from the given package root.
func traceTraverse(pkg string, relation map[string][]string) (packages []string, edges map[string][]string) {
	frontier := relation[pkg]
	packageMap := map[string]int{pkg: 0}
	edges = map[string][]string{pkg: relation[pkg]}

	for depth := 0; len(frontier) > 0; depth++ {
		nextFrontier := []string{}
		for _, pkg := range frontier {
			if _, ok := packageMap[pkg]; ok {
				continue
			}
			packageMap[pkg] = depth

			edges[pkg] = append(edges[pkg], relation[pkg]...)
			nextFrontier = append(nextFrontier, relation[pkg]...)
		}

		frontier = nextFrontier
	}

	packages = make([]string, 0, len(packages))
	for k := range packageMap {
		packages = append(packages, k)
	}
	sort.Strings(packages)

	return packages, edges
}

// traceDotify serializes the given package and edge data into a DOT-formatted graph.
func traceDotify(packages []string, dependencyEdges, dependentEdges map[string][]string) {
	fmt.Printf("digraph deps {\n")

	pathTree := &treeNode{
		children: map[string]*treeNode{
			"": nestPaths("", getAllIntermediatePaths(packages)),
		},
	}
	displayPackageTree(pathTree, packages, 1)

	for k, vs := range dependencyEdges {
		for _, v := range vs {
			fmt.Printf("    %s -> %s [fillcolor=red]\n", normalize(k), normalize(v))
		}
	}
	for k, vs := range dependentEdges {
		for _, v := range vs {
			fmt.Printf("    %s -> %s [fillcolor=blue]\n", normalize(v), normalize(k))
		}
	}

	fmt.Printf("}\n")
}

func displayPackageTree(node *treeNode, packages []string, level int) {
	for pkg, children := range node.children {
		if len(children.children) == 0 {
			fmt.Printf("%s%s [label=\"%s\"]\n", indent(level), normalize(pkg), labelize(pkg))
		} else {
			fmt.Printf("%ssubgraph cluster_%s {\n", indent(level), normalize(pkg))
			fmt.Printf("%slabel = \"%s\"\n", indent(level+1), labelize(pkg))

			found := false
			for _, node := range packages {
				if pkg == node {
					found = true
					break
				}
			}
			if found {
				fmt.Printf("%s%s [label=\"%s\"]\n", indent(level+1), normalize(pkg), labelize(pkg))
			}

			displayPackageTree(children, packages, level+1)
			fmt.Printf("%s}\n", indent(level))
		}
	}
}

func indent(level int) string {
	return strings.Repeat(" ", 4*level)
}

// getAllIntermediatePaths calls getIntermediatePaths on the given values, then
// deduplicates and orders the results.
func getAllIntermediatePaths(pkgs []string) []string {
	uniques := map[string]struct{}{}
	for _, pkg := range pkgs {
		for _, pkg := range getIntermediatePaths(pkg) {
			uniques[pkg] = struct{}{}
		}
	}

	flattened := make([]string, 0, len(uniques))
	for key := range uniques {
		flattened = append(flattened, key)
	}
	sort.Strings(flattened)

	return flattened
}

// getIntermediatePaths returns all proper (path) prefixes of the given package.
// For example, a/b/c will return the set containing {a/b/c, a/b, a}.
func getIntermediatePaths(pkg string) []string {
	if dirname := filepath.Dir(pkg); dirname != "." {
		return append([]string{pkg}, getIntermediatePaths(dirname)...)
	}

	return []string{pkg}
}

type treeNode struct {
	children map[string]*treeNode
}

// nestPaths constructs the treeNode forming the subtree rooted at the given prefix.
func nestPaths(prefix string, pkgs []string) *treeNode {
	nodes := map[string]*treeNode{}

outer:
	for _, pkg := range pkgs {
		// Skip self and anything not within the current prefix
		if pkg == prefix || !isParent(pkg, prefix) {
			continue
		}

		// Skip anything already claimed by this level
		for prefix := range nodes {
			if isParent(pkg, prefix) {
				continue outer
			}
		}

		nodes[pkg] = nestPaths(pkg, pkgs)
	}

	return &treeNode{nodes}
}

// isParent returns true if child is a proper (path) suffix of parent.
func isParent(child, parent string) bool {
	return parent == "" || strings.HasPrefix(child, parent+"/")
}

// labelize returns the last segment of the given package path.
func labelize(pkg string) string {
	if pkg == "" {
		pkg = "sg/sg"
	}

	return filepath.Base(pkg)
}

var nonAlphaPattern = regexp.MustCompile(`[^a-z]`)

// normalize makes a package path suitable for a dot node name.
func normalize(pkg string) string {
	if pkg == "" {
		pkg = "sg/sg"
	}

	return nonAlphaPattern.ReplaceAllString(pkg, "_")
}

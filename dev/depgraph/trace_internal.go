package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
	"github.com/sourcegraph/sourcegraph/dev/depgraph/visualization"
)

// Handles a command of the following form:
//
// depgraph trace-internal {enterprise/cmd/frontend}
//
// Outputs a graph in dot format (convert with `dot -Tsvg {file.dot} -o file.svg`).
func traceInternal(graph *graph.DependencyGraph) error {
	if len(os.Args) < 3 {
		fmt.Printf("No path supplied.\n")
		os.Exit(1)
	}
	prefix := os.Args[2]

	packages := make([]string, 0, len(graph.Packages))
	for _, pkg := range graph.Packages {
		if strings.HasPrefix(pkg, prefix) {
			packages = append(packages, pkg)
		}
	}

	dependencyEdges := map[string][]string{}
	for pkg, dependencies := range graph.Dependencies {
		if strings.HasPrefix(pkg, prefix) {
			for _, dependency := range dependencies {
				if strings.HasPrefix(dependency, prefix) {
					dependencyEdges[pkg] = append(dependencyEdges[pkg], dependency)
				}
			}
		}
	}

	fmt.Printf("%s\n", visualization.Dotify(packages, dependencyEdges, nil))
	return nil
}

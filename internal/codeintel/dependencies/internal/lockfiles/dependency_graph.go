package lockfiles

import (
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type edge struct {
	source, target reposource.PackageDependency
}

func newDependencyGraph() *dependencyGraph {
	return &dependencyGraph{
		dependencies: make(map[reposource.PackageDependency][]reposource.PackageDependency),
		edges:        map[edge]struct{}{},
	}
}

type dependencyGraph struct {
	dependencies map[reposource.PackageDependency][]reposource.PackageDependency

	edges map[edge]struct{}
}

func (dg *dependencyGraph) addPackage(pkg reposource.PackageDependency) {
	if _, ok := dg.dependencies[pkg]; !ok {
		dg.dependencies[pkg] = []reposource.PackageDependency{}
	}
}
func (dg *dependencyGraph) addDependency(a, b reposource.PackageDependency) {
	dg.dependencies[a] = append(dg.dependencies[a], b)
	dg.edges[edge{a, b}] = struct{}{}
}

func (dg *dependencyGraph) Roots() map[reposource.PackageDependency]struct{} {
	roots := make(map[reposource.PackageDependency]struct{}, len(dg.dependencies))
	for pkg := range dg.dependencies {
		roots[pkg] = struct{}{}
	}

	for edge := range dg.edges {
		delete(roots, edge.target)
	}

	return roots
}

func (dg *dependencyGraph) String() string {
	var out strings.Builder

	for root := range dg.Roots() {
		printDependencies(&out, dg, 0, root)
	}

	return out.String()
}

func printDependencies(out io.Writer, graph *dependencyGraph, level int, node reposource.PackageDependency) {
	deps, ok := graph.dependencies[node]
	if !ok || len(deps) == 0 {
		fmt.Fprintf(out, "%s%s\n", strings.Repeat("\t", level), node.RepoName())
		return
	}

	fmt.Fprintf(out, "%s%s:\n", strings.Repeat("\t", level), node.RepoName())

	for _, dep := range deps {
		printDependencies(out, graph, level+1, dep)
	}
}

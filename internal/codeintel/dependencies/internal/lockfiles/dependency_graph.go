package lockfiles

import (
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type Edge struct {
	Source, Target reposource.PackageDependency
}

func newDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		dependencies: make(map[reposource.PackageDependency][]reposource.PackageDependency),
		edges:        map[Edge]struct{}{},
	}
}

type DependencyGraph struct {
	dependencies map[reposource.PackageDependency][]reposource.PackageDependency
	edges        map[Edge]struct{}
}

func (dg *DependencyGraph) addPackage(pkg reposource.PackageDependency) {
	if _, ok := dg.dependencies[pkg]; !ok {
		dg.dependencies[pkg] = []reposource.PackageDependency{}
	}
}
func (dg *DependencyGraph) addDependency(a, b reposource.PackageDependency) {
	dg.dependencies[a] = append(dg.dependencies[a], b)
	dg.edges[Edge{a, b}] = struct{}{}
}

func (dg *DependencyGraph) Roots() map[reposource.PackageDependency]struct{} {
	roots := make(map[reposource.PackageDependency]struct{}, len(dg.dependencies))
	for pkg := range dg.dependencies {
		roots[pkg] = struct{}{}
	}

	for edge := range dg.edges {
		delete(roots, edge.Target)
	}

	return roots
}

func (dg *DependencyGraph) AllEdges() (edges []Edge) {
	for edge := range dg.edges {
		edges = append(edges, edge)
	}
	return edges
}

func (dg *DependencyGraph) String() string {
	var out strings.Builder

	for root := range dg.Roots() {
		printDependencies(&out, dg, 0, root)
	}

	return out.String()
}

func printDependencies(out io.Writer, graph *DependencyGraph, level int, node reposource.PackageDependency) {
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

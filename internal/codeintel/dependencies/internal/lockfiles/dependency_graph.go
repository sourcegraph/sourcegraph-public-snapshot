package lockfiles

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type Edge struct {
	Source, Target reposource.PackageVersion
}

func newDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		dependencies: make(map[reposource.PackageVersion][]reposource.PackageVersion),
		edges:        map[Edge]struct{}{},
	}
}

type DependencyGraph struct {
	dependencies map[reposource.PackageVersion][]reposource.PackageVersion
	edges        map[Edge]struct{}
}

func (dg *DependencyGraph) addPackage(pkg reposource.PackageVersion) {
	if _, ok := dg.dependencies[pkg]; !ok {
		dg.dependencies[pkg] = []reposource.PackageVersion{}
	}
}
func (dg *DependencyGraph) addDependency(a, b reposource.PackageVersion) {
	dg.dependencies[a] = append(dg.dependencies[a], b)
	dg.edges[Edge{a, b}] = struct{}{}
}

func (dg *DependencyGraph) Roots() (roots []reposource.PackageVersion) {
	set := make(map[reposource.PackageVersion]struct{}, len(dg.dependencies))
	for pkg := range dg.dependencies {
		set[pkg] = struct{}{}
	}

	for edge := range dg.edges {
		delete(set, edge.Target)
	}

	for k := range set {
		roots = append(roots, k)
	}

	return roots
}

func (dg *DependencyGraph) AllEdges() (edges []Edge) {
	for edge := range dg.edges {
		edges = append(edges, edge)
	}
	return edges
}

type pkgSet = map[reposource.PackageVersion]struct{}

func (dg *DependencyGraph) String() string {
	var out strings.Builder

	json.NewEncoder(&out).Encode(dg.AsMap())

	return out.String()
}

func (dg *DependencyGraph) AsMap() map[string]interface{} {
	roots := dg.Roots()
	if len(roots) == 0 {
		// If we don't have roots (because of circular dependencies), we use
		// every package as a root.
		// Ideally we'd use other information (such as the data in
		// `package.json` files) to find out what the direct dependencies are.
		//
		// TODO: this should probably go to `Roots()` with a boolean that says
		// `circular bool` and that we persist to the database.
		for pkg := range dg.dependencies {
			roots = append(roots, pkg)
		}
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].Less(roots[j]) })

	type item struct {
		pkg reposource.PackageVersion
		out map[string]interface{}
	}

	queue := make([]item, len(roots))

	out := make(map[string]interface{}, len(roots))
	for i, root := range roots {
		queue[i] = item{pkg: root, out: out}
	}

	visited := pkgSet{}
	for len(queue) != 0 {
		var current item
		current, queue = queue[0], queue[1:]

		subOut := map[string]interface{}{}
		// Write current item to its out map
		current.out[current.pkg.PackageVersionSyntax()] = subOut

		_, alreadyVisited := visited[current.pkg]
		visited[current.pkg] = struct{}{}

		deps, ok := dg.dependencies[current.pkg]
		if !ok || len(deps) == 0 || (alreadyVisited) {
			continue
		}

		sortedDeps := deps
		sort.Slice(sortedDeps, func(i, j int) bool { return sortedDeps[i].Less(sortedDeps[j]) })

		for _, dep := range sortedDeps {
			queue = append(queue, item{pkg: dep, out: subOut})
		}
	}

	return out
}

func printDependenciesToMap(out map[string]interface{}, graph *DependencyGraph, visited pkgSet, node reposource.PackageVersion) {
	_, alreadyVisited := visited[node]
	visited[node] = struct{}{}

	key := node.PackageVersionSyntax()
	val := map[string]interface{}{}
	out[key] = val

	deps, ok := graph.dependencies[node]
	if !ok || len(deps) == 0 || (alreadyVisited) {
		return
	}

	sortedDeps := deps
	sort.Slice(sortedDeps, func(i, j int) bool { return sortedDeps[i].Less(sortedDeps[j]) })

	for _, dep := range sortedDeps {
		printDependenciesToMap(val, graph, visited, dep)
	}
}

package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type Repo struct {
	ID      int
	Scheme  string
	Name    string
	Version string
}

type PackageDependency interface {
	RepoName() api.RepoName
	GitTagFromVersion() string
	Scheme() string
	PackageSyntax() string
	PackageVersion() string
}

type PackageDependencyLiteral struct {
	RepoNameValue          api.RepoName
	GitTagFromVersionValue string
	SchemeValue            string
	PackageSyntaxValue     string
	PackageVersionValue    string
}

func TestPackageDependencyLiteral(
	repoNameValue api.RepoName,
	gitTagFromVersionValue string,
	schemeValue string,
	packageSyntaxValue string,
	packageVersionValue string,
) PackageDependency {
	return PackageDependencyLiteral{
		RepoNameValue:          repoNameValue,
		GitTagFromVersionValue: gitTagFromVersionValue,
		SchemeValue:            schemeValue,
		PackageSyntaxValue:     packageSyntaxValue,
		PackageVersionValue:    packageVersionValue,
	}
}

func (d PackageDependencyLiteral) RepoName() api.RepoName    { return d.RepoNameValue }
func (d PackageDependencyLiteral) GitTagFromVersion() string { return d.GitTagFromVersionValue }
func (d PackageDependencyLiteral) Scheme() string            { return d.SchemeValue }
func (d PackageDependencyLiteral) PackageSyntax() string     { return d.PackageSyntaxValue }
func (d PackageDependencyLiteral) PackageVersion() string    { return d.PackageVersionValue }

func SerializePackageDependencies(deps []reposource.PackageVersion) []PackageDependency {
	serializableRepoDeps := make([]PackageDependency, 0, len(deps))
	for _, dep := range deps {
		serializableRepoDeps = append(serializableRepoDeps, SerializePackageDependency(dep))
	}

	return serializableRepoDeps
}

func SerializePackageDependency(dep reposource.PackageVersion) PackageDependency {
	return PackageDependencyLiteral{
		RepoNameValue:          dep.RepoName(),
		GitTagFromVersionValue: dep.GitTagFromVersion(),
		SchemeValue:            dep.Scheme(),
		PackageSyntaxValue:     dep.PackageSyntax(),
		PackageVersionValue:    dep.PackageVersion(),
	}
}

type DependencyGraph interface {
	Roots() []PackageDependency
	AllEdges() [][]PackageDependency
	Empty() bool
}

var _ DependencyGraph = DependencyGraphLiteral{}

func TestDependencyGraphLiteral(roots []PackageDependency, edges [][]PackageDependency) DependencyGraph {
	return DependencyGraphLiteral{Edges: edges, RootPkgs: roots}
}

type DependencyGraphLiteral struct {
	RootPkgs []PackageDependency
	Edges    [][]PackageDependency
}

func (dg DependencyGraphLiteral) AllEdges() [][]PackageDependency { return dg.Edges }
func (dg DependencyGraphLiteral) Roots() []PackageDependency      { return dg.RootPkgs }
func (dg DependencyGraphLiteral) Empty() bool                     { return len(dg.RootPkgs) == 0 }

func SerializeDependencyGraph(graph *lockfiles.DependencyGraph) DependencyGraph {
	if graph == nil {
		return nil
	}

	var (
		edges           = graph.AllEdges()
		serializedEdges = make([][]PackageDependency, len(edges))

		roots           = graph.Roots()
		serializedRoots = make([]PackageDependency, len(roots))
	)

	for i, edge := range edges {
		serializedEdges[i] = []PackageDependency{
			SerializePackageDependency(edge.Source),
			SerializePackageDependency(edge.Target),
		}
	}

	for i, root := range roots {
		serializedRoots[i] = SerializePackageDependency(root)
	}

	return DependencyGraphLiteral{RootPkgs: serializedRoots, Edges: serializedEdges}
}

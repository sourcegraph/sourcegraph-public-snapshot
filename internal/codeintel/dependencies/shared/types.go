package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

// IndexFidelity describes the fidelity of the indexed dependency graph we
// could get by parsing a lockfile.
type IndexFidelity string

const (
	// IndexFidelityFlat is the default fidelity if we couldn't build a graph
	// and create a flat list of all dependencies found in a lockfile.
	IndexFidelityFlat IndexFidelity = "flat"

	// IndexFidelityCircular is the fidelity if we couldn't determine the roots
	// of the dependency, since it's circular, and create a forest of
	// dependencies.
	// That means we can't say what's a direct dependency and what not, but we
	// can tell which dependency depends on which other dependency.
	IndexFidelityCircular IndexFidelity = "circular"

	// IndexFidelityGraph is the fidelity of a full graph.
	IndexFidelityGraph IndexFidelity = "graph"
)

type LockfileIndex struct {
	ID                   int
	RepositoryID         int
	Commit               string
	LockfileReferenceIDs []int
	Lockfile             string
	Fidelity             IndexFidelity
}

type Repo struct {
	ID      int
	Scheme  string
	Name    reposource.PackageName
	Version string
}

type PackageDependency interface {
	RepoName() api.RepoName
	GitTagFromVersion() string
	Scheme() string
	PackageSyntax() reposource.PackageName
	PackageVersion() string
}

type PackageDependencyLiteral struct {
	RepoNameValue          api.RepoName
	GitTagFromVersionValue string
	SchemeValue            string
	PackageSyntaxValue     reposource.PackageName
	PackageVersionValue    string
}

func TestPackageDependencyLiteral(
	repoNameValue api.RepoName,
	gitTagFromVersionValue string,
	schemeValue string,
	packageSyntaxValue reposource.PackageName,
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

func (d PackageDependencyLiteral) RepoName() api.RepoName                { return d.RepoNameValue }
func (d PackageDependencyLiteral) GitTagFromVersion() string             { return d.GitTagFromVersionValue }
func (d PackageDependencyLiteral) Scheme() string                        { return d.SchemeValue }
func (d PackageDependencyLiteral) PackageSyntax() reposource.PackageName { return d.PackageSyntaxValue }
func (d PackageDependencyLiteral) PackageVersion() string                { return d.PackageVersionValue }

func SerializePackageDependencies(deps []reposource.VersionedPackage) []PackageDependency {
	serializableRepoDeps := make([]PackageDependency, 0, len(deps))
	for _, dep := range deps {
		serializableRepoDeps = append(serializableRepoDeps, SerializePackageDependency(dep))
	}

	return serializableRepoDeps
}

func SerializePackageDependency(dep reposource.VersionedPackage) PackageDependency {
	return PackageDependencyLiteral{
		RepoNameValue:          dep.RepoName(),
		GitTagFromVersionValue: dep.GitTagFromVersion(),
		SchemeValue:            dep.Scheme(),
		PackageSyntaxValue:     dep.PackageSyntax(),
		PackageVersionValue:    dep.PackageVersion(),
	}
}

type DependencyGraph interface {
	Roots() ([]PackageDependency, bool)
	AllEdges() [][]PackageDependency
	Empty() bool
}

var _ DependencyGraph = DependencyGraphLiteral{}

func TestDependencyGraphLiteral(roots []PackageDependency, rootsUndeterminable bool, edges [][]PackageDependency) DependencyGraph {
	return DependencyGraphLiteral{Edges: edges, RootPkgs: roots, RootsUndeterminable: rootsUndeterminable}
}

type DependencyGraphLiteral struct {
	RootPkgs            []PackageDependency
	RootsUndeterminable bool

	Edges [][]PackageDependency
}

func (dg DependencyGraphLiteral) AllEdges() [][]PackageDependency { return dg.Edges }
func (dg DependencyGraphLiteral) Roots() ([]PackageDependency, bool) {
	return dg.RootPkgs, dg.RootsUndeterminable
}
func (dg DependencyGraphLiteral) Empty() bool { return len(dg.RootPkgs) == 0 }

func SerializeDependencyGraph(graph *lockfiles.DependencyGraph) DependencyGraph {
	if graph == nil {
		return nil
	}

	var (
		edges           = graph.AllEdges()
		serializedEdges = make([][]PackageDependency, len(edges))

		roots, rootsUndeterminable = graph.Roots()
		serializedRoots            = make([]PackageDependency, len(roots))
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

	return DependencyGraphLiteral{
		RootPkgs:            serializedRoots,
		RootsUndeterminable: rootsUndeterminable,
		Edges:               serializedEdges,
	}
}

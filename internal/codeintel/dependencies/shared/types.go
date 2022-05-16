package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
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

func SerializePackageDependencies(deps []reposource.PackageDependency) []PackageDependency {
	serializableRepoDeps := make([]PackageDependency, 0, len(deps))
	for _, dep := range deps {
		serializableRepoDeps = append(serializableRepoDeps, SerializePackageDependency(dep))
	}

	return serializableRepoDeps
}

func SerializePackageDependency(dep reposource.PackageDependency) PackageDependency {
	return PackageDependencyLiteral{
		RepoNameValue:          dep.RepoName(),
		GitTagFromVersionValue: dep.GitTagFromVersion(),
		SchemeValue:            dep.Scheme(),
		PackageSyntaxValue:     dep.PackageSyntax(),
		PackageVersionValue:    dep.PackageVersion(),
	}
}

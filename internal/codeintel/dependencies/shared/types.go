package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

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

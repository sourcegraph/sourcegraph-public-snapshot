package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type PackageRepoReference struct {
	ID       int
	Scheme   string
	Name     reposource.PackageName
	Versions []PackageRepoRefVersion
}

type PackageRepoRefVersion struct {
	ID           int
	PackageRefID int
	Version      string
}

type MinimalPackageRepoRef struct {
	Scheme   string
	Name     reposource.PackageName
	Versions []string
}

type PackageDependencyLiteral struct {
	RepoNameValue          api.RepoName
	GitTagFromVersionValue string
	SchemeValue            string
	PackageSyntaxValue     reposource.PackageName
	PackageVersionValue    string
}

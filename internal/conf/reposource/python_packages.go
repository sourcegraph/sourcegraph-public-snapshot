package reposource

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// TODO: should scheme be Python instead?
const poetryScheme = "poetry"

func NewPoetryDependency(name, version string) *PoetryDependency {
	return &PoetryDependency{
		Name:    name,
		Version: version,
	}
}

type PoetryDependency struct {
	Name    string
	Version string

	// The URL of the package to download. Possibly empty.
	PackageURL string
}

func (p *PoetryDependency) Scheme() string {
	return poetryScheme
}

func (p *PoetryDependency) PackageSyntax() string {
	return p.Name
}

func (p *PoetryDependency) PackageManagerSyntax() string {
	if p.Version == "" {
		return p.Name
	}
	return p.Name + "==" + p.Version
}

func (p *PoetryDependency) PackageVersion() string {
	return p.Version
}

func (p *PoetryDependency) RepoName() api.RepoName {
	return api.RepoName(poetryScheme + "/" + p.Name)
}

func (p *PoetryDependency) GitTagFromVersion() string {
	return p.Version
}

func (p *PoetryDependency) Less(other PackageDependency) bool {
	o := other.(*PoetryDependency)

	if p.Name == o.Name {
		return versionGreaterThan(p.Version, o.Version)
	}

	return p.Name > o.Name
}

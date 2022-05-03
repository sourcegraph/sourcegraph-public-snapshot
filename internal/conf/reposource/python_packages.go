package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func NewPythonDependency(name, version string) *PythonDependency {
	return &PythonDependency{
		Name:    name,
		Version: version,
	}
}

type PythonDependency struct {
	Name    string
	Version string

	// The URL of the package to download. Possibly empty.
	PackageURL string
}

func (p *PythonDependency) Scheme() string {
	return "python"
}

func (p *PythonDependency) PackageSyntax() string {
	return p.Name
}

func (p *PythonDependency) PackageManagerSyntax() string {
	if p.Version == "" {
		return p.Name
	}
	return p.Name + "==" + p.Version
}

func (p *PythonDependency) PackageVersion() string {
	return p.Version
}

func (p *PythonDependency) RepoName() api.RepoName {
	return api.RepoName("python/" + p.Name)
}

func (p *PythonDependency) GitTagFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *PythonDependency) Less(other PackageDependency) bool {
	o := other.(*PythonDependency)

	if p.Name == o.Name {
		// TODO: validate once we add a dependency source for vcs syncer.
		return versionGreaterThan(p.Version, o.Version)
	}

	return p.Name > o.Name
}

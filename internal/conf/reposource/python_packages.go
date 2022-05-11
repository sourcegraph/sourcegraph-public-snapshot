package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PythonDependency struct {
	Name    string
	Version string

	// The URL of the package to download. Possibly empty.
	PackageURL string
}

func NewPythonDependency(name, version string) *PythonDependency {
	return &PythonDependency{
		Name:    name,
		Version: version,
	}
}

// ParsePythonDependency parses a string in a '<name>(==<version>)?' format into an
// PythonDependency.
func ParsePythonDependency(dependency string) (*PythonDependency, error) {
	var dep PythonDependency
	if i := strings.LastIndex(dependency, "=="); i == -1 {
		dep.Name = dependency
	} else {
		dep.Name = strings.TrimSpace(dependency[:i])
		dep.Version = strings.TrimSpace(dependency[i+2:])
	}
	return &dep, nil
}

// ParsePythonDependencyFromRepoName is a convenience function to parse a repo name in a
// 'python/<name>(==<version>)?' format into a PythonDependency.
func ParsePythonDependencyFromRepoName(name string) (*PythonDependency, error) {
	dependency := strings.TrimPrefix(name, "python/")
	if len(dependency) == len(name) {
		return nil, errors.New("invalid python dependency repo name, missing python/ prefix")
	}
	return ParsePythonDependency(dependency)
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

func (p *PythonDependency) Description() string { return "" }

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

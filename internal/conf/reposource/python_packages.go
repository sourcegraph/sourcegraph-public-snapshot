package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PythonPackageVersion struct {
	Name    string
	Version string

	// The URL of the package to download. Possibly empty.
	PackageURL string
}

func NewPythonPackageVersion(name, version string) *PythonPackageVersion {
	return &PythonPackageVersion{
		Name:    name,
		Version: version,
	}
}

// ParsePackageVersion parses a string in a '<name>(==<version>)?' format into an
// PythonPackageVersion.
func ParsePackageVersion(dependency string) (*PythonPackageVersion, error) {
	var dep PythonPackageVersion
	if i := strings.LastIndex(dependency, "=="); i == -1 {
		dep.Name = dependency
	} else {
		dep.Name = strings.TrimSpace(dependency[:i])
		dep.Version = strings.TrimSpace(dependency[i+2:])
	}
	return &dep, nil
}

// ParsePythonPackageFromRepoName is a convenience function to parse a repo name in a
// 'python/<name>(==<version>)?' format into a PythonPackageVersion.
func ParsePythonPackageFromRepoName(name string) (*PythonPackageVersion, error) {
	dependency := strings.TrimPrefix(name, "python/")
	if len(dependency) == len(name) {
		return nil, errors.New("invalid python dependency repo name, missing python/ prefix")
	}
	return ParsePackageVersion(dependency)
}

func (p *PythonPackageVersion) Scheme() string {
	return "python"
}

func (p *PythonPackageVersion) PackageSyntax() string {
	return p.Name
}

func (p *PythonPackageVersion) PackageVersionSyntax() string {
	if p.Version == "" {
		return p.Name
	}
	return p.Name + "==" + p.Version
}

func (p *PythonPackageVersion) PackageVersion() string {
	return p.Version
}

func (p *PythonPackageVersion) Description() string { return "" }

func (p *PythonPackageVersion) RepoName() api.RepoName {
	return api.RepoName("python/" + p.Name)
}

func (p *PythonPackageVersion) GitTagFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *PythonPackageVersion) Less(other PackageVersion) bool {
	o := other.(*PythonPackageVersion)

	if p.Name == o.Name {
		// TODO: validate once we add a dependency source for vcs syncer.
		return versionGreaterThan(p.Version, o.Version)
	}

	return p.Name > o.Name
}

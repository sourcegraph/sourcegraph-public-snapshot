package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PythonVersionedPackage struct {
	Name    string
	Version string

	// The URL of the package to download. Possibly empty.
	PackageURL string
}

func NewPythonVersionedPackage(name, version string) *PythonVersionedPackage {
	return &PythonVersionedPackage{
		Name:    name,
		Version: version,
	}
}

// ParseVersionedPackage parses a string in a '<name>(==<version>)?' format into an
// PythonVersionedPackage.
func ParseVersionedPackage(dependency string) (*PythonVersionedPackage, error) {
	var dep PythonVersionedPackage
	if i := strings.LastIndex(dependency, "=="); i == -1 {
		dep.Name = dependency
	} else {
		dep.Name = strings.TrimSpace(dependency[:i])
		dep.Version = strings.TrimSpace(dependency[i+2:])
	}
	return &dep, nil
}

func ParsePythonPackageFromName(name string) (*PythonVersionedPackage, error) {
	return ParseVersionedPackage(name)
}

// ParsePythonPackageFromRepoName is a convenience function to parse a repo name in a
// 'python/<name>(==<version>)?' format into a PythonVersionedPackage.
func ParsePythonPackageFromRepoName(name string) (*PythonVersionedPackage, error) {
	dependency := strings.TrimPrefix(name, "python/")
	if len(dependency) == len(name) {
		return nil, errors.New("invalid python dependency repo name, missing python/ prefix")
	}
	return ParseVersionedPackage(dependency)
}

func (p *PythonVersionedPackage) Scheme() string {
	return "python"
}

func (p *PythonVersionedPackage) PackageSyntax() string {
	return p.Name
}

func (p *PythonVersionedPackage) VersionedPackageSyntax() string {
	if p.Version == "" {
		return p.Name
	}
	return p.Name + "==" + p.Version
}

func (p *PythonVersionedPackage) PackageVersion() string {
	return p.Version
}

func (p *PythonVersionedPackage) Description() string { return "" }

func (p *PythonVersionedPackage) RepoName() api.RepoName {
	return api.RepoName("python/" + p.Name)
}

func (p *PythonVersionedPackage) GitTagFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *PythonVersionedPackage) Less(other VersionedPackage) bool {
	o := other.(*PythonVersionedPackage)

	if p.Name == o.Name {
		// TODO: validate once we add a dependency source for vcs syncer.
		return versionGreaterThan(p.Version, o.Version)
	}

	return p.Name > o.Name
}

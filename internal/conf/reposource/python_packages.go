package reposource

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PythonVersionedPackage struct {
	Name    PackageName
	Version string
}

func NewPythonVersionedPackage(name PackageName, version string) *PythonVersionedPackage {
	return &PythonVersionedPackage{
		Name:    name,
		Version: version,
	}
}

// ParseVersionedPackage parses a string in a '<name>(==<version>)?' format into an
// PythonVersionedPackage.
func ParseVersionedPackage(dependency string) *PythonVersionedPackage {
	var dep PythonVersionedPackage
	if i := strings.LastIndex(dependency, "=="); i == -1 {
		dep.Name = PackageName(dependency)
	} else {
		dep.Name = PackageName(strings.TrimSpace(dependency[:i]))
		dep.Version = strings.TrimSpace(dependency[i+2:])
	}
	return &dep
}

func ParsePythonPackageFromName(name PackageName) *PythonVersionedPackage {
	return ParseVersionedPackage(string(name))
}

// ParsePythonPackageFromRepoName is a convenience function to parse a repo name in a
// 'python/<name>(==<version>)?' format into a PythonVersionedPackage.
func ParsePythonPackageFromRepoName(name api.RepoName) (*PythonVersionedPackage, error) {
	dependency := strings.TrimPrefix(string(name), "python/")
	if len(dependency) == len(name) {
		return nil, errors.New("invalid python dependency repo name, missing python/ prefix")
	}
	return ParseVersionedPackage(dependency), nil
}

func (p *PythonVersionedPackage) Scheme() string {
	return "python"
}

func (p *PythonVersionedPackage) PackageSyntax() PackageName {
	return p.Name
}

func (p *PythonVersionedPackage) VersionedPackageSyntax() string {
	if p.Version == "" {
		return string(p.Name)
	}
	return fmt.Sprintf("%s==%s", p.Name, p.Version)
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

package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RustVersionedPackage struct {
	Name    PackageName
	Version string
}

func NewRustVersionedPackage(name PackageName, version string) *RustVersionedPackage {
	return &RustVersionedPackage{
		Name:    name,
		Version: version,
	}
}

// ParseRustVersionedPackage parses a string in a '<name>(@version>)?' format into an
// RustVersionedPackage.
func ParseRustVersionedPackage(dependency string) *RustVersionedPackage {
	var dep RustVersionedPackage
	if i := strings.LastIndex(dependency, "@"); i == -1 {
		dep.Name = PackageName(dependency)
	} else {
		dep.Name = PackageName(strings.TrimSpace(dependency[:i]))
		dep.Version = strings.TrimSpace(dependency[i+1:])
	}
	return &dep
}

func ParseRustPackageFromName(name PackageName) *RustVersionedPackage {
	return ParseRustVersionedPackage(string(name))
}

// ParseRustPackageFromRepoName is a convenience function to parse a repo name in a
// 'crates/<name>(@<version>)?' format into a RustVersionedPackage.
func ParseRustPackageFromRepoName(name api.RepoName) (*RustVersionedPackage, error) {
	dependency := strings.TrimPrefix(string(name), "crates/")
	if len(dependency) == len(name) {
		return nil, errors.Newf("invalid Rust dependency repo name, missing crates/ prefix '%s'", name)
	}
	return ParseRustVersionedPackage(dependency), nil
}

func (p *RustVersionedPackage) Scheme() string {
	return "rust-analyzer"
}

func (p *RustVersionedPackage) PackageSyntax() PackageName {
	return p.Name
}

func (p *RustVersionedPackage) VersionedPackageSyntax() string {
	if p.Version == "" {
		return string(p.Name)
	}
	return string(p.Name) + "@" + p.Version
}

func (p *RustVersionedPackage) PackageVersion() string {
	return p.Version
}

func (p *RustVersionedPackage) Description() string { return "" }

func (p *RustVersionedPackage) RepoName() api.RepoName {
	return api.RepoName("crates/" + p.Name)
}

func (p *RustVersionedPackage) GitTagFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *RustVersionedPackage) Less(other VersionedPackage) bool {
	o := other.(*RustVersionedPackage)

	if p.Name == o.Name {
		// TODO: validate once we add a dependency source for vcs syncer.
		return versionGreaterThan(p.Version, o.Version)
	}

	return p.Name > o.Name
}

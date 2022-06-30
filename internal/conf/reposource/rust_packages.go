package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RustPackageVersion struct {
	Name    string
	Version string
}

func NewRustPackageVersion(name, version string) *RustPackageVersion {
	return &RustPackageVersion{
		Name:    name,
		Version: version,
	}
}

// ParseRustPackageVersion parses a string in a '<name>(@version>)?' format into an
// RustPackageVersion.
func ParseRustPackageVersion(dependency string) (*RustPackageVersion, error) {
	var dep RustPackageVersion
	if i := strings.LastIndex(dependency, "@"); i == -1 {
		dep.Name = dependency
	} else {
		dep.Name = strings.TrimSpace(dependency[:i])
		dep.Version = strings.TrimSpace(dependency[i+1:])
	}
	return &dep, nil
}

// ParseRustPackageFromRepoName is a convenience function to parse a repo name in a
// 'crates/<name>(@<version>)?' format into a RustPackageVersion.
func ParseRustPackageFromRepoName(name string) (*RustPackageVersion, error) {
	dependency := strings.TrimPrefix(name, "crates/")
	if len(dependency) == len(name) {
		return nil, errors.Newf("invalid Rust dependency repo name, missing crates/ prefix '%s'", name)
	}
	return ParseRustPackageVersion(dependency)
}

func (p *RustPackageVersion) Scheme() string {
	return "rust-analyzer"
}

func (p *RustPackageVersion) PackageSyntax() string {
	return p.Name
}

func (p *RustPackageVersion) PackageVersionSyntax() string {
	if p.Version == "" {
		return p.Name
	}
	return p.Name + "@" + p.Version
}

func (p *RustPackageVersion) PackageVersion() string {
	return p.Version
}

func (p *RustPackageVersion) Description() string { return "" }

func (p *RustPackageVersion) RepoName() api.RepoName {
	return api.RepoName("crates/" + p.Name)
}

func (p *RustPackageVersion) GitTagFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *RustPackageVersion) Less(other PackageVersion) bool {
	o := other.(*RustPackageVersion)

	if p.Name == o.Name {
		// TODO: validate once we add a dependency source for vcs syncer.
		return versionGreaterThan(p.Version, o.Version)
	}

	return p.Name > o.Name
}

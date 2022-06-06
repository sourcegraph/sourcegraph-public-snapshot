package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RustDependency struct {
	Name    string
	Version string
}

func NewRustDependency(name, version string) *RustDependency {
	return &RustDependency{
		Name:    name,
		Version: version,
	}
}

// ParseRustDependency parses a string in a '<name>(@version>)?' format into an
// RustDependency.
func ParseRustDependency(dependency string) (*RustDependency, error) {
	var dep RustDependency
	if i := strings.LastIndex(dependency, "@"); i == -1 {
		dep.Name = dependency
	} else {
		dep.Name = strings.TrimSpace(dependency[:i])
		dep.Version = strings.TrimSpace(dependency[i+1:])
	}
	return &dep, nil
}

// ParseRustDependencyFromRepoName is a convenience function to parse a repo name in a
// 'crates/<name>(@<version>)?' format into a RustDependency.
func ParseRustDependencyFromRepoName(name string) (*RustDependency, error) {
	dependency := strings.TrimPrefix(name, "crates/")
	if len(dependency) == len(name) {
		return nil, errors.Newf("invalid Rust dependency repo name, missing crates/ prefix '%s'", name)
	}
	return ParseRustDependency(dependency)
}

func (p *RustDependency) Scheme() string {
	return "rust-analyzer"
}

func (p *RustDependency) PackageSyntax() string {
	return p.Name
}

func (p *RustDependency) PackageManagerSyntax() string {
	if p.Version == "" {
		return p.Name
	}
	return p.Name + "@" + p.Version
}

func (p *RustDependency) PackageVersion() string {
	return p.Version
}

func (p *RustDependency) Description() string { return "" }

func (p *RustDependency) RepoName() api.RepoName {
	return api.RepoName("crates/" + p.Name)
}

func (p *RustDependency) GitTagFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *RustDependency) Less(other PackageDependency) bool {
	o := other.(*RustDependency)

	if p.Name == o.Name {
		// TODO: validate once we add a dependency source for vcs syncer.
		return versionGreaterThan(p.Version, o.Version)
	}

	return p.Name > o.Name
}

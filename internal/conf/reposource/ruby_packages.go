package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const rubyPackagesPrefix = "rubygems/"

type RubyVersionedPackage struct {
	Name    PackageName
	Version string
}

func NewRubyVersionedPackage(name PackageName, version string) *RubyVersionedPackage {
	return &RubyVersionedPackage{
		Name:    name,
		Version: version,
	}
}

// ParseRubyVersionedPackage parses a string in a '<name>(@version>)?' format into an
// RubyVersionedPackage.
func ParseRubyVersionedPackage(dependency string) *RubyVersionedPackage {
	var dep RubyVersionedPackage
	if i := strings.LastIndex(dependency, "@"); i == -1 {
		dep.Name = PackageName(dependency)
	} else {
		dep.Name = PackageName(strings.TrimSpace(dependency[:i]))
		dep.Version = strings.TrimSpace(dependency[i+1:])
	}
	return &dep
}

func ParseRubyPackageFromName(name PackageName) *RubyVersionedPackage {
	return ParseRubyVersionedPackage(string(name))
}

// ParseRubyPackageFromRepoName is a convenience function to parse a repo name in a
// 'crates/<name>(@<version>)?' format into a RubyVersionedPackage.
func ParseRubyPackageFromRepoName(name api.RepoName) (*RubyVersionedPackage, error) {
	dependency := strings.TrimPrefix(string(name), rubyPackagesPrefix)
	if len(dependency) == len(name) {
		return nil, errors.Newf("invalid Ruby dependency repo name, missing %s prefix '%s'", rubyPackagesPrefix, name)
	}
	return ParseRubyVersionedPackage(dependency), nil
}

func (p *RubyVersionedPackage) Scheme() string {
	return "scip-ruby"
}

func (p *RubyVersionedPackage) PackageSyntax() PackageName {
	return p.Name
}

func (p *RubyVersionedPackage) VersionedPackageSyntax() string {
	if p.Version == "" {
		return string(p.Name)
	}
	return string(p.Name) + "@" + p.Version
}

func (p *RubyVersionedPackage) PackageVersion() string {
	return p.Version
}

func (p *RubyVersionedPackage) Description() string { return "" }

func (p *RubyVersionedPackage) RepoName() api.RepoName {
	return api.RepoName(rubyPackagesPrefix + p.Name)
}

func (p *RubyVersionedPackage) GitTagFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *RubyVersionedPackage) Less(other VersionedPackage) bool {
	o := other.(*RubyVersionedPackage)

	if p.Name == o.Name {
		return versionGreaterThan(p.Version, o.Version)
	}

	return p.Name > o.Name
}

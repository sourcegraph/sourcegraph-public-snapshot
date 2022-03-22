package reposource

import (
	"strings"

	"golang.org/x/mod/module"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// GoDependency is a "versioned package" for use by go commands, such as `go
// get`.
//
// See also: [NOTE: Dependency-terminology]
type GoDependency struct {
	mod module.Version
}

// ParseGoModDependency parses a string in a '<name>@<version>' format into an
// GoDependency.
func ParseGoModDependency(dependency string) (*GoDependency, error) {
	var mod module.Version
	if i := strings.LastIndex(dependency, "@"); i == -1 {
		mod.Path = dependency
	} else {
		mod.Path = dependency[:i]
		mod.Version = dependency[i+1:]
	}

	err := module.Check(mod.Path, mod.Version)
	if err != nil {
		return nil, err
	}
	return &GoDependency{mod: mod}, nil
}

func (d *GoDependency) Scheme() string {
	return "go"
}

// PackageSyntax returns the name of the Go module.
func (d *GoDependency) PackageSyntax() string {
	return d.mod.Path
}

// PackageManagerSyntax returns the dependency in Go syntax. The returned string
// can (for example) be passed to `go get`.
func (d *GoDependency) PackageManagerSyntax() string {
	return d.mod.String()
}

func (d *GoDependency) PackageVersion() string {
	return d.mod.Version
}

// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
//
// The returned value is used for repo:... in queries.
func (d *GoDependency) RepoName() api.RepoName {
	return api.RepoName("go/" + d.mod.Path)
}

func (d *GoDependency) GitTagFromVersion() string {
	return "v" + d.mod.Version
}

func (d *GoDependency) Equal(other *GoDependency) bool {
	return d == other || (d != nil && other != nil && d.mod == other.mod)
}

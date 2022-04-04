package reposource

import (
	"strings"

	"golang.org/x/mod/module"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GoDependency is a "versioned package" for use by go commands, such as `go
// get`.
//
// See also: [NOTE: Dependency-terminology]
type GoDependency struct {
	Module module.Version
}

// NewGoDependency returns a GoDependency for the given module.Version.
func NewGoDependency(mod module.Version) *GoDependency {
	return &GoDependency{Module: mod}
}

// ParseGoDependency parses a string in a '<name>(@<version>)?' format into an
// GoDependency.
func ParseGoDependency(dependency string) (*GoDependency, error) {
	var mod module.Version
	if i := strings.LastIndex(dependency, "@"); i == -1 {
		mod.Path = dependency
	} else {
		mod.Path = dependency[:i]
		mod.Version = dependency[i+1:]
	}

	var err error
	if mod.Version != "" {
		err = module.Check(mod.Path, mod.Version)
	} else {
		err = module.CheckPath(mod.Path)
	}

	if err != nil {
		return nil, err
	}

	return &GoDependency{Module: mod}, nil
}

// ParseGoDependencyFromRepoName is a convenience function to parse a repo name in a
// 'go/<name>(@<version>)?' format into a GoDependency.
func ParseGoDependencyFromRepoName(name string) (*GoDependency, error) {
	dependency := strings.TrimPrefix(name, "go/")
	if len(dependency) == len(name) {
		return nil, errors.New("invalid go dependency repo name, missing go/ prefix")
	}
	return ParseGoDependency(dependency)
}

func (d *GoDependency) Scheme() string {
	return "go"
}

// PackageSyntax returns the name of the Go module.
func (d *GoDependency) PackageSyntax() string {
	return d.Module.Path
}

// PackageManagerSyntax returns the dependency in Go syntax. The returned string
// can (for example) be passed to `go get`.
func (d *GoDependency) PackageManagerSyntax() string {
	return d.Module.String()
}

func (d *GoDependency) PackageVersion() string {
	return d.Module.Version
}

// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
//
// The returned value is used for repo:... in queries.
func (d *GoDependency) RepoName() api.RepoName {
	return api.RepoName("go/" + d.Module.Path)
}

func (d *GoDependency) GitTagFromVersion() string {
	return d.Module.Version
}

func (d *GoDependency) Equal(other *GoDependency) bool {
	return d == other || (d != nil && other != nil && d.Module == other.Module)
}

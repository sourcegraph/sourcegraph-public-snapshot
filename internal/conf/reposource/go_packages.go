package reposource

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GoModule represents a Go module.
type GoModule struct {
	// Required name for a module, always non-empty.
	name string
}

// PackageSyntax returns the formal syntax of the Go module.
func (pkg *GoModule) PackageSyntax() string {
	return pkg.name
}

// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
//
// The returned value is used for repo:... in queries.
func (pkg *GoModule) RepoName() api.RepoName {
	return api.RepoName("gomod/" + pkg.name)
}

// GoDependency is a "versioned package" for use by go commands, such as `go
// get`.
//
// See also: [NOTE: Dependency-terminology]
type GoDependency struct {
	*GoModule

	// Version is the version or tag (such as "latest") for a dependency.
	Version string
}

// ParseGoModDependency parses a string in a '<name> <version> <checksum>' format
// into an GoDependency.
//
// We do not yet support verifying checksum of a module.
func ParseGoModDependency(dependency string) (*GoDependency, error) {
	fields := strings.Fields(dependency)
	if len(fields) < 2 {
		return nil, errors.Errorf("want at least 2 fields but got %d", len(fields))
	}

	return &GoDependency{
		GoModule: &GoModule{
			name: fields[0],
		},
		Version: strings.TrimSuffix(strings.TrimPrefix(fields[1], "v"), "/go.mod"),
	}, nil
}

func (d *GoDependency) Scheme() string {
	return "go"
}

// PackageManagerSyntax returns the dependency in Go syntax. The returned string
// can (for example) be passed to `go get`.
func (d *GoDependency) PackageManagerSyntax() string {
	return fmt.Sprintf("%s@%s", d.PackageSyntax(), d.GitTagFromVersion())
}

func (d *GoDependency) PackageVersion() string {
	return d.Version
}

func (d *GoDependency) GitTagFromVersion() string {
	return "v" + d.Version
}

package reposource

import (
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GoVersionedPackage is a "versioned package" for use by go commands, such as `go
// get`.
//
// See also: [NOTE: Dependency-terminology]
type GoVersionedPackage struct {
	Module module.Version
}

// NewGoVersionedPackage returns a GoVersionedPackage for the given module.Version.
func NewGoVersionedPackage(mod module.Version) *GoVersionedPackage {
	return &GoVersionedPackage{Module: mod}
}

// ParseGoVersionedPackage parses a string in a '<name>(@<version>)?' format into an
// GoVersionedPackage.
func ParseGoVersionedPackage(dependency string) (*GoVersionedPackage, error) {
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

	return &GoVersionedPackage{Module: mod}, nil
}

func ParseGoDependencyFromName(name PackageName) (*GoVersionedPackage, error) {
	return ParseGoVersionedPackage(string(name))
}

// ParseGoDependencyFromRepoName is a convenience function to parse a repo name in a
// 'go/<name>(@<version>)?' format into a GoVersionedPackage.
func ParseGoDependencyFromRepoName(name api.RepoName) (*GoVersionedPackage, error) {
	dependency := strings.TrimPrefix(string(name), "go/")
	if len(dependency) == len(name) {
		return nil, errors.New("invalid go dependency repo name, missing go/ prefix")
	}
	return ParseGoVersionedPackage(dependency)
}

func (d *GoVersionedPackage) Scheme() string {
	return "go"
}

// PackageSyntax returns the name of the Go module.
func (d *GoVersionedPackage) PackageSyntax() PackageName {
	return PackageName(d.Module.Path)
}

// VersionedPackageSyntax returns the dependency in Go syntax. The returned string
// can (for example) be passed to `go get`.
func (d *GoVersionedPackage) VersionedPackageSyntax() string {
	return d.Module.String()
}

func (d *GoVersionedPackage) PackageVersion() string {
	return d.Module.Version
}

// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
//
// The returned value is used for repo:... in queries.
func (d *GoVersionedPackage) RepoName() api.RepoName {
	return api.RepoName("go/" + d.Module.Path)
}

func (d *GoVersionedPackage) Description() string { return "" }

func (d *GoVersionedPackage) GitTagFromVersion() string {
	return d.Module.Version
}

func (d *GoVersionedPackage) Equal(o *GoVersionedPackage) bool {
	return d == o || (d != nil && o != nil && d.Module == o.Module)
}

// Less sorts d against other by Path, breaking ties by comparing Version fields.
// The Version fields are interpreted as semantic versions (using semver.Compare)
// optionally followed by a tie-breaking suffix introduced by a slash character,
// like in "v0.0.1/go.mod". Copied from golang.org/x/mod.
func (d *GoVersionedPackage) Less(other VersionedPackage) bool {
	o := other.(*GoVersionedPackage)

	if d.Module.Path != o.Module.Path {
		return d.Module.Path > o.Module.Path
	}
	// To help go.sum formatting, allow version/file.
	// Compare semver prefix by semver rules,
	// file by string order.
	vi := d.Module.Version
	vj := o.Module.Version
	var fi, fj string
	if k := strings.Index(vi, "/"); k >= 0 {
		vi, fi = vi[:k], vi[k:]
	}
	if k := strings.Index(vj, "/"); k >= 0 {
		vj, fj = vj[:k], vj[k:]
	}
	if vi != vj {
		return semver.Compare(vi, vj) > 0
	}
	return fi > fj
}

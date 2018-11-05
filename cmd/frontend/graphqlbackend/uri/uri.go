// Package uri parses and generates URIs that identify files in a repository.
package uri

import (
	"net/url"
	pathpkg "path"
	"strconv"
	"strings"
)

// A URI identifies a file in a repository.
type URI string

// Components are the parts of a URI.
type Components struct {
	Repo int32  // the ID of the repository
	Rev  string // the revision (Git revspec)
	Path string // the file path (no leading "/")
}

// File returns a URI that identifies a file.
func File(components *Components) string {
	return (&url.URL{
		Scheme: "repo",
		Host:   strconv.Itoa(int(components.Repo)),
		Path:   pathpkg.Join("/", components.Rev, strings.TrimPrefix(components.Path, "/")),
	}).String()
}

// Resolver resolves URIs. URI resolution requires external state (such as the set of existing
// repository names) and can't be done statically.
type Resolver struct{}

// Resolve resolves a URI to its components.
func (r *Resolver) Resolve(uriStr string) (*Components, error) {
	// TODO
}

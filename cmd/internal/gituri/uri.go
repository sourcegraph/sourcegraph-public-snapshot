package gituri

import (
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// A URI is a wrapper around url.URL that makes it easier to get and
// manipulate Sourcegraph-specific components. All URIs are valid
// URLs, but Sourcegraph assigns special meaning to certain URL components as described below.
//
// Sourcegraph URIs can refer to repos (at an optional revision), or a
// file or directory thereof.
//
// The format is "CLONEURL?REV#PATH". For example:
//
//   git://github.com/facebook/react?master
//   git://github.com/gorilla/mux?HEAD
//   git://github.com/golang/go?0dc31fb#src/net/http/server.go
//   git://github.com/golang/tools?79f4a1#godoc/page.go
//
// A Sourcegraph URI is not guaranteed (or intended) to be a unique or
// canonical reference to a resource. A repository can be clonable at
// several different URLs, and any of them can be used in the URI. A
// given file in a repository has any number of URIs that refer to it
// (e.g., using the branch name vs. the commit ID, using clean
// vs. non-clean file paths, etc.).
type URI struct {
	url.URL
}

// Parse parses uriStr to a URI. The uriStr should be an absolute URL.
func Parse(uriStr string) (*URI, error) {
	u, err := url.Parse(uriStr)
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, &url.Error{Op: "gituri.Parse", URL: uriStr, Err: errors.New("sourcegraph URI must be absolute")}
	}
	return &URI{*u}, nil
}

// CloneURL returns the repository clone URL component of the URI.
func (u *URI) CloneURL() *url.URL {
	return &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   u.Path,
	}
}

// Repo returns the repository name (e.g., "github.com/foo/bar").
func (u *URI) Repo() api.RepoName { return api.RepoName(u.Host + strings.TrimPrefix(u.Path, ".git")) }

// Rev returns the repository revision component of the URI (the raw
// query string).
func (u *URI) Rev() string { return u.RawQuery }

// FilePath returns the cleaned file path component of the URI (in the
// URL fragment). Leading slashes are removed. If it is ".", an empty
// string is returned.
func (u *URI) FilePath() string { return cleanPath(u.Fragment) }

// ResolveFilePath returns the cleaned file path component obtained by
// appending p to the URI's file path. It is called "resolve" not
// "join" because it strips p's leading slash (if any).
func (u *URI) ResolveFilePath(p string) string {
	return cleanPath(path.Join(u.FilePath(), strings.TrimPrefix(p, "/")))
}

// WithFilePath returns a copy of u with the file path p overwriting
// the existing file path (if any).
func (u *URI) WithFilePath(p string) *URI {
	copy := *u
	copy.Fragment = cleanPath(p)
	return &copy
}

func cleanPath(p string) string {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	if p == "." {
		p = ""
	}
	return p
}

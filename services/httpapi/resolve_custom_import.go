package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/gddo/gosrc"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/resolveutil"
)

type resolvedPath struct {
	Path string
}

// isGoRepoPath returns whether pkg is (likely to be) a Go stdlib
// package import path.
func isGoRepoPath(pkg string) bool {
	// If no path components have a ".", then guess that it's a Go
	// stdlib package.
	parts := strings.Split(pkg, "/")
	for _, p := range parts {
		if strings.Contains(p, ".") {
			return false
		}
	}
	return true
}

// serveResolveCustomImportsInfo returns the def/ref info path after resolving custom import paths
func serveResolveCustomImportsInfo(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	repo := q.Get("repo")
	pkg := q.Get("pkg")
	def := q.Get("def")

	// Attempt to resolve custom import paths
	resolvedRepo, err := resolveutil.ResolveCustomImportPath(repo)
	if err == nil {
		repo = resolvedRepo.RepoURI
	}

	resolvedPkg, err := resolveutil.ResolveCustomImportPath(pkg)
	if err == nil {
		pkg = resolvedPkg.CanonicalImportPath
	}

	if repo == "" || pkg == "" || def == "" {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("repo, pkg, and def must be specified in query string")}
	}

	return writeJSON(w, resolvedPath{Path: fmt.Sprintf("/%s/-/def/GoPackage/%s/-/%s/-/info", repo, pkg, def)})
}

// serveResolveCustomImportsInfo returns the def/ref info path after resolving custom import paths
func serveResolveCustomImportsTree(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	repo := q.Get("repo")
	pkg := q.Get("pkg")
	var path string

	// Attempt to resolve custom import paths
	resolvedRepo, err := resolveutil.ResolveCustomImportPath(repo)
	if err == nil {
		repo = resolvedRepo.RepoURI
	}

	resolvedPkg, err := resolveutil.ResolveCustomImportPath(pkg)
	if err == nil {
		pkg = resolvedPkg.CanonicalImportPath
	}

	if gosrc.IsGoRepoPath(pkg) {
		path = "/src/" + pkg
	} else {
		path = strings.TrimPrefix(pkg, repo)
	}

	if repo == "" || pkg == "" {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("repo, pkg, and path must be specified in query string")}
	}

	return writeJSON(w, resolvedPath{Path: fmt.Sprintf("/%s/-/tree%s", repo, path)})
}

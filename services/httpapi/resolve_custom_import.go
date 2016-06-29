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

	return writeJSON(w, resolvedPath{Path: fmt.Sprintf("/%s/-/info/GoPackage/%s/-/%s", repo, pkg, def)})
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

	if path == "" {
		// if pkg == repo, thus path == "", the Path returned from this API should be the pathname to the repo
		return writeJSON(w, resolvedPath{Path: fmt.Sprintf("/%s", repo)})
	}

	return writeJSON(w, resolvedPath{Path: fmt.Sprintf("/%s/-/tree%s", repo, path)})
}

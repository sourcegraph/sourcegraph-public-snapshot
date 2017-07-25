package ui2

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// TODO(slimsag): tests for everything in this file.

// pageVars are passed to JS via window.pageVars; this is distinct from
// window.context (JSContext) in the fact that this data is shared between
// template handlers and the JS code (where performing a round-trip would be
// silly). It can also only be present for some pages, whereas window.context
// is for all pages.
type pageVars struct {
	ResolvedRev string // absolute revision of current page (on any repo page).
}

type Common struct {
	Context       jscontext.JSContext
	PageVars      *pageVars
	AssetURL      string
	RepoShortName string
	Repo          *sourcegraph.Repo
	RevSpec       sourcegraph.RepoRevSpec
}

// repoShortName trims the first path element of the given repo uri if it has
// at least two path components.
func repoShortName(uri string) string {
	split := strings.Split(uri, "/")
	if len(split) < 2 {
		return uri
	}
	return strings.Join(split[1:], "/")
}

func newCommon(r *http.Request) (*Common, error) {
	// TODO(slimsag): handle auto cloning repositories here?
	//
	// TODO: handle http://localhost:3080/github.com/docker/docker redirect to moby/moby
	repo, revSpec, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}
	return &Common{
		Context: jscontext.NewJSContextFromRequest(r),
		PageVars: &pageVars{
			ResolvedRev: revSpec.CommitID,
		},
		AssetURL:      assets.URL("/").String(),
		RepoShortName: repoShortName(repo.URI),
		Repo:          repo,
		RevSpec:       revSpec,
	}, nil
}

func serveHome(w http.ResponseWriter, r *http.Request) error {
	return renderTemplate(w, "home.html", nil)
}

// treeView is the data structure shared/treeview.html expects.
type treeView struct {
	RepoURI, Rev, Dir string
	Files             []os.FileInfo
}

func (t *treeView) URL(i os.FileInfo) string {
	routeName := routeBlob
	if i.IsDir() {
		routeName = routeTree
	}
	return urlTo(routeName, "Repo", t.RepoURI, "Rev", t.Rev, "Path", t.Dir+"/"+i.Name()).String()
}

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(r)
	if err != nil {
		return err
	}

	vcsrepo, err := localstore.RepoVCS.Open(r.Context(), common.Repo.ID)
	if err != nil {
		return err
	}

	dir := "/"
	files, err := vcsrepo.ReadDir(r.Context(), vcs.CommitID(common.RevSpec.CommitID), dir, false)
	if err != nil {
		return err
	}

	return renderTemplate(w, "repo.html", &struct {
		*Common
		TreeView *treeView
	}{
		Common: common,
		TreeView: &treeView{
			RepoURI: common.Repo.URI,
			Dir:     dir,
			Files:   files,
		},
	})
}

func serveTree(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(r)
	if err != nil {
		return err
	}

	vcsrepo, err := localstore.RepoVCS.Open(r.Context(), common.Repo.ID)
	if err != nil {
		return err
	}

	dir := mux.Vars(r)["Path"]
	files, err := vcsrepo.ReadDir(r.Context(), vcs.CommitID(common.RevSpec.CommitID), dir, false)
	if err != nil {
		return err
	}

	return renderTemplate(w, "tree.html", &struct {
		*Common
		TreeView *treeView
	}{
		Common: common,
		TreeView: &treeView{
			RepoURI: common.Repo.URI,
			Dir:     dir,
			Files:   files,
		},
	})
}

// blobView is the data structure shared/blobview.html expects.
type blobView struct {
	Path, Name string
	Lines      []string
}

func serveBlob(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(r)
	if err != nil {
		return err
	}

	vcsrepo, err := localstore.RepoVCS.Open(r.Context(), common.Repo.ID)
	if err != nil {
		return err
	}

	fp := mux.Vars(r)["Path"]
	file, err := vcsrepo.ReadFile(r.Context(), vcs.CommitID(common.RevSpec.CommitID), fp)
	if err != nil {
		return err
	}

	return renderTemplate(w, "blob.html", &struct {
		*Common
		BlobView *blobView
	}{
		Common: common,
		BlobView: &blobView{
			Path:  fp,
			Name:  path.Base(fp),
			Lines: strings.Split(string(file), "\n"),
		},
	})
}

package ui2

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// TODO(slimsag): tests for everything in this file.

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
	// TODO(slimsag): handle auto cloning repositories here?
	//
	// TODO: handle http://localhost:3080/github.com/docker/docker redirect to moby/moby
	repo, revSpec, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	vcsrepo, err := localstore.RepoVCS.Open(r.Context(), repo.ID)
	if err != nil {
		return err
	}

	dir := "/"
	files, err := vcsrepo.ReadDir(r.Context(), vcs.CommitID(revSpec.CommitID), dir, false)
	if err != nil {
		return err
	}

	return renderTemplate(w, "repo.html", &struct {
		AssetURL string
		Repo     *sourcegraph.Repo
		RevSpec  sourcegraph.RepoRevSpec
		TreeView *treeView
	}{
		AssetURL: assets.URL("/").String(),
		Repo:     repo,
		RevSpec:  revSpec,
		TreeView: &treeView{
			RepoURI: repo.URI,
			Dir:     dir,
			Files:   files,
		},
	})
}

func serveTree(w http.ResponseWriter, r *http.Request) error {
	// TODO(slimsag): handle auto cloning repositories here?
	repo, revSpec, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	vcsrepo, err := localstore.RepoVCS.Open(r.Context(), repo.ID)
	if err != nil {
		return err
	}

	dir := mux.Vars(r)["Path"]
	files, err := vcsrepo.ReadDir(r.Context(), vcs.CommitID(revSpec.CommitID), dir, false)
	if err != nil {
		return err
	}

	return renderTemplate(w, "tree.html", &struct {
		AssetURL string
		Repo     *sourcegraph.Repo
		RevSpec  sourcegraph.RepoRevSpec
		TreeView *treeView
	}{
		AssetURL: assets.URL("/").String(),
		Repo:     repo,
		RevSpec:  revSpec,
		TreeView: &treeView{
			RepoURI: repo.URI,
			Dir:     dir,
			Files:   files,
		},
	})
}

func serveBlob(w http.ResponseWriter, r *http.Request) error {
	// TODO(slimsag): handle auto cloning repositories here?
	repo, revSpec, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	vcsrepo, err := localstore.RepoVCS.Open(r.Context(), repo.ID)
	if err != nil {
		return err
	}

	path := mux.Vars(r)["Path"]
	file, err := vcsrepo.ReadFile(r.Context(), vcs.CommitID(revSpec.CommitID), path)
	if err != nil {
		return err
	}

	return renderTemplate(w, "blob.html", &struct {
		AssetURL       string
		Repo           *sourcegraph.Repo
		RevSpec        sourcegraph.RepoRevSpec
		Path, Contents string
	}{
		AssetURL: assets.URL("/").String(),
		Repo:     repo,
		RevSpec:  revSpec,
		Path:     path,
		Contents: string(file),
	})
}

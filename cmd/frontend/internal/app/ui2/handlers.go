package ui2

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
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
	Context  jscontext.JSContext
	Route    string
	PageVars *pageVars
	AssetURL string

	// The fields below have zero values when not on a repo page.
	RepoShortName string
	Repo          *sourcegraph.Repo
	Rev           string                  // unresolved / user-specified revision (e.x.: "master")
	RevSpec       sourcegraph.RepoRevSpec // resolved SHA1 revision
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

// newCommon builds a *Common data structure, returning an error if one occurs.
//
// In the event of the repository being cloned, or having been renamed, the
// request is handled by newCommon and nil, nil is returned. Basic usage looks
// like:
//
// 	common, err := newCommon(w, r)
// 	if err != nil {
// 		return err
// 	}
// 	if common == nil {
// 		return nil // request was handled
// 	}
//
func newCommon(w http.ResponseWriter, r *http.Request, route string) (*Common, error) {
	common := &Common{
		Context:  jscontext.NewJSContextFromRequest(r),
		Route:    route,
		PageVars: &pageVars{},
		AssetURL: assets.URL("/").String(),
	}

	if _, ok := mux.Vars(r)["Repo"]; ok {
		// Common repo pages (blob, tree, etc).
		common.Rev = mux.Vars(r)["Rev"]
		var err error
		common.Repo, common.RevSpec, err = handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
		if err != nil {
			if e, ok := err.(*handlerutil.URLMovedError); ok {
				// The repository has been renamed, e.g. "github.com/docker/docker"
				// was renamed to "github.com/moby/moby" -> redirect the user now.
				http.Redirect(w, r, e.NewURL, http.StatusMovedPermanently)
				return nil, nil
			}
			if e, ok := err.(vcs.RepoNotExistError); ok && e.CloneInProgress {
				// Repo is cloning.
				//
				// TODO(slimsag): Move to template, auto-refresh page upon
				// completion, etc.
				fmt.Fprintf(w, "<html><h3>Repository is cloning...<h3/><p>(please refresh in a few seconds)</p></html>")
				return nil, nil
			}
			return nil, err
		}
		common.PageVars.ResolvedRev = common.RevSpec.CommitID
		common.RepoShortName = repoShortName(common.Repo.URI)
	}
	return common, nil
}

func serveHome(w http.ResponseWriter, r *http.Request) error {
	if !envvar.DeploymentOnPrem() && !actor.FromContext(r.Context()).IsAuthenticated() {
		// The user is not signed in and we are not on-prem so we are going to
		// redirect to about.sourcegraph.com.
		u, err := url.Parse(aboutRedirectScheme + "://" + aboutRedirectHost)
		if err != nil {
			return err
		}
		q := url.Values{}
		if r.Host != "sourcegraph.com" {
			// This allows about.sourcegraph.com to properly redirect back to
			// dev or staging environment after sign in.
			q.Set("host", r.Host)
		}
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
		return nil
	}

	// Serve the signed-in homepage.
	common, err := newCommon(w, r, routeHome)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	return renderTemplate(w, "home.html", &struct {
		*Common
	}{
		Common: common,
	})
}

func serveSearch(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, routeSearch)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	return renderTemplate(w, "search.html", &struct {
		*Common
	}{
		Common: common,
	})
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
	return urlTo(routeName, "Repo", t.RepoURI, "Rev", t.Rev, "Path", path.Join(t.Dir, i.Name())).String()
}

// navbar is the data structure shared/navbar.html expects.
type navbar struct {
	RepoURL        string
	RepoName       string      // e.x. "gorilla / mux"
	PathComponents [][2]string // [URL, path component]
	ViewOnGitHub   string      // link to view on GitHub, optional
}

func newNavbar(repoURI, rev, fpath string, isDir bool) *navbar {
	n := &navbar{
		RepoURL:  urlTo(routeRepoOrMain, "Repo", repoURI, "Rev", rev).String(),
		RepoName: strings.Replace(repoShortName(repoURI), "/", " / ", -1),
	}
	if strings.HasPrefix(repoURI, "github.com/") {
		n.ViewOnGitHub = "https://" + repoURI
	}
	split := strings.Split(fpath, "/")
	for i, p := range split {
		if p == "" {
			continue
		}

		// Only the last path component can be a file.
		routeName := routeTree
		if i == len(split)-1 && !isDir {
			routeName = routeBlob
		}

		// Construct a URL to this path.
		fpath := path.Join("/", path.Join(split[:i+1]...))
		u := urlTo(routeName, "Repo", repoURI, "Rev", rev, "Path", fpath).String()
		n.PathComponents = append(n.PathComponents, [2]string{u, p})
	}
	return n
}

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, routeRepoOrMain)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
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
		Navbar   *navbar
	}{
		Common: common,
		TreeView: &treeView{
			RepoURI: common.Repo.URI,
			Rev:     common.Rev,
			Dir:     dir,
			Files:   files,
		},
		Navbar: newNavbar(common.Repo.URI, common.Rev, dir, true),
	})
}

func serveTree(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, routeTree)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
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
		Navbar   *navbar
	}{
		Common: common,
		TreeView: &treeView{
			RepoURI: common.Repo.URI,
			Rev:     common.Rev,
			Dir:     dir,
			Files:   files,
		},
		Navbar: newNavbar(common.Repo.URI, common.Rev, dir, true),
	})
}

// blobView is the data structure shared/blobview.html expects.
type blobView struct {
	Path, Name string
	File       string
}

func serveBlob(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, routeBlob)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
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
		Navbar   *navbar
	}{
		Common: common,
		BlobView: &blobView{
			Path: fp,
			Name: path.Base(fp),
			File: string(file),
		},
		Navbar: newNavbar(common.Repo.URI, common.Rev, fp, false),
	})
}

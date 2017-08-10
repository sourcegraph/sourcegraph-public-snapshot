package ui2

import (
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
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
	Rev      string // unresolved revision specifier of current page (on any repo page). e.g. A branch, empty string (default branch), commit hash, etc.
	CommitID string // absolute revision of current page (on any repo page).
}

type Common struct {
	Context  jscontext.JSContext
	Route    string
	PageVars *pageVars
	AssetURL string

	// The fields below have zero values when not on a repo page.
	RepoShortName, RepoShortNameSpaced string // "gorilla/mux" and "gorilla / mux"
	Repo                               *sourcegraph.Repo
	Rev                                string                  // unresolved / user-specified revision (e.x.: "@master")
	RevSpec                            sourcegraph.RepoRevSpec // resolved SHA1 revision
	OpenOnDesktop                      template.URL
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

func (c *Common) addOpenOnDesktop(fpath string) {
	c.OpenOnDesktop = template.URL("src-insiders:open?resource=repo://" + path.Join(c.Repo.URI, fpath))
}

// newCommon builds a *Common data structure, returning an error if one occurs.
//
// In the event of the repository being cloned, or having been renamed, the
// request is handled by newCommon and nil, nil is returned. Basic usage looks
// like:
//
// 	common, err := newCommon(w, r, routeName, serveError)
// 	if err != nil {
// 		return err
// 	}
// 	if common == nil {
// 		return nil // request was handled
// 	}
//
func newCommon(w http.ResponseWriter, r *http.Request, route string, serveError func(w http.ResponseWriter, r *http.Request, err error, statusCode int)) (*Common, error) {
	common := &Common{
		Context:  jscontext.NewJSContextFromRequest(r),
		Route:    route,
		PageVars: &pageVars{},
		AssetURL: assets.URL("/").String(),
	}

	if _, ok := mux.Vars(r)["Repo"]; ok {
		// Common repo pages (blob, tree, etc).
		var err error
		common.Repo, common.RevSpec, err = handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
		if err != nil {
			if e, ok := err.(*handlerutil.URLMovedError); ok {
				// The repository has been renamed, e.g. "github.com/docker/docker"
				// was renamed to "github.com/moby/moby" -> redirect the user now.
				http.Redirect(w, r, e.NewURL, http.StatusMovedPermanently)
				return nil, nil
			}
			if legacyerr.ErrCode(err) == legacyerr.NotFound {
				// Repo does not exist.
				serveError(w, r, err, http.StatusNotFound)
				return nil, nil
			}
			if e, ok := err.(vcs.RepoNotExistError); ok && e.CloneInProgress {
				// Repo is cloning.
				common.RepoShortName = repoShortName(common.Repo.URI)
				return nil, renderTemplate(w, "cloning.html", &struct {
					*Common
				}{
					Common: common,
				})
			}
			return nil, err
		}
		common.Rev = mux.Vars(r)["Rev"]
		common.PageVars.Rev = strings.TrimPrefix(common.Rev, "@")
		common.PageVars.CommitID = common.RevSpec.CommitID
		common.RepoShortName = repoShortName(common.Repo.URI)
		common.RepoShortNameSpaced = strings.Join(strings.Split(repoShortName(common.Repo.URI), "/"), " / ")
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
	common, err := newCommon(w, r, routeHome, serveError)
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
	common, err := newCommon(w, r, routeSearch, serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	shortQuery := r.URL.Query().Get("q")
	if len(shortQuery) > 8 {
		shortQuery = shortQuery[:8]
	}

	return renderTemplate(w, "search.html", &struct {
		*Common
		ShortQuery string
	}{
		Common:     common,
		ShortQuery: shortQuery,
	})
}

// navbar is the data structure shared/navbar.html expects.
type navbar struct {
	RepoURL        string
	RepoName       string      // e.x. "gorilla / mux"
	PathComponents [][2]string // [URL, path component]
	ViewOnGitHub   string      // link to view on GitHub, optional
}

func githubURL(repo *sourcegraph.Repo, rev, fpath string, isDir bool) string {
	revOrDefault := rev
	if revOrDefault == "" {
		revOrDefault = repo.DefaultBranch
		if revOrDefault == "" {
			revOrDefault = "master"
		}
	}
	if fpath == "" {
		fpath = "/"
	}
	switch {
	case fpath == "/" && rev == "": // repo root
		return "https://" + repo.URI
	case fpath == "/" && rev != "": // repo@rev root
		return "https://" + path.Join(repo.URI, "tree", rev)
	case fpath != "/" && !isDir: // blob / file
		return "https://" + path.Join(repo.URI, "blob", revOrDefault, fpath)
	default: // tree
		return "https://" + path.Join(repo.URI, "tree", revOrDefault, fpath)
	}
}

func newNavbar(repo *sourcegraph.Repo, rev, fpath string, isDir bool) *navbar {
	n := &navbar{
		RepoURL:  urlTo(routeRepoOrMain, "Repo", repo.URI, "Rev", rev).String(),
		RepoName: strings.Replace(repoShortName(repo.URI), "/", " / ", -1),
	}
	if strings.HasPrefix(repo.URI, "github.com/") {
		n.ViewOnGitHub = githubURL(repo, strings.TrimPrefix(rev, "@"), fpath, isDir)
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
		u := urlTo(routeName, "Repo", repo.URI, "Rev", rev, "Path", fpath).String()
		n.PathComponents = append(n.PathComponents, [2]string{u, p})
	}
	return n
}

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, routeRepoOrMain, serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	common.addOpenOnDesktop("/")

	return renderTemplate(w, "repo.html", &struct {
		*Common
		Navbar *navbar
	}{
		Common: common,
		Navbar: newNavbar(common.Repo, common.Rev, "/", true),
	})
}

func serveTree(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, routeTree, serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	fp := mux.Vars(r)["Path"]
	common.addOpenOnDesktop(fp)

	return renderTemplate(w, "tree.html", &struct {
		*Common
		Navbar   *navbar
		FileName string
	}{
		Common:   common,
		Navbar:   newNavbar(common.Repo, common.Rev, fp, true),
		FileName: path.Base(fp),
	})
}

// blobView is the data structure shared/blobview.html expects.
type blobView struct {
	Path, Name string
	File       string
}

func serveBlob(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, routeBlob, serveError)
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
	common.addOpenOnDesktop(fp)
	file, err := vcsrepo.ReadFile(r.Context(), vcs.CommitID(common.RevSpec.CommitID), fp)
	if err != nil {
		return err
	}

	return renderTemplate(w, "blob.html", &struct {
		*Common
		BlobView *blobView
		Navbar   *navbar
		FileName string
	}{
		Common: common,
		BlobView: &blobView{
			Path: fp,
			Name: path.Base(fp),
			File: string(file),
		},
		Navbar:   newNavbar(common.Repo, common.Rev, fp, false),
		FileName: path.Base(fp),
	})
}

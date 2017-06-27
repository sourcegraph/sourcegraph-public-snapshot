package ui

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/bundle"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/errorutil"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func init() {
	router.Get(routeJobs).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://boards.greenhouse.io/sourcegraph", http.StatusFound)
	}))

	// Redirect from old /land/ def landing URLs to new /info/ URLs
	router.Get(oldRouteDefLanding).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		infoURL, err := router.Get(routeDefLanding).URL(
			"Repo", vars["Repo"], "Path", vars["Path"], "Rev", vars["Rev"], "UnitType", vars["UnitType"], "Unit", vars["Unit"])
		if err != nil {
			repoURL, err := router.Get(routeRepo).URL("Repo", vars["Repo"], "Rev", vars["Rev"])
			if err != nil {
				// Last recourse is redirect to homepage
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			// Redirect to repo page if info page URL could not be constructed
			http.Redirect(w, r, repoURL.String(), http.StatusFound)
			return
		}
		// Redirect to /info/ page
		http.Redirect(w, r, infoURL.String(), http.StatusMovedPermanently)
	}))

	router.Get(routeBlob).Handler(traceutil.TraceRoute(handler(serveBlob)))
	router.Get(routeDefRedirectToDefLanding).Handler(traceutil.TraceRoute(http.HandlerFunc(serveDefRedirectToDefLanding)))
	router.Get(routeDefLanding).Handler(traceutil.TraceRoute(errorutil.Handler(serveDefLanding)))
	router.Get(routeRepo).Handler(traceutil.TraceRoute(handler(serveRepo)))
	router.Get(routeRepoLanding).Handler(traceutil.TraceRoute(errorutil.Handler(serveRepoLanding)))
	router.Get(routeTree).Handler(traceutil.TraceRoute(handler(serveTree)))
	router.Get(routeAboutSubdomain).Handler(traceutil.TraceRoute(http.HandlerFunc(redirectAboutSubdomain)))
	router.Get(routeAppPaths).Handler(traceutil.TraceRoute(errorutil.Handler(serveAny)))
	router.Get(routeHomePage).Handler(traceutil.TraceRoute(errorutil.Handler(serveHome)))
	router.PathPrefix("/").Methods("GET").Handler(traceutil.TraceRouteFallback("app.serve-any", errorutil.Handler(serveAny)))
	router.NotFoundHandler = traceutil.TraceRouteFallback("app.serve-any-404", errorutil.Handler(serveAny))
}

func Router() *mux.Router {
	return router
}

// handler wraps h, calling tmplExec with the HTTP equivalent error
// code of h's return value (or HTTP 200 if err == nil).
func handler(h func(w http.ResponseWriter, r *http.Request) (*meta, error)) http.Handler {
	return errorutil.Handler(func(w http.ResponseWriter, r *http.Request) error {
		m, err := h(w, r)
		if m == nil {
			m = &meta{}
			if err != nil {
				m.Title = http.StatusText(errcode.HTTP(err))
			}
		}
		if ee, ok := err.(*handlerutil.URLMovedError); ok {
			return handlerutil.RedirectToNewRepoURI(w, r, ee.NewURL)
		}
		errorcode := errcode.HTTP(err)
		if errorcode >= 500 {
			log15.Error("HTTP UI error", "status", errorcode, "err", err.Error())
		}
		return tmplExec(w, r, errorcode, *m)
	})
}

// These handlers return the proper HTTP status code but otherwise do
// not pass data to the JavaScript UI code.

func serveBlob(w http.ResponseWriter, r *http.Request) (*meta, error) {
	q := r.URL.Query()

	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}

	info, err := stat(r.Context(), repoRev, routevar.ToTreeEntry(mux.Vars(r)).Path)
	if err != nil && !(err.Error() == "file does not exist" && r.URL.Query().Get("tmpZapRef") != "") { // TODO proper error value
		return nil, err
	}
	if info != nil && !info.Mode().IsRegular() {
		return nil, &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("tree entry is not a file")}
	}

	var m *meta
	if info == nil && q.Get("tmpZapRef") != "" {
		m = treeOrBlobMeta("", repo)
	} else {
		m = treeOrBlobMeta(info.Name(), repo)
	}
	m.CanonicalURL = canonicalRepoURL(
		conf.AppURL,
		getRouteName(r),
		mux.Vars(r),
		q,
		repo.DefaultBranch,
		repoRev.CommitID,
	)
	return m, nil
}

// serveDefRedirectToDefLanding redirects from /REPO/refs/... and
// /REPO/def/... URLs to the def landing page. Those URLs used to
// point to JavaScript-backed pages in the UI for a refs list and code
// view, respectively, but now def URLs are only for SEO (and thus
// those URLs are only handled by this package).
func serveDefRedirectToDefLanding(w http.ResponseWriter, r *http.Request) {
	routeVars := mux.Vars(r)
	pairs := make([]string, 0, len(routeVars)*2)
	for k, v := range routeVars {
		if k == "dummy" { // only used for matching string "def" or "refs"
			continue
		}
		pairs = append(pairs, k, v)
	}
	u, err := router.Get(routeDefLanding).URL(pairs...)
	if err != nil {
		log15.Error("Def redirect URL construction failed.", "url", r.URL.String(), "routeVars", routeVars, "err", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
}

func serveRepo(w http.ResponseWriter, r *http.Request) (*meta, error) {
	rr := routevar.ToRepoRev(mux.Vars(r))
	if rr.Rev == "" {
		// Just fetch the repo. Even if the rev doesn't exist, we
		// still want to return HTTP 200 OK, because the repo might be
		// in the process of being cloned. In that case, the 200 OK
		// refers to the existence of the repo, not the rev, which is
		// desirable.
		repo, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
		if err != nil {
			return nil, err
		}
		m := repoMeta(repo)
		m.CanonicalURL = canonicalRepoURL(
			conf.AppURL,
			getRouteName(r),
			mux.Vars(r),
			r.URL.Query(),
			repo.DefaultBranch,
			"",
		)
		return m, nil
	}

	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}

	m := repoMeta(repo)
	m.CanonicalURL = canonicalRepoURL(
		conf.AppURL,
		getRouteName(r),
		mux.Vars(r),
		r.URL.Query(),
		repo.DefaultBranch,
		repoRev.CommitID,
	)
	return m, nil
}

func serveTree(w http.ResponseWriter, r *http.Request) (*meta, error) {
	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}

	info, err := stat(r.Context(), repoRev, routevar.ToTreeEntry(mux.Vars(r)).Path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsDir() {
		return nil, &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("tree entry is not a dir")}
	}

	m := treeOrBlobMeta(info.Name(), repo)
	m.CanonicalURL = canonicalRepoURL(
		conf.AppURL,
		getRouteName(r),
		mux.Vars(r),
		r.URL.Query(),
		repo.DefaultBranch,
		repoRev.CommitID,
	)
	return m, nil
}

func aboutBaseURL() string {
	return "https://about.sourcegraph.com/"
}

func redirectAboutSubdomain(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["Path"]
	http.Redirect(w, r, aboutBaseURL()+path, http.StatusTemporaryRedirect)
}

// serveHome served the home page at "/"
func serveHome(w http.ResponseWriter, r *http.Request) error {
	if !envvar.DeploymentOnPrem() && !actor.FromContext(r.Context()).IsAuthenticated() {
		// The user is not signed in and we are not on-prem so we are going to redirect to about.sourcegraph.com.
		u, err := url.Parse(aboutBaseURL())
		if err != nil {
			return err
		}
		query := url.Values{}
		if r.Host != "sourcegraph.com" {
			// This allows about.sourcegraph.com to properly redirect back to
			// dev or staging environment after sign in.
			query.Set("host", r.Host)
		}
		u.RawQuery = query.Encode()
		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
		return nil
	}
	return serveAny(w, r)
}

// serveAny is the fallback/catch-all route. It preloads nothing and
// returns a page that will merely bootstrap the JavaScript app.
func serveAny(w http.ResponseWriter, r *http.Request) error {
	return tmplExec(w, r, http.StatusOK, meta{Index: true, Follow: true})
}

// testWriteMetaJSON, if set, causes tmplExec to bypass the actual
// template rendering and just write the meta information to the
// response as JSON. It is used when we want to unit-test the handlers
// and don't want to have to set up the JS bundle to make
// bundle.RenderEntrypoint succeed.
var testWriteMetaJSON bool

func tmplExec(w http.ResponseWriter, r *http.Request, statusCode int, m meta) error {
	if testWriteMetaJSON {
		w.WriteHeader(statusCode)
		return json.NewEncoder(w).Encode(m)
	}

	data := &struct {
		tmpl.Common
		Meta meta
	}{
		Meta: m,
	}
	// Check the config to determine the entrypoint.
	rawConfig := r.URL.Query().Get("config")
	var config map[string]interface{}
	json.Unmarshal([]byte(rawConfig), &config)
	if config != nil {
		if config["standaloneWorkbench"] == true {
			return bundle.RenderEntrypoint(w, r, statusCode, nil, data, true)
		}
	}
	return bundle.RenderEntrypoint(w, r, statusCode, nil, data, false)
}

func getRouteName(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route != nil {
		return route.GetName()
	}
	return ""
}

func stat(ctx context.Context, repoRev sourcegraph.RepoRevSpec, path string) (os.FileInfo, error) {
	// Cap operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repoRev.Repo)
	if err != nil {
		return nil, err
	}

	commit := vcs.CommitID(repoRev.CommitID)
	return vcsrepo.Lstat(ctx, commit, path)
}

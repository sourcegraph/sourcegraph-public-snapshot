package ui

import (
	"errors"
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"context"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func init() {
	router.Get(routeJobs).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://boards.greenhouse.io/sourcegraph", http.StatusFound)
	}))
	router.Get(routeReposIndex).Handler(httptrace.TraceRoute(internal.Handler(serveRepoIndex)))
	router.Get(routeLangsIndex).Handler(httptrace.TraceRoute(internal.Handler(serveRepoIndex)))
	router.Get(routeBlob).Handler(httptrace.TraceRoute(handler(serveBlob)))
	router.Get(routeDefRedirectToDefLanding).Handler(httptrace.TraceRoute(http.HandlerFunc(serveDefRedirectToDefLanding)))
	router.Get(routeDefLanding).Handler(httptrace.TraceRoute(internal.Handler(serveDefLanding)))
	router.Get(routeRepo).Handler(httptrace.TraceRoute(handler(serveRepo)))
	router.Get(routeRepoLanding).Handler(httptrace.TraceRoute(internal.Handler(serveRepoLanding)))
	router.Get(routeTree).Handler(httptrace.TraceRoute(handler(serveTree)))
	router.Get(routeTopLevel).Handler(httptrace.TraceRoute(internal.Handler(serveAny)))
	router.PathPrefix("/").Methods("GET").Handler(httptrace.TraceRoute(internal.Handler(serveAny)))
	router.NotFoundHandler = internal.Handler(serveAny)
}

func Router() *mux.Router {
	return router
}

// handler wraps h, calling tmplExec with the HTTP equivalent error
// code of h's return value (or HTTP 200 if err == nil).
func handler(h func(w http.ResponseWriter, r *http.Request) (*meta, error)) http.Handler {
	return internal.Handler(func(w http.ResponseWriter, r *http.Request) error {
		m, err := h(w, r)
		if m == nil {
			m = &meta{}
			if err != nil {
				m.Title = http.StatusText(errcode.HTTP(err))
			}
		}
		return tmplExec(w, r, errcode.HTTP(err), *m)
	})
}

// These handlers return the proper HTTP status code but otherwise do
// not pass data to the JavaScript UI code.

func repoTreeGet(ctx context.Context, routeVars map[string]string) (*sourcegraph.TreeEntry, *sourcegraph.Repo, *sourcegraph.RepoRevSpec, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)

	repo, repoRev, err := handlerutil.GetRepoAndRev(ctx, routeVars)
	if err != nil {
		return nil, nil, nil, err
	}

	entry := routevar.ToTreeEntry(routeVars)
	e, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{RepoRev: repoRev, Path: entry.Path},
		Opt:   nil,
	})
	return e, repo, &repoRev, err
}

func serveBlob(w http.ResponseWriter, r *http.Request) (*meta, error) {
	entry, repo, repoRev, err := repoTreeGet(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}
	if entry.Type != sourcegraph.FileEntry {
		return nil, &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("tree entry is not a file")}
	}

	m := treeOrBlobMeta(entry.Name, repo)
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
	entry, repo, repoRev, err := repoTreeGet(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}
	if entry.Type != sourcegraph.DirEntry {
		return nil, &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("tree entry is not a dir")}
	}

	m := treeOrBlobMeta(entry.Name, repo)
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

// serveAny is the fallback/catch-all route. It preloads nothing and
// returns a page that will merely bootstrap the JavaScript app.
func serveAny(w http.ResponseWriter, r *http.Request) error {
	return tmplExec(w, r, http.StatusOK, meta{Index: true, Follow: true})
}

func tmplExec(w http.ResponseWriter, r *http.Request, statusCode int, m meta) error {
	return tmpl.Exec(r, w, "ui.html", statusCode, nil, &struct {
		tmpl.Common
		Meta meta
	}{
		Meta: m,
	})
}

func getRouteName(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route != nil {
		return route.GetName()
	}
	return ""
}

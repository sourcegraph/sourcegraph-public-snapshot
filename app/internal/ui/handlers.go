package ui

import (
	"errors"
	"net/http"

	"context"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func init() {
	router.Get(routeBlob).Handler(httptrace.TraceRoute(handler(serveBlob)))
	router.Get(routeBuild).Handler(httptrace.TraceRoute(handler(serveBuild)))
	router.Get(routeDef).Handler(httptrace.TraceRoute(handler(serveDef)))
	router.Get(routeDefInfo).Handler(httptrace.TraceRoute(handler(serveDefInfo)))
	router.Get(routeRepo).Handler(httptrace.TraceRoute(handler(serveRepo)))
	router.Get(routeRepoBuilds).Handler(httptrace.TraceRoute(handler(serveRepoBuilds)))
	router.Get(routeTree).Handler(httptrace.TraceRoute(handler(serveTree)))
	router.Get(routeTopLevel).Handler(httptrace.TraceRoute(internal.Handler(serveAny)))
	router.PathPrefix("/").Methods("GET").Handler(httptrace.TraceRoute(internal.Handler(serveAny)))
	router.NotFoundHandler = internal.Handler(serveAny)

	// Attach to app handler's catch-all UI route. This is better than
	// adding the UI routes to the app router directly because it
	// keeps the two routers separate.
	internal.Handlers[approuter.UI] = router
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
		conf.AppURL(r.Context()),
		getRouteName(r),
		mux.Vars(r),
		r.URL.Query(),
		repo.DefaultBranch,
		repoRev.CommitID,
	)
	return m, nil
}

func serveBuild(w http.ResponseWriter, r *http.Request) (*meta, error) {
	_, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}

	// NOTE: We don't actually try to fetch the build here, but that's
	// OK. The frontend JS will notice and display the error if the
	// build doesn't exist or is inaccessible. It's not super
	// important to return proper 404s for builds, relative to other
	// URLs that are linked more often.

	return nil, nil
}

func serveDef(w http.ResponseWriter, r *http.Request) (*meta, error) {
	return serveDefCommon(w, r, false)
}

func serveDefInfo(w http.ResponseWriter, r *http.Request) (*meta, error) {
	return serveDefCommon(w, r, true)
}

func serveDefCommon(w http.ResponseWriter, r *http.Request, isDefInfo bool) (*meta, error) {
	def, repo, err := handlerutil.GetDefCommon(r.Context(), mux.Vars(r), &sourcegraph.DefGetOptions{Doc: true})
	if err != nil {
		return nil, err
	}
	m := defMeta(def, repo.URI, !isDefInfo)

	if isDefInfo {
		// DefInfo canonical URL is DefInfo.
		m.CanonicalURL = canonicalRepoURL(
			conf.AppURL(r.Context()),
			getRouteName(r),
			mux.Vars(r),
			r.URL.Query(),
			repo.DefaultBranch,
			def.CommitID,
		)
	} else {
		// Def canonical URL is the blob page. We don't want Googlebot
		// thinking we have tons of repetitive pages (each local var
		// on a blob page, for example), so let's tell it that all Def
		// pages are actually canonically the blob.
		m.CanonicalURL = canonicalRepoURL(
			conf.AppURL(r.Context()),
			routeBlob,
			routevar.TreeEntryRouteVars(routevar.TreeEntry{
				RepoRev: routevar.ToRepoRev(mux.Vars(r)),
				Path:    "/" + def.File,
			}),
			r.URL.Query(),
			repo.DefaultBranch,
			def.CommitID,
		)
	}

	// Don't noindex pages with a canonical URL. See
	// https://www.seroundtable.com/archives/020151.html.
	canonRev := isCanonicalRev(mux.Vars(r), repo.DefaultBranch)
	m.Index = allowRobots(repo) && shouldIndexDef(def) && canonRev && isDefInfo // DefInfo is a better landing page than Def.

	return m, nil
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
			conf.AppURL(r.Context()),
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
		conf.AppURL(r.Context()),
		getRouteName(r),
		mux.Vars(r),
		r.URL.Query(),
		repo.DefaultBranch,
		repoRev.CommitID,
	)
	return m, nil
}

func serveRepoBuilds(w http.ResponseWriter, r *http.Request) (*meta, error) {
	_, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}
	return nil, nil
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
		conf.AppURL(r.Context()),
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

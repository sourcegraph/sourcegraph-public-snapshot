package ui

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func init() {
	router.Get(routeBlob).Handler(handler(serveBlob))
	router.Get(routeBuild).Handler(handler(serveBuild))
	router.Get(routeDef).Handler(handler(serveDef))
	router.Get(routeDefInfo).Handler(handler(serveDefInfo))
	router.Get(routeRepo).Handler(handler(serveRepo))
	router.Get(routeRepoBuilds).Handler(handler(serveRepoBuilds))
	router.Get(routeTree).Handler(handler(serveTree))
	router.Get(routeTopLevel).Handler(internal.Handler(serveAny))
	router.PathPrefix("/").Methods("GET").Handler(internal.Handler(serveAny))
	router.NotFoundHandler = internal.Handler(serveAny)

	// Attach to app handler's catch-all UI route. This is better than
	// adding the UI routes to the app router directly because it
	// keeps the two routers separate.
	internal.Handlers[approuter.UI] = router
}

// handler wraps h, calling tmplExec with the HTTP equivalent error
// code of h's return value (or HTTP 200 if err == nil).
func handler(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return internal.Handler(func(w http.ResponseWriter, r *http.Request) error {
		err := h(w, r)
		return tmplExec(w, r, errcode.HTTP(err))
	})
}

// These handlers return the proper HTTP status code but otherwise do
// not pass data to the JavaScript UI code.

func repoTreeGet(ctx context.Context, routeVars map[string]string) (*sourcegraph.TreeEntry, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)

	_, repoRev, err := handlerutil.GetRepoAndRev(ctx, routeVars)
	if err != nil {
		return nil, err
	}

	entry := routevar.ToTreeEntry(routeVars)
	return cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{RepoRev: repoRev, Path: entry.Path},
		Opt:   nil,
	})
}

func serveBlob(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)
	entry, err := repoTreeGet(ctx, mux.Vars(r))
	if err != nil {
		return err
	}
	if entry.Type != sourcegraph.FileEntry {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("tree entry is not a file")}
	}
	return nil
}

func serveBuild(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)
	_, err := handlerutil.GetRepo(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	// NOTE: We don't actually try to fetch the build here, but that's
	// OK. The frontend JS will notice and display the error if the
	// build doesn't exist or is inaccessible. It's not super
	// important to return proper 404s for builds, relative to other
	// URLs that are linked more often.

	return nil
}

func serveDef(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)
	_, _, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), &sourcegraph.DefGetOptions{Doc: true})
	if err != nil {
		return err
	}
	return nil
}

func serveDefInfo(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)
	_, _, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), &sourcegraph.DefGetOptions{Doc: true})
	if err != nil {
		return err
	}
	return nil
}

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	rr := routevar.ToRepoRev(mux.Vars(r))
	if rr.Rev == "" {
		// Just fetch the repo. Even if the rev doesn't exist, we
		// still want to return HTTP 200 OK, because the repo might be
		// in the process of being cloned. In that case, the 200 OK
		// refers to the existence of the repo, not the rev, which is
		// desirable.
		_, err := handlerutil.GetRepo(ctx, mux.Vars(r))
		return err
	}

	_, _, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
	return err
}

func serveRepoBuilds(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)
	_, err := handlerutil.GetRepo(ctx, mux.Vars(r))
	if err != nil {
		return err
	}
	return nil
}

func serveTree(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)
	entry, err := repoTreeGet(ctx, mux.Vars(r))
	if err != nil {
		return err
	}
	if entry.Type != sourcegraph.DirEntry {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("tree entry is not a dir")}
	}
	return nil
}

// serveAny is the fallback/catch-all route. It preloads nothing and
// returns a page that will merely bootstrap the JavaScript app.
func serveAny(w http.ResponseWriter, r *http.Request) error {
	return tmplExec(w, r, http.StatusOK)
}

func tmplExec(w http.ResponseWriter, r *http.Request, statusCode int) error {
	return tmpl.Exec(r, w, "ui.html", statusCode, nil, &struct {
		tmpl.Common
		Stores *json.RawMessage
	}{})
}

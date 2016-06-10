package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveRepoBuilds(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.BuildListOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	repo, err := handlerutil.GetRepoID(ctx, mux.Vars(r))
	if err != nil {
		return err
	}
	opt.Repo = repo

	builds, err := cl.Builds.List(ctx, &opt)
	if err != nil {
		return err
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	return writeJSON(w, builds)
}

func serveRepoBuild(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	buildSpec, err := getBuildSpec(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	build, err := cl.Builds.Get(ctx, buildSpec)
	if err != nil {
		return err
	}

	return writeJSON(w, build)
}

func serveRepoBuildsCreate(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var op sourcegraph.BuildsCreateOp
	err := json.NewDecoder(r.Body).Decode(&op)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	// Don't let the user specify the config
	op.Config = sourcegraph.BuildConfig{
		Queue: true,
		// Builds triggered from the UI have a high priority
		Priority: 100,
	}

	repo, err := handlerutil.GetRepoID(ctx, mux.Vars(r))
	if err != nil {
		return err
	}
	op.Repo = repo

	build, err := cl.Builds.Create(ctx, &op)
	if err != nil {
		return err
	}
	return writeJSON(w, build)
}

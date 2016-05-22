package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveRepoBuilds(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	repoSpec, err := getRepoSpec(r)
	if err != nil {
		return err
	}

	var opt sourcegraph.BuildListOptions
	err = schemaDecoder.Decode(&opt, r.URL.Query())
	opt.Repo = repoSpec.URI
	if err != nil {
		return err
	}

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

	buildSpec, err := getBuildSpec(r)
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

	_, repoSpec, err := handlerutil.GetRepo(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	op.Repo = repoSpec
	build, err := cl.Builds.Create(ctx, &op)
	if err != nil {
		return err
	}
	return writeJSON(w, build)
}

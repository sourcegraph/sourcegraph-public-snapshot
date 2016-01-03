package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoBuild(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	_, repoRevSpec, _, err := handlerutil.GetRepoAndRev(r, s.Repos)
	if err != nil {
		return err
	}

	build, err := s.Builds.GetRepoBuild(ctx, &repoRevSpec)
	if err != nil {
		return err
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	return writeJSON(w, build)
}

func serveRepoBuildsCreate(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var op sourcegraph.BuildsCreateOp
	err := json.NewDecoder(r.Body).Decode(&op)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	_, repoSpec, err := handlerutil.GetRepo(r, s.Repos)
	if err != nil {
		return err
	}

	op.Repo = repoSpec
	build, err := s.Builds.Create(ctx, &op)
	if err != nil {
		return err
	}
	return writeJSON(w, build)
}

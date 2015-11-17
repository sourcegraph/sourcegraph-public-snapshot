package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoBuildInfo(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var opt sourcegraph.BuildsGetRepoBuildInfoOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	_, repoRevSpec, _, err := handlerutil.GetRepoAndRev(r, s.Repos)
	if err != nil {
		return err
	}

	buildInfo, err := s.Builds.GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: repoRevSpec, Opt: &opt})
	if err != nil {
		return err
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	return writeJSON(w, buildInfo)
}

func serveRepoBuildsCreate(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var opt sourcegraph.BuildCreateOptions
	err := json.NewDecoder(r.Body).Decode(&opt)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	_, repoRevSpec, _, err := handlerutil.GetRepoAndRev(r, s.Repos)
	if err != nil {
		return err
	}

	build, err := s.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{RepoRev: repoRevSpec, Opt: &opt})
	if err != nil {
		return err
	}
	return writeJSON(w, build)
}

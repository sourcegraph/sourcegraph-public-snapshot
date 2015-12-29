package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveBuild(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	buildSpec, err := getBuildSpec(r)
	if err != nil {
		return err
	}

	build, err := s.Builds.Get(ctx, buildSpec)
	if err != nil {
		return err
	}

	return writeJSON(w, build)
}

func serveBuildTasks(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	buildSpec, err := getBuildSpec(r)
	if err != nil {
		return err
	}

	var opt sourcegraph.BuildTaskListOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	tasks, err := s.Builds.ListBuildTasks(ctx, &sourcegraph.BuildsListBuildTasksOp{
		Build: *buildSpec,
		Opt:   &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, tasks)
}

func serveBuilds(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	// Delete the "_" cache-busting query param that jQuery adds when
	// the "cache" $.ajaxSettings entry is false. We use this to
	// immediately return the new build when the user triggers a new
	// build to be created. Otherwise the old build (or lack thereof)
	// would be cached for our default cache time.
	q := r.URL.Query()
	delete(q, "_")
	r.URL.RawQuery = q.Encode()

	var opt sourcegraph.BuildListOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	builds, err := s.Builds.List(ctx, &opt)
	if err != nil {
		return err
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	return writeJSON(w, builds)
}

func getBuildSpec(r *http.Request) (*sourcegraph.BuildSpec, error) {
	v := mux.Vars(r)
	commit, repo := v["CommitID"], v["Repo"]
	attempt, err := strconv.ParseInt(v["Attempt"], 10, 32)
	if commit == "" || repo == "" || err != nil {
		return nil, &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	return &sourcegraph.BuildSpec{
		Attempt:  uint32(attempt),
		CommitID: commit,
		Repo:     sourcegraph.RepoSpec{URI: repo},
	}, nil
}

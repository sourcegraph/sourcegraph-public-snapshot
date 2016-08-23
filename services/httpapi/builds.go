package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveBuildTasks(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	buildSpec, err := getBuildSpec(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	var opt sourcegraph.BuildTaskListOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	tasks, err := cl.Builds.ListBuildTasks(r.Context(), &sourcegraph.BuildsListBuildTasksOp{
		Build: *buildSpec,
		Opt:   &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, tasks)
}

func serveBuilds(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	var tmp struct {
		Repo repoIDOrPath
		sourcegraph.BuildListOptions
	}
	if err := schemaDecoder.Decode(&tmp, r.URL.Query()); err != nil {
		return err
	}
	opt := tmp.BuildListOptions
	if tmp.Repo != "" {
		var err error
		opt.Repo, err = getRepoID(r.Context(), tmp.Repo)
		if err != nil {
			return err
		}
	}

	builds, err := cl.Builds.List(r.Context(), &opt)
	if err != nil {
		return err
	}

	// Add a RepoPath field to each build because most API clients
	// would need that information.
	builds2, err := addBuildRepoPaths(r.Context(), builds.Builds)
	if err != nil {
		return err
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	return writeJSON(w, struct{ Builds []buildWithRepoPath }{builds2})
}

func serveBuildTaskLog(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	var opt sourcegraph.BuildGetLogOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	taskSpec, err := getBuildTaskSpec(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	entries, err := cl.Builds.GetTaskLog(r.Context(), &sourcegraph.BuildsGetTaskLogOp{Task: *taskSpec, Opt: &opt})
	if err != nil {
		return err
	}

	return writePlainLogEntries(w, entries)
}

func getBuildSpec(ctx context.Context, vars map[string]string) (*sourcegraph.BuildSpec, error) {
	repo, err := handlerutil.GetRepoID(ctx, vars)
	if err != nil {
		return nil, err
	}
	build, err := strconv.ParseUint(vars["Build"], 10, 64)
	if err != nil {
		return nil, &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	return &sourcegraph.BuildSpec{Repo: repo, ID: build}, nil
}

func getBuildTaskSpec(ctx context.Context, vars map[string]string) (*sourcegraph.TaskSpec, error) {
	buildSpec, err := getBuildSpec(ctx, vars)
	if err != nil {
		return nil, err
	}

	taskID, err := strconv.ParseUint(vars["Task"], 10, 64)
	if err != nil {
		return nil, &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	return &sourcegraph.TaskSpec{Build: *buildSpec, ID: taskID}, nil
}

type buildWithRepoPath struct {
	*sourcegraph.Build
	RepoPath string
}

func addBuildRepoPaths(ctx context.Context, builds []*sourcegraph.Build) ([]buildWithRepoPath, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	builds2 := make([]buildWithRepoPath, len(builds))
	memo := map[int32]string{}
	for i, b := range builds {
		builds2[i].Build = b
		if repoPath, present := memo[b.Repo]; present {
			builds2[i].RepoPath = repoPath
		} else {
			repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{ID: b.Repo})
			if err != nil {
				return nil, err
			}
			builds2[i].RepoPath = repo.URI
		}
	}
	return builds2, nil
}

func writePlainLogEntries(w http.ResponseWriter, entries *sourcegraph.LogEntries) error {
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	if entries.MaxID != "" {
		w.Header().Add("x-sourcegraph-log-max-id", entries.MaxID)
	}

	printFunc := fmt.Fprintln
	for i, e := range entries.Entries {
		// Don't print an artificial trailing newline.
		if i == len(entries.Entries)-1 {
			printFunc = fmt.Fprint
		}

		if _, err := printFunc(w, e); err != nil {
			return err
		}
	}
	return nil
}

package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func serveBuildTasks(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	buildSpec, err := getBuildSpec(r)
	if err != nil {
		return err
	}

	var opt sourcegraph.BuildTaskListOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	tasks, err := cl.Builds.ListBuildTasks(ctx, &sourcegraph.BuildsListBuildTasksOp{
		Build: *buildSpec,
		Opt:   &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, tasks)
}

func serveBuilds(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.BuildListOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
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

func serveBuildTaskLog(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.BuildGetLogOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	taskSpec, err := getBuildTaskSpec(r)
	if err != nil {
		return err
	}

	entries, err := cl.Builds.GetTaskLog(ctx, &sourcegraph.BuildsGetTaskLogOp{Task: taskSpec, Opt: &opt})
	if err != nil {
		return err
	}

	return writePlainLogEntries(w, entries)
}

func getBuildSpec(r *http.Request) (*sourcegraph.BuildSpec, error) {
	v := mux.Vars(r)
	build, err := strconv.ParseUint(v["Build"], 10, 64)
	if err != nil {
		return nil, &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	return &sourcegraph.BuildSpec{
		Repo: routevar.ToRepo(mux.Vars(r)),
		ID:   build,
	}, nil
}

func getBuildTaskSpec(r *http.Request) (sourcegraph.TaskSpec, error) {
	buildSpec, err := getBuildSpec(r)
	if err != nil {
		return sourcegraph.TaskSpec{}, err
	}

	v := mux.Vars(r)
	taskID, err := strconv.ParseUint(v["Task"], 10, 64)
	if err != nil {
		return sourcegraph.TaskSpec{}, &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	return sourcegraph.TaskSpec{Build: *buildSpec, ID: taskID}, nil
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

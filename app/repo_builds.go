package app

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"

	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoBuilds(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.BuildListOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r)
	if err != nil {
		return err
	}

	// Set defaults for Builds.List call options.
	buildslistOpt := defaultBuildListOptions(opt)
	buildslistOpt.Repo = rc.Repo.URI
	builds, err := apiclient.Builds.List(ctx, &buildslistOpt)
	if err != nil {
		return err
	}

	pg, err := paginatePrevNext(opt, builds.StreamResponse)
	if err != nil {
		return err
	}

	buildsAndCommits, err := fetchCommitsForBuilds(ctx, builds.Builds)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/builds.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		BuildsAndCommits []buildAndCommit
		PageLinks        []pageLink

		tmpl.Common
	}{
		RepoCommon:       *rc,
		BuildsAndCommits: buildsAndCommits,
		PageLinks:        pg,
	})
}

func serveRepoBuildsCreate(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r)
	if err != nil {
		return err
	}

	// Default options.
	op := sourcegraph.BuildsCreateOp{
		Config: sourcegraph.BuildConfig{
			Queue: true,
		},
	}
	if err := r.ParseForm(); err != nil {
		return err
	}
	if err := schemautil.Decode(&op, r.PostForm); err != nil {
		return err
	}
	op.Repo = rc.Repo.RepoSpec()

	build, err := apiclient.Builds.Create(ctx, &op)
	if err != nil {
		return err
	}

	http.Redirect(w, r, router.Rel.URLToRepoBuild(rc.Repo.URI, build.ID).String(), http.StatusSeeOther)
	return nil
}

func serveRepoBuild(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r)
	if err != nil {
		return err
	}

	build, _, err := getRepoBuild(r, rc.Repo)
	if err != nil {
		return err
	}

	commit0, err := apiclient.Repos.GetCommit(ctx, &sourcegraph.RepoRevSpec{RepoSpec: rc.Repo.RepoSpec(), Rev: build.CommitID, CommitID: build.CommitID})
	if handlerutil.IsRepoNoVCSDataError(err) {
		// Commit remains nil, will not be displayed in template.
	} else if err != nil {
		return err
	}
	var commit *payloads.AugmentedCommit
	if commit0 != nil {
		var commits []*payloads.AugmentedCommit
		commits, err = handlerutil.AugmentCommits(r, rc.Repo.URI, []*vcs.Commit{commit0})
		if err != nil {
			return err
		}
		commit = commits[0]
	}

	return tmpl.Exec(r, w, "repo/build.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Build  *sourcegraph.Build
		Commit *payloads.AugmentedCommit

		tmpl.Common
	}{
		RepoCommon: *rc,
		Build:      build,
		Commit:     commit,

		Common: tmpl.Common{FullWidth: true},
	})
}

func serveRepoBuildUpdate(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	rc, err := handlerutil.GetRepoCommon(r)
	if err != nil {
		return err
	}

	_, buildSpec, err := getRepoBuild(r, rc.Repo)
	if err != nil {
		return err
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	var buildUpdate sourcegraph.BuildUpdate
	if err := schemautil.Decode(&buildUpdate, r.PostForm); err != nil {
		return err
	}

	if _, err := apiclient.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{Build: buildSpec, Info: buildUpdate}); err != nil {
		return err
	}

	http.Redirect(w, r, router.Rel.URLToRepoBuild(rc.Repo.URI, buildSpec.ID).String(), http.StatusSeeOther)
	return nil
}

func serveRepoBuildTaskLog(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	var opt sourcegraph.BuildGetLogOptions
	if err := schemautil.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	rc, err := handlerutil.GetRepoCommon(r)
	if err != nil {
		return err
	}

	_, _, err = getRepoBuild(r, rc.Repo)
	if err != nil {
		return err
	}

	taskSpec, err := getBuildTaskSpec(r)
	if err != nil {
		return err
	}

	entries, err := apiclient.Builds.GetTaskLog(ctx, &sourcegraph.BuildsGetTaskLogOp{Task: taskSpec, Opt: &opt})
	if err != nil {
		return err
	}

	return writePlainLogEntries(w, entries)
}

func getBuildSpec(r *http.Request) (sourcegraph.BuildSpec, error) {
	v := mux.Vars(r)
	repo := v["Repo"]
	buildID, err := strconv.ParseUint(v["Build"], 10, 64)
	if repo == "" || err != nil {
		return sourcegraph.BuildSpec{}, &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	return sourcegraph.BuildSpec{
		Repo: sourcegraph.RepoSpec{URI: repo},
		ID:   buildID,
	}, nil
}

func getRepoBuild(r *http.Request, repo *sourcegraph.Repo) (*sourcegraph.Build, sourcegraph.BuildSpec, error) {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	buildSpec, err := getBuildSpec(r)
	if err != nil {
		return nil, sourcegraph.BuildSpec{}, err
	}

	build, err := apiclient.Builds.Get(ctx, &buildSpec)
	if err != nil {
		return nil, buildSpec, err
	}

	if repo.URI != build.Repo {
		return nil, buildSpec, &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("no such build for this repository")}
	}

	return build, buildSpec, nil
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
	return sourcegraph.TaskSpec{Build: buildSpec, ID: taskID}, nil
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

// buildStatus returns a textual status description for the build.
func buildStatus(b *sourcegraph.Build) string {
	if b.Failure {
		return "Failed"
	}
	if b.Success {
		return "Succeeded"
	}
	if b.StartedAt != nil && b.EndedAt == nil {
		return "In progress"
	}
	return "Queued"
}

// buildClass returns the CSS class for the build.
func buildClass(b *sourcegraph.Build) string {
	switch buildStatus(b) {
	case "Failed":
		return "danger"
	case "Succeeded":
		return "success"
	case "In progress":
		return "info"
	}
	return "default"
}

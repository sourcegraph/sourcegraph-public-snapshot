package app

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/buildutil"
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

	rc, err := handlerutil.GetRepoCommon(r, &handlerutil.GetRepoCommonOpt{AllowNonEnabledRepos: true})
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

	pg, err := paginatePrevNext(opt, builds.ListResponse)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/builds.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Builds    []*sourcegraph.Build
		PageLinks []pageLink

		tmpl.Common
	}{
		RepoCommon: *rc,
		Builds:     builds.Builds,
		PageLinks:  pg,
	})
}

func serveRepoNoBuildError(w http.ResponseWriter, r *http.Request, err *handlerutil.NoBuildError) error {
	return tmpl.Exec(r, w, "repo/no_build.html", http.StatusOK, nil, &struct {
		*handlerutil.NoBuildError
		tmpl.Common
	}{NoBuildError: err})
}

func serveRepoBuildsCreate(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r, nil)
	if err != nil {
		return err
	}

	// Default options.
	form := sourcegraph.BuildCreateOptions{
		BuildConfig: sourcegraph.BuildConfig{
			Import:   true,
			Queue:    true,
			Priority: int32(buildutil.DefaultPriority(rc.Repo.Private, buildutil.Manual)),
		},
		Force: true,
	}
	if err := r.ParseForm(); err != nil {
		return err
	}

	commitID := r.PostForm.Get("CommitID")
	// Resolve revspec to full commit ID. (This allows them to specify
	// a revspec in the CommitID field and have it resolved here.)
	if commitID == "" {
		commitID = rc.Repo.DefaultBranch
	}

	commit, err := apiclient.Repos.GetCommit(ctx, &sourcegraph.RepoRevSpec{RepoSpec: rc.Repo.RepoSpec(), Rev: commitID})
	if err != nil {
		return err
	}
	commitID = string(commit.ID)
	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: rc.Repo.RepoSpec(), Rev: commitID, CommitID: commitID}
	delete(r.PostForm, "CommitID")

	if err := schemautil.Decode(&form, r.PostForm); err != nil {
		return err
	}

	build, err := apiclient.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{RepoRev: repoRevSpec, Opt: &form})
	if err != nil {
		return err
	}

	http.Redirect(w, r, router.Rel.URLToRepoBuild(rc.Repo.URI, build.Spec().CommitID, build.Spec().Attempt).String(), http.StatusSeeOther)
	return nil
}

func serveRepoBuild(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r, &handlerutil.GetRepoCommonOpt{AllowNonEnabledRepos: true})
	if err != nil {
		return err
	}

	build, buildSpec, err := getRepoBuild(r, rc.Repo)
	if err != nil {
		return err
	}

	tasks, err := apiclient.Builds.ListBuildTasks(
		ctx,
		&sourcegraph.BuildsListBuildTasksOp{
			Build: buildSpec,
			Opt: &sourcegraph.BuildTaskListOptions{
				ListOptions: sourcegraph.ListOptions{PerPage: 99999},
			},
		},
	)
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
		commit, err = handlerutil.AugmentCommit(r, rc.Repo.URI, commit0)
		if err != nil {
			return err
		}
	}

	return tmpl.Exec(r, w, "repo/build.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Build  *sourcegraph.Build
		Commit *payloads.AugmentedCommit
		Tasks  []*sourcegraph.BuildTask

		tmpl.Common
	}{
		RepoCommon: *rc,
		Build:      build,
		Commit:     commit,
		Tasks:      tasks.BuildTasks,
	})
}

func serveRepoBuildUpdate(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	rc, err := handlerutil.GetRepoCommon(r, nil)
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

	http.Redirect(w, r, router.Rel.URLToRepoBuild(rc.Repo.URI, buildSpec.CommitID, buildSpec.Attempt).String(), http.StatusSeeOther)
	return nil
}

func serveRepoBuildLog(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	var opt sourcegraph.BuildGetLogOptions
	if err := schemautil.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	rc, err := handlerutil.GetRepoCommon(r, &handlerutil.GetRepoCommonOpt{AllowNonEnabledRepos: true})
	if err != nil {
		return err
	}

	_, buildSpec, err := getRepoBuild(r, rc.Repo)
	if err != nil {
		return err
	}

	entries, err := apiclient.Builds.GetLog(ctx, &sourcegraph.BuildsGetLogOp{Build: buildSpec, Opt: &opt})
	if err != nil {
		return err
	}

	return writePlainLogEntries(w, entries)
}

func serveRepoBuildTaskLog(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	var opt sourcegraph.BuildGetLogOptions
	if err := schemautil.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	rc, err := handlerutil.GetRepoCommon(r, &handlerutil.GetRepoCommonOpt{AllowNonEnabledRepos: true})
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
	commit, repo := v["CommitID"], v["Repo"]
	attempt, err := strconv.ParseUint(v["Attempt"], 10, 32)
	if commit == "" || repo == "" || err != nil {
		return sourcegraph.BuildSpec{}, &handlerutil.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	return sourcegraph.BuildSpec{
		Attempt:  uint32(attempt),
		CommitID: commit,
		Repo:     sourcegraph.RepoSpec{URI: repo},
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
		return nil, buildSpec, &handlerutil.HTTPErr{Status: http.StatusNotFound, Err: errors.New("no such build for this repository")}
	}

	return build, buildSpec, nil
}

func getBuildTaskSpec(r *http.Request) (sourcegraph.TaskSpec, error) {
	buildSpec, err := getBuildSpec(r)
	if err != nil {
		return sourcegraph.TaskSpec{}, err
	}

	v := mux.Vars(r)
	taskID, err := strconv.ParseInt(v["TaskID"], 10, 64)
	if err != nil {
		return sourcegraph.TaskSpec{}, &handlerutil.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	return sourcegraph.TaskSpec{BuildSpec: buildSpec, TaskID: taskID}, nil
}

func writePlainLogEntries(w http.ResponseWriter, entries *sourcegraph.LogEntries) error {
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	w.Header().Add("x-sourcegraph-log-max-id", entries.MaxID)

	for _, e := range entries.Entries {
		if _, err := fmt.Fprintln(w, e); err != nil {
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

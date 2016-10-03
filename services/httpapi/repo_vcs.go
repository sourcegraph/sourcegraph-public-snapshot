package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
)

func serveRepoResolveRev(w http.ResponseWriter, r *http.Request) error {
	repoRev := routevar.ToRepoRev(mux.Vars(r))
	res, err := resolveLocalRepoRev(r.Context(), repoRev)
	if err != nil {
		return err
	}

	var cacheControl string
	if len(repoRev.Rev) == 40 {
		cacheControl = "private, max-age=600"
	} else {
		cacheControl = "private, max-age=15"
	}
	w.Header().Set("cache-control", cacheControl)
	return writeJSON(w, res)
}

func serveCommit(w http.ResponseWriter, r *http.Request) error {
	rawRepoRev := routevar.ToRepoRev(mux.Vars(r))
	abs := len(rawRepoRev.Rev) == 40 // absolute commit ID?

	repoRev, err := resolveLocalRepoRev(r.Context(), rawRepoRev)
	if err != nil {
		return err
	}
	commit, err := backend.Repos.GetCommit(r.Context(), repoRev)
	if err != nil {
		return err
	}

	var cacheControl string
	if abs {
		cacheControl = "private, max-age=3600"
	} else {
		cacheControl = "private, max-age=15"
	}
	w.Header().Set("cache-control", cacheControl)
	return writeJSON(w, commit)
}

func serveRepoCommits(w http.ResponseWriter, r *http.Request) error {
	repoID, err := resolveLocalRepo(r.Context(), routevar.ToRepo(mux.Vars(r)))
	if err != nil {
		return err
	}

	var opt sourcegraph.RepoListCommitsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	commits, err := backend.Repos.ListCommits(r.Context(), &sourcegraph.ReposListCommitsOp{Repo: repoID, Opt: &opt})
	if err != nil {
		return err
	}

	var cacheControl string
	if len(opt.Head) == 40 && (len(opt.Base) == 0 || len(opt.Base) == 40) {
		cacheControl = "private, max-age=600"
	} else {
		cacheControl = "private, max-age=15"
	}
	w.Header().Set("cache-control", cacheControl)
	return writeJSON(w, commits)
}

func serveRepoRefresh(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.MirrorReposRefreshVCSOp
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	repo, err := handlerutil.GetRepoID(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	actor := auth.ActorFromContext(r.Context())
	repoupdater.Enqueue(repo, actor.UserSpec())
	w.WriteHeader(http.StatusAccepted)
	return nil
}

func serveRepoBranches(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoListBranchesOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	repo, err := handlerutil.GetRepoID(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	branches, err := backend.Repos.ListBranches(r.Context(), &sourcegraph.ReposListBranchesOp{Repo: repo, Opt: &opt})
	if err != nil {
		return err
	}
	return writeJSON(w, branches)
}

func serveRepoTags(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoListTagsOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	repo, err := handlerutil.GetRepoID(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	tags, err := backend.Repos.ListTags(r.Context(), &sourcegraph.ReposListTagsOp{Repo: repo, Opt: &opt})
	if err != nil {
		return err
	}
	return writeJSON(w, tags)
}

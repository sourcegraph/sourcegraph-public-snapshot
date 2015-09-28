package httpapi

import (
	"net/http"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func serveRepoBranches(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var opt sourcegraph.RepoListBranchesOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	_, repoSpec, err := handlerutil.GetRepo(r, s.Repos)
	if err != nil {
		return err
	}

	branches, err := s.Repos.ListBranches(ctx, &sourcegraph.ReposListBranchesOp{Repo: repoSpec, Opt: &opt})
	if err != nil {
		return err
	}
	return writeJSON(w, branches)
}

func serveRepoTags(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var opt sourcegraph.RepoListTagsOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	_, repoSpec, err := handlerutil.GetRepo(r, s.Repos)
	if err != nil {
		return err
	}

	tags, err := s.Repos.ListTags(ctx, &sourcegraph.ReposListTagsOp{Repo: repoSpec, Opt: &opt})
	if err != nil {
		return err
	}
	return writeJSON(w, tags)
}

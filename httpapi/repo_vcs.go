package httpapi

import (
	"net/http"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoBranches(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	var opt sourcegraph.RepoListBranchesOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	_, repoSpec, err := handlerutil.GetRepo(r, cl.Repos)
	if err != nil {
		return err
	}

	branches, err := cl.Repos.ListBranches(ctx, &sourcegraph.ReposListBranchesOp{Repo: repoSpec, Opt: &opt})
	if err != nil {
		return err
	}
	return writeJSON(w, branches)
}

func serveRepoTags(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	var opt sourcegraph.RepoListTagsOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	_, repoSpec, err := handlerutil.GetRepo(r, cl.Repos)
	if err != nil {
		return err
	}

	tags, err := cl.Repos.ListTags(ctx, &sourcegraph.ReposListTagsOp{Repo: repoSpec, Opt: &opt})
	if err != nil {
		return err
	}
	return writeJSON(w, tags)
}

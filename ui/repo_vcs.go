package ui

import (
	"net/http"

	"github.com/sourcegraph/mux"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoCommits(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	repo, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	var opt sourcegraph.RepoListCommitsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	commits, err := cl.Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{Repo: repo, Opt: &opt})
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
	return httputil.WriteJSON(w, commits)
}

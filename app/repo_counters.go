package app

import (
	"net/http"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func serveRepoCounters(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	rc, err := handlerutil.GetRepoCommon(r, nil)
	if err != nil {
		return err
	}

	repoSpec := rc.Repo.RepoSpec()
	counters, err := apiclient.RepoBadges.ListCounters(ctx, &repoSpec)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/counters.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Counters []*sourcegraph.Counter

		tmpl.Common
	}{
		RepoCommon: *rc,
		Counters:   counters.Counters,
	})
}

package app

import (
	"net/http"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func serveRepoBadges(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r, nil)
	if err != nil {
		return err
	}

	repoSpec := rc.Repo.RepoSpec()
	badges, err := apiclient.RepoBadges.ListBadges(ctx, &repoSpec)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/badges.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Badges []*sourcegraph.Badge

		tmpl.Common
	}{
		RepoCommon: *rc,
		Badges:     badges.Badges,
	})
}

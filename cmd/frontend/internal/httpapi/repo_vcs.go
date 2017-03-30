package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

func serveRepoResolveRev(w http.ResponseWriter, r *http.Request) error {
	repoRev := routevar.ToRepoRev(mux.Vars(r))
	repo, err := backend.Repos.GetByURI(r.Context(), repoRev.Repo)
	if err != nil {
		return err
	}
	rev, err := backend.Repos.ResolveRev(r.Context(), &sourcegraph.ReposResolveRevOp{
		Repo: repo.ID,
		Rev:  repoRev.Rev,
	})
	if err != nil {
		return err
	}

	if err := backend.Repos.RefreshIndex(r.Context(), repoRev.Repo); err != nil {
		return err
	}

	var cacheControl string
	if len(repoRev.Rev) == 40 {
		cacheControl = "private, max-age=600"
	} else {
		cacheControl = "private, max-age=15"
	}
	w.Header().Set("cache-control", cacheControl)
	return writeJSON(w, &sourcegraph.RepoRevSpec{
		Repo:     repo.ID,
		CommitID: rev.CommitID,
	})
}

package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func serveRepoResolveRev(w http.ResponseWriter, r *http.Request) error {
	repoRev := routevar.ToRepoRev(mux.Vars(r))
	res, err := resolveLocalRepoRev(r.Context(), repoRev)
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
	return writeJSON(w, res)
}

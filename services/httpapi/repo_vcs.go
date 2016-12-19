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

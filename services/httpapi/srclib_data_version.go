package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func serveSrclibDataVersion(w http.ResponseWriter, r *http.Request) error {
	repoRev, err := resolveLocalRepoRev(r.Context(), routevar.ToRepoRev(mux.Vars(r)))
	if err != nil {
		return err
	}

	var opt struct {
		Path string
	}
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dataVer, err := backend.Repos.GetSrclibDataVersionForPath(r.Context(), &sourcegraph.TreeEntrySpec{
		RepoRev: *repoRev,
		Path:    opt.Path,
	})
	if err != nil {
		return err
	}
	return writeJSON(w, dataVer)
}

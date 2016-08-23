package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func serveSrclibDataVersion(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

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

	dataVer, err := cl.Repos.GetSrclibDataVersionForPath(r.Context(), &sourcegraph.TreeEntrySpec{
		RepoRev: *repoRev,
		Path:    opt.Path,
	})
	if err != nil {
		return err
	}
	return writeJSON(w, dataVer)
}

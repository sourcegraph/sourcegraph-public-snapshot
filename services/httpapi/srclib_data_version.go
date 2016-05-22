package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveSrclibDataVersion(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	repoRev, err := resolveRepoRev(ctx, routevar.ToRepoRev(mux.Vars(r)))
	if err != nil {
		return err
	}

	var opt struct {
		Path string
	}
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dataVer, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
		RepoRev: *repoRev,
		Path:    opt.Path,
	})
	if err != nil {
		return err
	}
	return writeJSON(w, dataVer)
}

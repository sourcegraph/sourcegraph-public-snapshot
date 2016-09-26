package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

// serveRefreshIndexes is not meant to be called from the UI, but is intended
// to be used by sourcegraph operators to manually update indexes.
func serveRefreshIndexes(w http.ResponseWriter, r *http.Request) error {
	repo, err := handlerutil.GetRepoID(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	var opt struct {
		Blocking bool
	}
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	cl := handlerutil.Client(r)
	if opt.Blocking {
		_, err = cl.Defs.RefreshIndex(r.Context(), &sourcegraph.DefsRefreshIndexOp{
			Repo:                repo,
			RefreshRefLocations: true,
			Force:               true,
		})
	} else {
		_, err = cl.Async.RefreshIndexes(r.Context(), &sourcegraph.AsyncRefreshIndexesOp{
			Repo:   repo,
			Source: "httpapi",
			Force:  true,
		})
	}
	if err != nil {
		return err
	}

	return writeJSON(w, map[string]string{"status": "ok"})
}

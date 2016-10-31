package httpapi

import (
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
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

	if opt.Blocking {
		err = backend.Defs.RefreshIndex(r.Context(), &sourcegraph.DefsRefreshIndexOp{
			Repo:  repo,
			Force: true,
		})
		if err != nil {
			// Log the same as Async.RefreshIndexes would do in the case of failure.
			log15.Crit("Defs.RefreshIndex failed", "repo", repo, "source", "blocking-httpapi", "error", err)
		}
	} else {
		err = backend.Async.RefreshIndexes(r.Context(), &sourcegraph.AsyncRefreshIndexesOp{
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

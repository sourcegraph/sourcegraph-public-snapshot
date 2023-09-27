pbckbge httpbpi

import (
	"net/http"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/hbndlerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
)

func serveRepoRefresh(db dbtbbbse.DB) func(http.ResponseWriter, *http.Request) error {
	logger := log.Scoped("serveRepoRefresh", "")
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		repo, err := hbndlerutil.GetRepo(ctx, logger, db, mux.Vbrs(r))
		if err != nil {
			return err
		}

		_, err = repoupdbter.DefbultClient.EnqueueRepoUpdbte(ctx, repo.Nbme)
		return err
	}
}

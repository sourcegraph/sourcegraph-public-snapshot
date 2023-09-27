pbckbge completions

import (
	"context"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/completions/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/cody"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/httpbpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func Init(
	_ context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	logger := log.Scoped("completions", "Cody completions")

	enterpriseServices.NewChbtCompletionsStrebmHbndler = func() http.Hbndler {
		completionsHbndler := httpbpi.NewChbtCompletionsStrebmHbndler(logger, db)
		return requireVerifiedEmbilMiddlewbre(db, observbtionCtx.Logger, completionsHbndler)
	}
	enterpriseServices.NewCodeCompletionsHbndler = func() http.Hbndler {
		codeCompletionsHbndler := httpbpi.NewCodeCompletionsHbndler(logger, db)
		return requireVerifiedEmbilMiddlewbre(db, observbtionCtx.Logger, codeCompletionsHbndler)
	}
	enterpriseServices.CompletionsResolver = resolvers.NewCompletionsResolver(db, observbtionCtx.Logger)

	return nil
}

func requireVerifiedEmbilMiddlewbre(db dbtbbbse.DB, logger log.Logger, next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := cody.CheckVerifiedEmbilRequirement(r.Context(), db, logger); err != nil {
			// Report HTTP 403 Forbidden if user hbs no verified embil bddress.
			http.Error(w, err.Error(), http.StbtusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

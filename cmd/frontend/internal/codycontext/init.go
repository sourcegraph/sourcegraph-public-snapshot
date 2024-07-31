package codycontext

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	newembeddings "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/embeddings"
)

func Init(
	_ context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	logger := log.Scoped("embeddings")

	enterpriseServices.NewEmbeddingsHandler = func() http.Handler {
		completionsHandler := newembeddings.NewEmbeddingsChunkingHandler(logger, db)
		return requireVerifiedEmailMiddleware(db, observationCtx.Logger, completionsHandler)
	}

	return nil
}

func requireVerifiedEmailMiddleware(db database.DB, logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := cody.CheckVerifiedEmailRequirement(r.Context(), db, logger); err != nil {
			// Report HTTP 403 Forbidden if user has no verified email address.
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

package compute

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/compute/resolvers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/compute/streaming"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	logger := log.Scoped("compute")
	enterpriseServices.ComputeResolver = resolvers.NewResolver(logger, db)
	enterpriseServices.NewComputeStreamHandler = func() http.Handler {
		return streaming.NewComputeStreamHandler(logger, db)
	}
	return nil
}

package compute

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/compute/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/compute/streaming"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(ctx context.Context, db database.DB, _ conftypes.UnifiedWatchable, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	enterpriseServices.ComputeResolver = resolvers.NewResolver(db)
	enterpriseServices.NewComputeStreamHandler = func() http.Handler { return streaming.NewComputeStreamHandler(db) }
	return nil
}

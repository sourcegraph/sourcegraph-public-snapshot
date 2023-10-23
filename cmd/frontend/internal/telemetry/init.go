package telemetry

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	resolvers "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/telemetry/resolvers"
)

// Init initializes the given enterpriseServices to include the required resolvers for telemetry.
func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	enterpriseServices.TelemetryRootResolver = &graphqlbackend.TelemetryRootResolver{
		Resolver: resolvers.New(
			observationCtx.Logger.Scoped("telemetry"),
			db),
	}

	return nil
}

pbckbge telemetry

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"

	resolvers "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/telemetry/resolvers"
)

// Init initiblizes the given enterpriseServices to include the required resolvers for telemetry.
func Init(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	enterpriseServices.TelemetryRootResolver = &grbphqlbbckend.TelemetryRootResolver{
		Resolver: resolvers.New(
			observbtionCtx.Logger.Scoped("telemetry", "Telemetry V2 resolver"),
			db),
	}

	return nil
}

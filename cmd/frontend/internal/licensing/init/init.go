package init

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/licensing/resolvers"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	codeIntelServices codeintel.Services,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	// Enforce the license's feature check for monitoring. If the license does not support the monitoring
	// feature, then alternative debug handlers will be invoked.
	// Uncomment this when licensing for FeatureMonitoring should be enforced.
	// See PR https://github.com/sourcegraph/sourcegraph/issues/42527 for more context.
	// app.SetPreMountGrafanaHook(enforcement.NewPreMountGrafanaHook())

	enterpriseServices.LicenseResolver = resolvers.LicenseResolver{}

	if envvar.SourcegraphDotComMode() {
		logger := log.Scoped("licensing")
		goroutine.Go(func() {
			productsubscription.StartCheckForUpcomingLicenseExpirations(logger, db)
		})
		goroutine.Go(func() {
			productsubscription.StartCheckForAnomalousLicenseUsage(logger, db)
		})
	}

	return nil
}

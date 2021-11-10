package insights

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// Init initializes the given enterpriseServices to include the required resolvers for insights.
func Init(ctx context.Context, postgres dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services, observationContext *observation.Context, datastore *codeintel.DataStores) error {
	if !insights.IsEnabled() {
		if conf.IsDeployTypeSingleDockerContainer(conf.DeployType()) {
			enterpriseServices.InsightsResolver = resolvers.NewDisabledResolver("backend-run code insights are not available on single-container deployments")
		} else {
			enterpriseServices.InsightsResolver = resolvers.NewDisabledResolver("code insights has been disabled")
		}
		return nil
	}
	timescale, err := insights.InitializeCodeInsightsDB("frontend")
	if err != nil {
		return err
	}
	enterpriseServices.InsightsResolver = resolvers.New(timescale, postgres)
	return nil
}

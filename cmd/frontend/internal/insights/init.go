package insights

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/insights/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/insights/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/insights"
	insightsdb "github.com/sourcegraph/sourcegraph/internal/insights/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Init initializes the given enterpriseServices to include the required resolvers for insights.
func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	enterpriseServices.InsightsAggregationResolver = resolvers.NewAggregationResolver(observationCtx, db)

	if !insights.IsEnabled() {
		if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) {
			enterpriseServices.InsightsResolver = resolvers.NewDisabledResolver("code insights are not available on single-container deployments")
		} else {
			enterpriseServices.InsightsResolver = resolvers.NewDisabledResolver("code insights has been disabled")
		}
		return nil
	}
	rawInsightsDB, err := insightsdb.InitializeCodeInsightsDB(observationCtx, "frontend")
	if err != nil {
		return err
	}
	enterpriseServices.InsightsResolver = resolvers.New(rawInsightsDB, db)
	enterpriseServices.CodeInsightsDataExportHandler = httpapi.NewExportHandler(db, rawInsightsDB).ExportFunc()

	return nil
}

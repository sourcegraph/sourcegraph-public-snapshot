package insights

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func IsEnabled() bool {
	return enterprise.IsCodeInsightsEnabled()
}

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

	if !IsEnabled() {
		if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) {
			enterpriseServices.InsightsResolver = resolvers.NewDisabledResolver("code insights are not available on single-container deployments")
		} else {
			enterpriseServices.InsightsResolver = resolvers.NewDisabledResolver("code insights has been disabled")
		}
		return nil
	}
	rawInsightsDB, err := InitializeCodeInsightsDB(observationCtx, "frontend")
	if err != nil {
		return err
	}
	enterpriseServices.InsightsResolver = resolvers.New(rawInsightsDB, db)

	return nil
}

// InitializeCodeInsightsDB connects to and initializes the Code Insights Postgres DB, running
// database migrations before returning. It is safe to call from multiple services/containers (in
// which case, one's migration will win and the other caller will receive an error and should exit
// and restart until the other finishes.)
func InitializeCodeInsightsDB(observationCtx *observation.Context, app string) (edb.InsightsDB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeInsightsDSN
	})
	db, err := connections.EnsureNewCodeInsightsDB(observationCtx, dsn, app)
	if err != nil {
		return nil, errors.Errorf("Failed to connect to codeinsights database: %s", err)
	}

	return edb.NewInsightsDB(db, observationCtx.Logger), nil
}

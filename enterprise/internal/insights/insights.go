package insights

import (
	"context"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
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

// IsEnabled tells if code insights are enabled or not.
func IsEnabled() bool {
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_INSIGHTS")); v {
		// Code insights can always be disabled. This can be a helpful escape hatch if e.g. there
		// are issues with (or connecting to) the codeinsights-db deployment and it is preventing
		// the Sourcegraph frontend or repo-updater from starting.
		//
		// It is also useful in dev environments if you do not wish to spend resources running Code
		// Insights.
		return false
	}
	if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) {
		// Code insights is not supported in single-container Docker demo deployments unless
		// explicity allowed, (for example by backend integration tests.)
		if v, _ := strconv.ParseBool(os.Getenv("ALLOW_SINGLE_DOCKER_CODE_INSIGHTS")); v {
			return true
		}
		return false
	}
	return true
}

// Init initializes the given enterpriseServices to include the required resolvers for insights.
func Init(ctx context.Context, postgres database.DB, _ conftypes.UnifiedWatchable, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	enterpriseServices.InsightsAggregationResolver = resolvers.NewAggregationResolver(postgres, observationContext)

	if !IsEnabled() {
		if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) {
			enterpriseServices.InsightsResolver = resolvers.NewDisabledResolver("code insights are not available on single-container deployments")
		} else {
			enterpriseServices.InsightsResolver = resolvers.NewDisabledResolver("code insights has been disabled")
		}
		return nil
	}
	db, err := InitializeCodeInsightsDB("frontend")
	if err != nil {
		return err
	}
	enterpriseServices.InsightsResolver = resolvers.New(db, postgres)

	return nil
}

// InitializeCodeInsightsDB connects to and initializes the Code Insights Postgres DB, running
// database migrations before returning. It is safe to call from multiple services/containers (in
// which case, one's migration will win and the other caller will receive an error and should exit
// and restart until the other finishes.)
func InitializeCodeInsightsDB(app string) (edb.InsightsDB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeInsightsDSN
	})
	db, err := connections.EnsureNewCodeInsightsDB(dsn, app, &observation.TestContext)
	if err != nil {
		return nil, errors.Errorf("Failed to connect to codeinsights database: %s", err)
	}

	return edb.NewInsightsDB(db), nil
}

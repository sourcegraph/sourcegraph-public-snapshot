package database

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// InitializeCodeInsightsDB connects to and initializes the Code Insights Postgres DB, running
// database migrations before returning. It is safe to call from multiple services/containers (in
// which case, one's migration will win and the other caller will receive an error and should exit
// and restart until the other finishes.)
func InitializeCodeInsightsDB(observationCtx *observation.Context, app string) (database.InsightsDB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeInsightsDSN
	})
	db, err := connections.EnsureNewCodeInsightsDB(observationCtx, dsn, app)
	if err != nil {
		return nil, errors.Errorf("Failed to connect to codeinsights database: %s", err)
	}

	return database.NewInsightsDB(db, observationCtx.Logger), nil
}

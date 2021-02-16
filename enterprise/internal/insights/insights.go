package insights

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// IsEnabled tells if code insights are enabled or not.
func IsEnabled() bool {
	if !conf.IsDev(conf.DeployType()) {
		// Code Insights is not yet deployed to non-dev/testing instances. We don't yet have
		// TimescaleDB in those deployments. https://github.com/sourcegraph/sourcegraph/issues/17218
		return false
	}
	if conf.IsDeployTypeSingleDockerContainer(conf.DeployType()) {
		// Code insights is not supported in single-container Docker demo deployments.
		return false
	}
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_INSIGHTS")); v {
		// Dev option for disabling code insights. Helpful if e.g. you have issues running the
		// codeinsights-db or don't want to spend resources on it.
		return false
	}
	return true
}

// Init initializes the given enterpriseServices to include the required resolvers for insights.
func Init(ctx context.Context, postgres dbutil.DB, enterpriseServices *enterprise.Services) error {
	if !IsEnabled() {
		return nil
	}
	timescale, err := InitializeCodeInsightsDB()
	if err != nil {
		return err
	}
	enterpriseServices.InsightsResolver = resolvers.New(timescale, postgres)
	return nil
}

// InitializeCodeInsightsDB connects to and initializes the Code Insights Timescale DB, running
// database migrations before returning. It is safe to call from multiple services/containers (in
// which case, one's migration will win and the other caller will receive an error and should exit
// and restart until the other finishes.)
func InitializeCodeInsightsDB() (*sql.DB, error) {
	timescaleDSN := conf.Get().ServiceConnections.CodeInsightsTimescaleDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.CodeInsightsTimescaleDSN; timescaleDSN != newDSN {
			log.Fatalf("Detected codeinsights database DSN change, restarting to take effect: %s", newDSN)
		}
	})

	db, err := dbconn.New(timescaleDSN, "")
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to codeinsights database: %s", err)
	}

	if err := dbconn.MigrateDB(db, dbconn.CodeInsights); err != nil {
		return nil, fmt.Errorf("Failed to perform codeinsights database migration: %s", err)
	}
	return db, nil
}

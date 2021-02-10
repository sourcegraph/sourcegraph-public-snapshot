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
)

// Init initializes the given enterpriseServices to include the required resolvers for insights.
func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	resolver, err := InitResolver(ctx, dbconn.Global)
	if err != nil {
		return err
	}
	if resolver == nil {
		return nil // e.g. code insights is disabled or not supported in this deployment type.
	}
	enterpriseServices.InsightsResolver = resolver
	return nil
}

// InitResolver connects to and initializes TimescaleDB and returns an initialized resolver.
func InitResolver(ctx context.Context, postgresAppDB *sql.DB) (*resolvers.Resolver, error) {
	if !conf.IsDev(conf.DeployType()) {
		// Code Insights is not yet deployed to non-dev/testing instances. We don't yet have
		// TimescaleDB in those deployments. https://github.com/sourcegraph/sourcegraph/issues/17218
		return nil, nil
	}
	if conf.IsDeployTypeSingleDockerContainer(conf.DeployType()) {
		// Code insights is not supported in single-container Docker demo deployments.
		return nil, nil
	}
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_INSIGHTS")); v {
		// Dev option for disabling code insights. Helpful if e.g. you have issues running the
		// codeinsights-db or don't want to spend resources on it.
		return nil, nil
	}
	timescale, err := initializeCodeInsightsDB()
	if err != nil {
		return nil, err
	}
	return resolvers.New(timescale, postgresAppDB), nil
}

// initializeCodeInsightsDB connects to and initializes the Code Insights Timescale DB, running
// database migrations before returning.
func initializeCodeInsightsDB() (*sql.DB, error) {
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

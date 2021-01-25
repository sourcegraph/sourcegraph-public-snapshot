package insights

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// Init initializes the given enterpriseServices to include the required resolvers for insights.
func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	db, err := initializeCodeInsightsDB()
	if err != nil {
		return err
	}
	enterpriseServices.InsightsResolver = resolvers.New(db)
	return nil
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

	if err := dbconn.MigrateDB(db, "codeinsights"); err != nil {
		return nil, fmt.Errorf("Failed to perform codeinsights database migration: %s", err)
	}
	return db, nil
}

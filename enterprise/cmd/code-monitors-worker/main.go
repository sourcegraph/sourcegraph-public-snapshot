package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/background"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func main() {
	background.StartBackgroundJobs(context.Background(), mustInitializeDB())
}

func mustInitializeDB() *sql.DB {
	postgresDSN := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.PostgresDSN; postgresDSN != newDSN {
			log.Fatalf("Detected database DSN change, restarting to take effect: %s", newDSN)
		}
	})

	if err := dbconn.SetupGlobalConnection(postgresDSN); err != nil {
		log.Fatalf("Failed to connect to frontend database: %s", err)
	}

	//
	// START FLAILING

	ctx := context.Background()
	go func() {
		for range time.NewTicker(5 * time.Second).C {
			allowAccessByDefault, authzProviders, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	// END FLAILING
	//

	return dbconn.Global
}

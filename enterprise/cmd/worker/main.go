package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/cmd/worker/workerdb"
	batchesjanitor "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches/janitor"
	batchesmigrations "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches/migrations"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel"
	codeintelmigrations "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/migrations"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executors"
	workerinsights "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/insights"
	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/versions"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	go setAuthzProviders()

	additionalJobs := map[string]job.Job{
		"codeintel-commitgraph":    codeintel.NewCommitGraphJob(),
		"codeintel-janitor":        codeintel.NewJanitorJob(),
		"codeintel-auto-indexing":  codeintel.NewIndexingJob(),
		"codehost-version-syncing": versions.NewSyncingJob(),
		"insights-job":             workerinsights.NewInsightsJob(),
		"batches-janitor":          batchesjanitor.NewJanitorJob(),
		"executors-janitor":        executors.NewJanitorJob(),
	}

	shared.Start(additionalJobs, registerEnterpriseMigrations)
}

func init() {
	oobmigration.ReturnEnterpriseMigrations = true
}

// setAuthProviders waits for the database to be initialized, then periodically refreshes the
// global authz providers. This changes the repositories that are visible for reads based on the
// current actor stored in an operation's context, which is likely an internal actor for many of
// the jobs configured in this service. This also enables repository update operations to fetch
// permissions from code hosts.
func setAuthzProviders() {
	db, err := workerdb.Init()
	if err != nil {
		return
	}

	ctx := context.Background()

	for range time.NewTicker(5 * time.Second).C {
		allowAccessByDefault, authzProviders, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), database.ExternalServices(db))
		authz.SetProviders(allowAccessByDefault, authzProviders)
	}
}

func registerEnterpriseMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	if err := batchesmigrations.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	if err := codeintelmigrations.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	if err := insights.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	return nil
}

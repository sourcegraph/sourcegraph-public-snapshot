package main

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches"
	batchesmigrations "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches/migrations"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel"
	freshcodeintel "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/fresh"
	codeintelmigrations "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/migrations"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executors"
	workerinsights "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/insights"
	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/versions"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

func main() {
	syncLogs := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer syncLogs()

	logger := log.Scoped("worker", "worker enterprise edition")

	go setAuthzProviders()

	additionalJobs := map[string]job.Job{
		"codehost-version-syncing":   versions.NewSyncingJob(),
		"insights-job":               workerinsights.NewInsightsJob(),
		"insights-query-runner-job":  workerinsights.NewInsightsQueryRunnerJob(),
		"batches-janitor":            batches.NewJanitorJob(),
		"batches-scheduler":          batches.NewSchedulerJob(),
		"batches-reconciler":         batches.NewReconcilerJob(),
		"batches-bulk-processor":     batches.NewBulkOperationProcessorJob(),
		"batches-workspace-resolver": batches.NewWorkspaceResolverJob(),
		"executors-janitor":          executors.NewJanitorJob(),
		"codemonitors-job":           codemonitors.NewCodeMonitorJob(),

		// fresh
		"codeintel-upload-janitor":         freshcodeintel.NewUploadJanitorJob(),
		"codeintel-upload-expirer":         freshcodeintel.NewUploadExpirerJob(),
		"codeintel-commitgraph-updater":    freshcodeintel.NewCommitGraphUpdaterJob(),
		"codeintel-autoindexing-scheduler": freshcodeintel.NewAutoindexingSchedulerJob(),

		// temporary
		"codeintel-janitor":       codeintel.NewJanitorJob(),
		"codeintel-auto-indexing": codeintel.NewIndexingJob(),
	}

	if err := shared.Start(logger, additionalJobs, registerEnterpriseMigrations); err != nil {
		logger.Fatal(err.Error())
	}
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
	sqlDB, err := workerdb.Init()
	if err != nil {
		return
	}

	ctx := context.Background()
	db := database.NewDB(sqlDB)

	for range time.NewTicker(eiauthz.RefreshInterval()).C {
		allowAccessByDefault, authzProviders, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices(), db)
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

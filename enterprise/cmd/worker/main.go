package main

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/telemetry"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel"
	freshcodeintel "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/fresh"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executors"
	workerinsights "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/permissions"
	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/versions"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func main() {
	liblog := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer liblog.Sync()

	logger := log.Scoped("worker", "worker enterprise edition")

	go setAuthzProviders(logger)

	additionalJobs := map[string]job.Job{
		"codehost-version-syncing":      versions.NewSyncingJob(),
		"insights-job":                  workerinsights.NewInsightsJob(),
		"insights-query-runner-job":     workerinsights.NewInsightsQueryRunnerJob(),
		"batches-janitor":               batches.NewJanitorJob(),
		"batches-scheduler":             batches.NewSchedulerJob(),
		"batches-reconciler":            batches.NewReconcilerJob(),
		"batches-bulk-processor":        batches.NewBulkOperationProcessorJob(),
		"batches-workspace-resolver":    batches.NewWorkspaceResolverJob(),
		"executors-janitor":             executors.NewJanitorJob(),
		"executors-metricsserver":       executors.NewMetricsServerJob(),
		"codemonitors-job":              codemonitors.NewCodeMonitorJob(),
		"bitbucket-project-permissions": permissions.NewBitbucketProjectPermissionsJob(),
		"export-usage-telemetry":        telemetry.NewTelemetryJob(),
		"webhook-build-job":             repos.NewWebhookBuildJob(),

		// fresh
		"codeintel-upload-janitor":         freshcodeintel.NewUploadJanitorJob(),
		"codeintel-upload-expirer":         freshcodeintel.NewUploadExpirerJob(),
		"codeintel-commitgraph-updater":    freshcodeintel.NewCommitGraphUpdaterJob(),
		"codeintel-upload-backfiller":      freshcodeintel.NewUploadBackfillerJob(),
		"codeintel-autoindexing-scheduler": freshcodeintel.NewAutoindexingSchedulerJob(),

		// temporary
		"codeintel-janitor":       codeintel.NewJanitorJob(),
		"codeintel-auto-indexing": codeintel.NewIndexingJob(),
	}

	if err := shared.Start(logger, additionalJobs, migrations.RegisterEnterpriseMigrators); err != nil {
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
func setAuthzProviders(logger log.Logger) {
	sqlDB, err := workerdb.Init()
	if err != nil {
		return
	}

	// authz also relies on UserMappings being setup.
	globals.WatchPermissionsUserMapping()

	ctx := context.Background()
	db := database.NewDB(logger, sqlDB)

	for range time.NewTicker(eiauthz.RefreshInterval()).C {
		allowAccessByDefault, authzProviders, _, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices(), db)
		authz.SetProviders(allowAccessByDefault, authzProviders)
	}
}

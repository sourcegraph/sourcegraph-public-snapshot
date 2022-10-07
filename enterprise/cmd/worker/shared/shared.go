package shared

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/worker/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executors"
	workerinsights "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/permissions"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/telemetry"
	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/versions"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos"
)

func AdditionalJobs(observationContext *observation.Context) map[string]job.Job {
	return map[string]job.Job{
		"codehost-version-syncing":      versions.NewSyncingJob(observationContext),
		"insights-job":                  workerinsights.NewInsightsJob(observationContext),
		"insights-query-runner-job":     workerinsights.NewInsightsQueryRunnerJob(observationContext),
		"batches-janitor":               batches.NewJanitorJob(),
		"batches-scheduler":             batches.NewSchedulerJob(),
		"batches-reconciler":            batches.NewReconcilerJob(),
		"batches-bulk-processor":        batches.NewBulkOperationProcessorJob(),
		"batches-workspace-resolver":    batches.NewWorkspaceResolverJob(),
		"executors-janitor":             executors.NewJanitorJob(observationContext),
		"executors-metricsserver":       executors.NewMetricsServerJob(),
		"codemonitors-job":              codemonitors.NewCodeMonitorJob(observationContext),
		"bitbucket-project-permissions": permissions.NewBitbucketProjectPermissionsJob(observationContext),
		"export-usage-telemetry":        telemetry.NewTelemetryJob(observationContext),
		"webhook-build-job":             repos.NewWebhookBuildJob(observationContext),

		"codeintel-policies-repository-matcher":       codeintel.NewPoliciesRepositoryMatcherJob(observationContext),
		"codeintel-autoindexing-dependency-scheduler": codeintel.NewAutoindexingDependencySchedulerJob(observationContext),
		"codeintel-autoindexing-janitor":              codeintel.NewAutoindexingJanitorJob(observationContext),
		"codeintel-autoindexing-scheduler":            codeintel.NewAutoindexingSchedulerJob(observationContext),
		"codeintel-commitgraph-updater":               codeintel.NewCommitGraphUpdaterJob(observationContext),
		"codeintel-metrics-reporter":                  codeintel.NewMetricsReporterJob(observationContext),
		"codeintel-upload-backfiller":                 codeintel.NewUploadBackfillerJob(observationContext),
		"codeintel-upload-expirer":                    codeintel.NewUploadExpirerJob(observationContext),
		"codeintel-upload-janitor":                    codeintel.NewUploadJanitorJob(observationContext),
		"codeintel-upload-graph-exporter":             codeintel.NewGraphExporterJob(observationContext),
		"codeintel-uploadstore-expirer":               codeintel.NewPreciseCodeIntelUploadExpirer(observationContext),

		"auth-sourcegraph-operator-cleaner": auth.NewSourcegraphOperatorCleaner(observationContext),

		// Note: experimental (not documented)
		"codeintel-ranking-sourcer": codeintel.NewRankingSourcerJob(observationContext),
	}
}

// SetAuthProviders waits for the database to be initialized, then periodically refreshes the
// global authz providers. This changes the repositories that are visible for reads based on the
// current actor stored in an operation's context, which is likely an internal actor for many of
// the jobs configured in this service. This also enables repository update operations to fetch
// permissions from code hosts.
func SetAuthzProviders(logger log.Logger, observationContext *observation.Context) {
	db, err := workerdb.InitDBWithLogger(logger, observationContext)
	if err != nil {
		return
	}

	// authz also relies on UserMappings being setup.
	globals.WatchPermissionsUserMapping()

	ctx := context.Background()

	for range time.NewTicker(eiauthz.RefreshInterval()).C {
		allowAccessByDefault, authzProviders, _, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices(), db)
		authz.SetProviders(allowAccessByDefault, authzProviders)
	}
}

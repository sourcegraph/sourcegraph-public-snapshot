package shared

import (
	"context"
	"time"

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

var AdditionalJobs = map[string]job.Job{
	"codehost-version-syncing":      versions.NewSyncingJob(),
	"insights-job":                  workerinsights.NewInsightsJob(),
	"insights-query-runner-job":     workerinsights.NewInsightsQueryRunnerJob(),
	"insights-data-retention-job":   workerinsights.NewInsightsDataRetentionJob(),
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

	"codeintel-policies-repository-matcher":       codeintel.NewPoliciesRepositoryMatcherJob(),
	"codeintel-autoindexing-dependency-scheduler": codeintel.NewAutoindexingDependencySchedulerJob(),
	"codeintel-autoindexing-janitor":              codeintel.NewAutoindexingJanitorJob(),
	"codeintel-autoindexing-scheduler":            codeintel.NewAutoindexingSchedulerJob(),
	"codeintel-commitgraph-updater":               codeintel.NewCommitGraphUpdaterJob(),
	"codeintel-metrics-reporter":                  codeintel.NewMetricsReporterJob(),
	"codeintel-upload-backfiller":                 codeintel.NewUploadBackfillerJob(),
	"codeintel-upload-expirer":                    codeintel.NewUploadExpirerJob(),
	"codeintel-upload-janitor":                    codeintel.NewUploadJanitorJob(),
	"codeintel-upload-graph-exporter":             codeintel.NewGraphExporterJob(),
	"codeintel-uploadstore-expirer":               codeintel.NewPreciseCodeIntelUploadExpirer(),

	"auth-sourcegraph-operator-cleaner": auth.NewSourcegraphOperatorCleaner(),

	// Note: experimental (not documented)
	"codeintel-ranking-sourcer": codeintel.NewRankingSourcerJob(),
}

// SetAuthProviders waits for the database to be initialized, then periodically refreshes the
// global authz providers. This changes the repositories that are visible for reads based on the
// current actor stored in an operation's context, which is likely an internal actor for many of
// the jobs configured in this service. This also enables repository update operations to fetch
// permissions from code hosts.
func SetAuthzProviders(observationCtx *observation.Context) {
	observationCtx = observation.ContextWithLogger(observationCtx.Logger.Scoped("authz-provider", ""), observationCtx)

	db, err := workerdb.InitDB(observationCtx)
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

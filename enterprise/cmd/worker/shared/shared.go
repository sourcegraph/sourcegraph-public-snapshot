package shared

import (
	"context"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/own"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/worker/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codemonitors"
	contextdetectionembeddings "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/embeddings/contextdetection"
	repoembeddings "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/embeddings/repo"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executors"
	workerinsights "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/permissions"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/telemetry"
	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	srp "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/subrepoperms"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/versions"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var additionalJobs = map[string]job.Job{
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

	"codeintel-policies-repository-matcher":       codeintel.NewPoliciesRepositoryMatcherJob(),
	"codeintel-autoindexing-summary-builder":      codeintel.NewAutoindexingSummaryBuilder(),
	"codeintel-autoindexing-dependency-scheduler": codeintel.NewAutoindexingDependencySchedulerJob(),
	"codeintel-autoindexing-scheduler":            codeintel.NewAutoindexingSchedulerJob(),
	"codeintel-commitgraph-updater":               codeintel.NewCommitGraphUpdaterJob(),
	"codeintel-metrics-reporter":                  codeintel.NewMetricsReporterJob(),
	"codeintel-upload-backfiller":                 codeintel.NewUploadBackfillerJob(),
	"codeintel-upload-expirer":                    codeintel.NewUploadExpirerJob(),
	"codeintel-upload-janitor":                    codeintel.NewUploadJanitorJob(),
	"codeintel-ranking-file-reference-counter":    codeintel.NewRankingFileReferenceCounter(),
	"codeintel-uploadstore-expirer":               codeintel.NewPreciseCodeIntelUploadExpirer(),
	"codeintel-crates-syncer":                     codeintel.NewCratesSyncerJob(),
	"codeintel-sentinel-cve-scanner":              codeintel.NewSentinelCVEScannerJob(),
	"codeintel-package-filter-applicator":         codeintel.NewPackagesFilterApplicatorJob(),

	"auth-sourcegraph-operator-cleaner":  auth.NewSourcegraphOperatorCleaner(),
	"auth-permission-sync-job-cleaner":   auth.NewPermissionSyncJobCleaner(),
	"auth-permission-sync-job-scheduler": auth.NewPermissionSyncJobScheduler(),

	"repo-embedding-janitor":              repoembeddings.NewRepoEmbeddingJanitorJob(),
	"repo-embedding-job":                  repoembeddings.NewRepoEmbeddingJob(),
	"context-detection-embedding-janitor": contextdetectionembeddings.NewContextDetectionEmbeddingJanitorJob(),
	"context-detection-embedding-job":     contextdetectionembeddings.NewContextDetectionEmbeddingJob(),

	"own-repo-indexing-queue": own.NewOwnRepoIndexingQueue(),
}

// SetAuthzProviders waits for the database to be initialized, then periodically refreshes the
// global authz providers. This changes the repositories that are visible for reads based on the
// current actor stored in an operation's context, which is likely an internal actor for many of
// the jobs configured in this service. This also enables repository update operations to fetch
// permissions from code hosts.
func setAuthzProviders(ctx context.Context, observationCtx *observation.Context) {
	observationCtx = observation.ContextWithLogger(observationCtx.Logger.Scoped("authz-provider", ""), observationCtx)
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return
	}

	// authz also relies on UserMappings being setup.
	globals.WatchPermissionsUserMapping()

	for range time.NewTicker(eiauthz.RefreshInterval()).C {
		allowAccessByDefault, authzProviders, _, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices(), db)
		authz.SetProviders(allowAccessByDefault, authzProviders)
	}
}

func getEnterpriseInit(logger log.Logger) func(database.DB) {
	return func(ossDB database.DB) {
		enterpriseDB := edb.NewEnterpriseDB(ossDB)

		var err error
		authz.DefaultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(enterpriseDB.SubRepoPerms())
		if err != nil {
			logger.Fatal("Failed to create sub-repo client", log.Error(err))
		}
	}
}

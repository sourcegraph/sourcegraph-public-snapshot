package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	UploadStoreConfig                         *uploadstore.Config
	AutoIndexingTaskInterval                  time.Duration
	AutoIndexingSkipManualInterval            time.Duration
	HunkCacheSize                             int
	IndexBatchSize                            int
	MinimumTimeSinceLastEnqueue               time.Duration
	MinimumSearchCount                        int
	MinimumSearchRatio                        int
	MinimumPreciseCount                       int
	DiagnosticsCountMigrationBatchSize        int
	DiagnosticsCountMigrationBatchInterval    time.Duration
	DefinitionsCountMigrationBatchSize        int
	DefinitionsCountMigrationBatchInterval    time.Duration
	ReferencesCountMigrationBatchSize         int
	ReferencesCountMigrationBatchInterval     time.Duration
	DocumentColumnSplitMigrationBatchSize     int
	DocumentColumnSplitMigrationBatchInterval time.Duration
	CommittedAtMigrationBatchSize             int
	CommittedAtMigrationBatchInterval         time.Duration
	DependencyIndexerSchedulerPollInterval    time.Duration
	DependencyIndexerSchedulerConcurrency     int
}

var config = &Config{}

func init() {
	uploadStoreConfig := &uploadstore.Config{}
	uploadStoreConfig.Load()
	config.UploadStoreConfig = uploadStoreConfig

	config.HunkCacheSize = config.GetInt("PRECISE_CODE_INTEL_HUNK_CACHE_SIZE", "1000", "The capacity of the git diff hunk cache.")
	config.AutoIndexingTaskInterval = config.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL", "10m", "The frequency with which to run periodic codeintel auto-indexing tasks.")
	config.AutoIndexingSkipManualInterval = config.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_SKIP_MANUAL", "24h", "The duration the auto-indexer will wait after a manual upload to a repository before it starts auto-indexing again. Manually queueing an auto-index run will cancel this waiting period.")
	config.IndexBatchSize = config.GetInt("PRECISE_CODE_INTEL_INDEX_BATCH_SIZE", "100", "The number of indexable repositories to schedule at a time.")
	config.MinimumTimeSinceLastEnqueue = config.GetInterval("PRECISE_CODE_INTEL_MINIMUM_TIME_SINCE_LAST_ENQUEUE", "24h", "The minimum time between auto-index enqueues for the same repository.")
	config.MinimumSearchCount = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_COUNT", "50", "The minimum number of search-based code intel events that triggers auto-indexing on a repository.")
	config.MinimumSearchRatio = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_RATIO", "50", "The minimum ratio of search-based to total code intel events that triggers auto-indexing on a repository.")
	config.MinimumPreciseCount = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_PRECISE_COUNT", "1", "The minimum number of precise code intel events that triggers auto-indexing on a repository.")
	config.DiagnosticsCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DIAGNOSTICS_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of document records to migrate at a time.")
	config.DiagnosticsCountMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_DIAGNOSTICS_COUNT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.DefinitionsCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DEFINITIONS_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of definition records to migrate at once.")
	config.DefinitionsCountMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_DEFINITIONS_COUNT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.ReferencesCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_REFERENCES_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of reference records to migrate at a time.")
	config.ReferencesCountMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_REFERENCES_COUNT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.DocumentColumnSplitMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DOCUMENT_COLUMN_SPLIT_MIGRATION_BATCH_SIZE", "100", "The maximum number of document records to migrate at a time.")
	config.DocumentColumnSplitMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_DOCUMENT_COLUMN_SPLIT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.CommittedAtMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_COMMITTED_AT_MIGRATION_BATCH_SIZE", "100", "The maximum number of upload records to migrate at a time.")
	config.CommittedAtMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_COMMITTED_AT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.DependencyIndexerSchedulerPollInterval = config.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL", "1s", "Interval between queries to the dependency indexing job queue.")
	config.DependencyIndexerSchedulerConcurrency = config.GetInt("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY", "1", "The maximum number of dependency graphs that can be processed concurrently.")
}

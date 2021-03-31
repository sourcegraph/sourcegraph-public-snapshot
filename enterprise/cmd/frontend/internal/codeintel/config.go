package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	UploadStoreConfig                  *uploadstore.Config
	CommitGraphUpdateTaskInterval      time.Duration
	CleanupTaskInterval                time.Duration
	AutoIndexingTaskInterval           time.Duration
	AutoIndexingSkipManualInterval     time.Duration
	HunkCacheSize                      int
	DataTTL                            time.Duration
	UploadTimeout                      time.Duration
	IndexBatchSize                     int
	MinimumTimeSinceLastEnqueue        time.Duration
	MinimumSearchCount                 int
	MinimumSearchRatio                 int
	MinimumPreciseCount                int
	DiagnosticsCountMigrationBatchSize int
}

var config = &Config{}

func init() {
	uploadStoreConfig := &uploadstore.Config{}
	uploadStoreConfig.Load()
	config.UploadStoreConfig = uploadStoreConfig

	config.HunkCacheSize = config.GetInt("PRECISE_CODE_INTEL_HUNK_CACHE_SIZE", "1000", "The capacity of the git diff hunk cache.")
	config.DataTTL = config.GetInterval("PRECISE_CODE_INTEL_DATA_TTL", "720h", "The maximum time an non-critical index can live in the database.")
	config.UploadTimeout = config.GetInterval("PRECISE_CODE_INTEL_UPLOAD_TIMEOUT", "24h", "The maximum time an upload can be in the 'uploading' state.")
	config.CommitGraphUpdateTaskInterval = config.GetInterval("PRECISE_CODE_INTEL_COMMIT_GRAPH_UPDATE_TASK_INTERVAL", "10s", "The frequency with which to run periodic codeintel commit graph update tasks.")
	config.CleanupTaskInterval = config.GetInterval("PRECISE_CODE_INTEL_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic codeintel cleanup tasks.")
	config.AutoIndexingTaskInterval = config.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL", "10m", "The frequency with which to run periodic codeintel auto-indexing tasks.")
	config.AutoIndexingSkipManualInterval = config.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_SKIP_MANUAL", "24h", "The duration the auto-indexer will wait after a manual upload to a repository before it starts auto-indexing again. Manually queueing an auto-index run will cancel this waiting period.")
	config.IndexBatchSize = config.GetInt("PRECISE_CODE_INTEL_INDEX_BATCH_SIZE", "100", "The number of indexable repositories to schedule at a time.")
	config.MinimumTimeSinceLastEnqueue = config.GetInterval("PRECISE_CODE_INTEL_MINIMUM_TIME_SINCE_LAST_ENQUEUE", "24h", "The minimum time between auto-index enqueues for the same repository.")
	config.MinimumSearchCount = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_COUNT", "50", "The minimum number of search-based code intel events that triggers auto-indexing on a repository.")
	config.MinimumSearchRatio = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_RATIO", "50", "The minimum ratio of search-based to total code intel events that triggers auto-indexing on a repository.")
	config.MinimumPreciseCount = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_PRECISE_COUNT", "1", "The minimum number of precise code intel events that triggers auto-indexing on a repository.")
	config.DiagnosticsCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DIAGNOSTICS_COUNT_MIGRATION_BATCH_SIZE", "1000", "The number of document records to migrate per second.")
}

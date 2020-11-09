package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	UploadStoreConfig *uploadstore.Config

	HunkCacheSize               int
	BackgroundTaskInterval      time.Duration
	IndexBatchSize              int
	MinimumTimeSinceLastEnqueue time.Duration
	MinimumSearchCount          int
	MinimumSearchRatio          int
	MinimumPreciseCount         int
	UploadTimeout               time.Duration
	DataTTL                     time.Duration
}

var config = &Config{}

func init() {
	uploadStoreConfig := &uploadstore.Config{}
	uploadStoreConfig.Load()
	config.UploadStoreConfig = uploadStoreConfig

	config.HunkCacheSize = config.GetInt("PRECISE_CODE_INTEL_HUNK_CACHE_SIZE", "1000", "The capacity of the git diff hunk cache.")
	config.BackgroundTaskInterval = config.GetInterval("PRECISE_CODE_INTEL_BACKGROUND_TASK_INTERVAL", "1m", "The frequency with which to run periodic codeintel background tasks.")
	config.IndexBatchSize = config.GetInt("PRECISE_CODE_INTEL_INDEX_BATCH_SIZE", "100", "The number of indexable repositories to schedule at a time.")
	config.MinimumTimeSinceLastEnqueue = config.GetInterval("PRECISE_CODE_INTEL_MINIMUM_TIME_SINCE_LAST_ENQUEUE", "24h", "The minimum time between auto-index enqueues for the same repository.")
	config.MinimumSearchCount = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_COUNT", "50", "The minimum number of search-based code intel events that triggers auto-indexing on a repository.")
	config.MinimumSearchRatio = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_RATIO", "50", "The minimum ratio of search-based to total code intel events that triggers auto-indexing on a repository.")
	config.MinimumPreciseCount = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_PRECISE_COUNT", "1", "The minimum number of precise code intel events that triggers auto-indexing on a repository.")
	config.UploadTimeout = config.GetInterval("PRECISE_CODE_INTEL_UPLOAD_TIMEOUT", "24h", "The maximum time an upload can be in the 'uploading' state.")
	config.DataTTL = config.GetInterval("PRECISE_CODE_INTEL_DATA_TTL", "720h", "The maximum time an non-critical index can live in the database.")
}

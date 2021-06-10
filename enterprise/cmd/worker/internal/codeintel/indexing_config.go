package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type indexingConfig struct {
	env.BaseConfig

	AutoIndexingTaskInterval               time.Duration
	AutoIndexingSkipManualInterval         time.Duration
	IndexBatchSize                         int
	MinimumTimeSinceLastEnqueue            time.Duration
	MinimumSearchCount                     int
	MinimumSearchRatio                     int
	MinimumPreciseCount                    int
	DependencyIndexerSchedulerPollInterval time.Duration
	DependencyIndexerSchedulerConcurrency  int
}

var indexingConfigInst = &indexingConfig{}

func (c *indexingConfig) Load() {
	c.AutoIndexingTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL", "10m", "The frequency with which to run periodic codeintel auto-indexing tasks.")
	c.AutoIndexingSkipManualInterval = c.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_SKIP_MANUAL", "24h", "The duration the auto-indexer will wait after a manual upload to a repository before it starts auto-indexing again. Manually queueing an auto-index run will cancel this waiting period.")
	c.IndexBatchSize = c.GetInt("PRECISE_CODE_INTEL_INDEX_BATCH_SIZE", "100", "The number of indexable repositories to schedule at a time.")
	c.MinimumTimeSinceLastEnqueue = c.GetInterval("PRECISE_CODE_INTEL_MINIMUM_TIME_SINCE_LAST_ENQUEUE", "24h", "The minimum time between auto-index enqueues for the same repository.")
	c.MinimumSearchCount = c.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_COUNT", "50", "The minimum number of search-based code intel events that triggers auto-indexing on a repository.")
	c.MinimumSearchRatio = c.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_RATIO", "50", "The minimum ratio of search-based to total code intel events that triggers auto-indexing on a repository.")
	c.MinimumPreciseCount = c.GetInt("PRECISE_CODE_INTEL_MINIMUM_PRECISE_COUNT", "1", "The minimum number of precise code intel events that triggers auto-indexing on a repository.")
	c.DependencyIndexerSchedulerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL", "1s", "Interval between queries to the dependency indexing job queue.")
	c.DependencyIndexerSchedulerConcurrency = c.GetInt("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY", "1", "The maximum number of dependency graphs that can be processed concurrently.")
}

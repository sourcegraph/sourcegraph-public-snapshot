package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type indexingConfig struct {
	env.BaseConfig

	AutoIndexingTaskInterval               time.Duration
	DependencyIndexerSchedulerPollInterval time.Duration
	DependencyIndexerSchedulerConcurrency  int
}

var indexingConfigInst = &indexingConfig{}

func (c *indexingConfig) Load() {
	c.AutoIndexingTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL", "10m", "The frequency with which to run periodic codeintel auto-indexing tasks.")
	c.DependencyIndexerSchedulerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL", "1s", "Interval between queries to the dependency indexing job queue.")
	c.DependencyIndexerSchedulerConcurrency = c.GetInt("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY", "1", "The maximum number of dependency graphs that can be processed concurrently.")
}

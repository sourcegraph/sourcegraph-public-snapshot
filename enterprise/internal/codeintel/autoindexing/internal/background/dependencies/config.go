package dependencies

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	ResetterInterval                       time.Duration
	DependencySyncSchedulerPollInterval    time.Duration
	DependencyIndexerSchedulerPollInterval time.Duration
	DependencyIndexerSchedulerConcurrency  int
}

func (c *Config) Load() {
	c.ResetterInterval = c.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_RESETTER_INTERVAL", "30s", "Interval between dependency sync and index resets.")
	c.DependencySyncSchedulerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_SYNC_SCHEDULER_POLL_INTERVAL", "1s", "Interval between queries to the dependency syncing job queue.")
	c.DependencyIndexerSchedulerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL", "1s", "Interval between queries to the dependency indexing job queue.")
	c.DependencyIndexerSchedulerConcurrency = c.GetInt("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY", "1", "The maximum number of dependency graphs that can be processed concurrently.")
}

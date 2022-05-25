package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type indexingConfig struct {
	env.BaseConfig

	DependencyIndexerSchedulerPollInterval time.Duration
	DependencyIndexerSchedulerConcurrency  int
}

var indexingConfigInst = &indexingConfig{}

func (c *indexingConfig) Load() {
	c.DependencyIndexerSchedulerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL", "1s", "Interval between queries to the dependency indexing job queue.")
	c.DependencyIndexerSchedulerConcurrency = c.GetInt("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY", "1", "The maximum number of dependency graphs that can be processed concurrently.")
}

func (c *indexingConfig) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	return errs
}

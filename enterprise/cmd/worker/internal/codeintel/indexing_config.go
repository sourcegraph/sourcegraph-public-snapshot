package codeintel

import (
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type indexingConfig struct {
	env.BaseConfig

	AutoIndexEnqueuerConfig                *enqueuer.Config
	AutoIndexingTaskInterval               time.Duration
	RepositoryProcessDelay                 time.Duration
	RepositoryBatchSize                    int
	DependencyIndexerSchedulerPollInterval time.Duration
	DependencyIndexerSchedulerConcurrency  int
}

var indexingConfigInst = &indexingConfig{}

func (c *indexingConfig) Load() {
	enqueuerConfig := &enqueuer.Config{}
	enqueuerConfig.Load()
	indexingConfigInst.AutoIndexEnqueuerConfig = enqueuerConfig

	c.AutoIndexingTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL", "10m", "The frequency with which to run periodic codeintel auto-indexing tasks.")
	c.RepositoryProcessDelay = c.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY", "24h", "The minimum frequency that the same repository can be considered for auto-index scheduling.")
	c.RepositoryBatchSize = c.GetInt("PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_BATCH_SIZE", "100", "The number of repositories to consider for auto-indexing scheduling at a time.")
	c.DependencyIndexerSchedulerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL", "1s", "Interval between queries to the dependency indexing job queue.")
	c.DependencyIndexerSchedulerConcurrency = c.GetInt("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY", "1", "The maximum number of dependency graphs that can be processed concurrently.")
}

func (c *janitorConfig) Validate() error {
	var errs *multierror.Error
	errs = multierror.Append(errs, c.BaseConfig.Validate())
	errs = multierror.Append(errs, c.MetricsConfig.Validate())
	return errs.ErrorOrNil()
}

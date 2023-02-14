package autoindexing

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type cleanupJobsConfig struct {
	env.BaseConfig

	Interval                       time.Duration
	MinimumTimeSinceLastCheck      time.Duration
	CommitResolverBatchSize        int
	CommitResolverMaximumCommitLag time.Duration
	FailedIndexBatchSize           int
	FailedIndexMaxAge              time.Duration
}

var ConfigCleanupInst = &cleanupJobsConfig{}

func (c *cleanupJobsConfig) Load() {
	minimumTimeSinceLastCheckName := env.ChooseFallbackVariableName("CODEINTEL_AUTOINDEXING_MINIMUM_TIME_SINCE_LAST_CHECK", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_MINIMUM_TIME_SINCE_LAST_CHECK")
	commitResolverBatchSizeName := env.ChooseFallbackVariableName("CODEINTEL_AUTOINDEXING_COMMIT_RESOLVER_BATCH_SIZE", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_BATCH_SIZE")
	commitResolverMaximumCommitLagName := env.ChooseFallbackVariableName("CODEINTEL_AUTOINDEXING_COMMIT_RESOLVER_MAXIMUM_COMMIT_LAG", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_MAXIMUM_COMMIT_LAG")

	c.Interval = c.GetInterval("CODEINTEL_AUTOINDEXING_CLEANUP_INTERVAL", "1m", "How frequently to run the autoindexing janitor routine.")
	c.MinimumTimeSinceLastCheck = c.GetInterval(minimumTimeSinceLastCheckName, "24h", "The minimum time the commit resolver will re-check an upload or index record.")
	c.CommitResolverBatchSize = c.GetInt(commitResolverBatchSizeName, "100", "The maximum number of unique commits to resolve at a time.")
	c.CommitResolverMaximumCommitLag = c.GetInterval(commitResolverMaximumCommitLagName, "0s", "The maximum acceptable delay between accepting an upload and its commit becoming resolvable. Be cautious about setting this to a large value, as uploads for unresolvable commits will be retried periodically during this interval.")
	c.FailedIndexBatchSize = c.GetInt("CODEINTEL_AUTOINDEXING_FAILED_INDEX_BATCH_SIZE", "1000", "The number of old, failed index records to delete at once.")
	c.FailedIndexMaxAge = c.GetInterval("CODEINTEL_AUTOINDEXING_FAILED_INDEX_MAX_AGE", "2190h", "The maximum age a non-relevant failed index record will remain queryable.")
}

type dependencyIndexJobsConfig struct {
	env.BaseConfig

	DependencyIndexerSchedulerPollInterval time.Duration
	DependencyIndexerSchedulerConcurrency  int
}

var ConfigDependencyIndexInst = &dependencyIndexJobsConfig{}

func (c *dependencyIndexJobsConfig) Load() {
	c.DependencyIndexerSchedulerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL", "1s", "Interval between queries to the dependency indexing job queue.")
	c.DependencyIndexerSchedulerConcurrency = c.GetInt("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY", "1", "The maximum number of dependency graphs that can be processed concurrently.")
}

type indexJobsConfig struct {
	env.BaseConfig

	SchedulerInterval      time.Duration
	RepositoryProcessDelay time.Duration
	RepositoryBatchSize    int
	PolicyBatchSize        int
	InferenceConcurrency   int

	OnDemandSchedulerInterval time.Duration
	OnDemandBatchsize         int
}

var ConfigIndexingInst = &indexJobsConfig{}

func (c *indexJobsConfig) Load() {
	intervalName := env.ChooseFallbackVariableName("CODEINTEL_AUTOINDEXING_SCHEDULER_INTERVAL", "PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL")
	repositoryProcessDelayName := env.ChooseFallbackVariableName("CODEINTEL_AUTOINDEXING_SCHEDULER_REPOSITORY_PROCESS_DELAY", "PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY")
	repositoryBatchSizeName := env.ChooseFallbackVariableName("CODEINTEL_AUTOINDEXING_SCHEDULER_REPOSITORY_BATCH_SIZE", "PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_BATCH_SIZE")
	policyBatchSizeName := env.ChooseFallbackVariableName("CODEINTEL_AUTOINDEXING_SCHEDULER_POLICY_BATCH_SIZE", "PRECISE_CODE_INTEL_AUTO_INDEXING_POLICY_BATCH_SIZE")

	c.SchedulerInterval = c.GetInterval(intervalName, "2m", "How frequently to run the auto-indexing scheduling routine.")
	c.RepositoryProcessDelay = c.GetInterval(repositoryProcessDelayName, "24h", "The minimum frequency that the same repository can be considered for auto-index scheduling.")
	c.RepositoryBatchSize = c.GetInt(repositoryBatchSizeName, "2500", "The number of repositories to consider for auto-indexing scheduling at a time.")
	c.PolicyBatchSize = c.GetInt(policyBatchSizeName, "100", "The number of policies to consider for auto-indexing scheduling at a time.")
	c.InferenceConcurrency = c.GetInt("CODEINTEL_AUTOINDEXING_INFERENCE_CONCURRENCY", "16", "The number of inference jobs running in parallel in the background scheduler.")

	c.OnDemandSchedulerInterval = c.GetInterval("CODEINTEL_AUTOINDEXING_ON_DEMAND_SCHEDULER_INTERVAL", "30s", "How frequently to run the on-demand auto-indexing scheduling routine.")
	c.OnDemandBatchsize = c.GetInt("CODEINTEL_AUTOINDEXING_ON_DEMAND_SCHEDULER_BATCH_SIZE", "100", "The number of repo/rev pairs to consider for on-demand auto-indexing scheduling at a time.")
}

type summaryBuilderConfig struct {
	env.BaseConfig

	Interval time.Duration
}

var SummaryBuilderConfigInst = &summaryBuilderConfig{}

func (c *summaryBuilderConfig) Load() {
	// TODO - lower this pre-merge.
	c.Interval = c.GetInterval("CODEINTEL_AUTOINDEXING_SUMMARY_BUILDER_INTERVAL", "1s", "How frequently to run the auto-indexing summary builder routine.")
}

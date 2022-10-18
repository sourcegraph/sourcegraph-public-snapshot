package cleanup

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	Interval                       time.Duration
	MinimumTimeSinceLastCheck      time.Duration
	CommitResolverBatchSize        int
	CommitResolverMaximumCommitLag time.Duration
	FailedIndexBatchSize           int
	FailedIndexMaxAge              time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
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

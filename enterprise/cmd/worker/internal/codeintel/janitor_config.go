package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type janitorConfig struct {
	env.BaseConfig

	UploadTimeout                                       time.Duration
	CleanupTaskInterval                                 time.Duration
	CommitResolverTaskInterval                          time.Duration
	CommitResolverMinimumTimeSinceLastCheck             time.Duration
	CommitResolverBatchSize                             int
	RepositoryProcessDelay                              time.Duration
	RepositoryBatchSize                                 int
	UploadProcessDelay                                  time.Duration
	UploadBatchSize                                     int
	CommitBatchSize                                     int
	BranchesCacheMaxKeys                                int
	DocumentationSearchCurrentMinimumTimeSinceLastCheck time.Duration
	DocumentationSearchCurrentBatchSize                 int

	MetricsConfig *executorqueue.Config
}

var janitorConfigInst = &janitorConfig{}

func (c *janitorConfig) Load() {
	c.UploadTimeout = c.GetInterval("PRECISE_CODE_INTEL_UPLOAD_TIMEOUT", "24h", "The maximum time an upload can be in the 'uploading' state.")
	c.CleanupTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic codeintel cleanup tasks.")
	c.CommitResolverTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_RESOLVER_TASK_INTERVAL", "10s", "The frequency with which to run the periodic commit resolver task.")
	c.CommitResolverMinimumTimeSinceLastCheck = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_RESOLVER_MINIMUM_TIME_SINCE_LAST_CHECK", "24h", "The minimum time the commit resolver will re-check an upload or index record.")
	c.CommitResolverBatchSize = c.GetInt("PRECISE_CODE_INTEL_COMMIT_RESOLVER_BATCH_SIZE", "100", "The maximum number of unique commits to resolve at a time.")
	c.RepositoryProcessDelay = c.GetInterval("PRECISE_CODE_INTEL_RETENTION_REPOSITORY_PROCESS_DELAY", "24h", "The minimum frequency that the same repository's uploads can be considered for expiration.")
	c.RepositoryBatchSize = c.GetInt("PRECISE_CODE_INTEL_RETENTION_REPOSITORY_BATCH_SIZE", "100", "The number of repositories to consider for expiration at a time.")
	c.UploadProcessDelay = c.GetInterval("PRECISE_CODE_INTEL_RETENTION_UPLOAD_PROCESS_DELAY", "24h", "The minimum frequency that the same upload record can be considered for expiration.")
	c.UploadBatchSize = c.GetInt("PRECISE_CODE_INTEL_RETENTION_UPLOAD_BATCH_SIZE", "100", "The number of uploads to consider for expiration at a time.")
	c.CommitBatchSize = c.GetInt("PRECISE_CODE_INTEL_RETENTION_COMMIT_BATCH_SIZE", "100", "The number of commits to process per upload at a time.")
	c.BranchesCacheMaxKeys = c.GetInt("PRECISE_CODE_INTEL_RETENTION_BRANCHES_CACHE_MAX_KEYS", "10000", "The number of maximum keys used to cache the set of branches visible from a commit.")
	c.DocumentationSearchCurrentMinimumTimeSinceLastCheck = c.GetInterval("PRECISE_CODE_INTEL_DOCUMENTATION_SEARCH_CURRENT_MINIMUM_TIME_SINCE_LAST_CHECK", "24h", "The minimum time the documentation search current janitor will re-check records for a unique search key.")
	c.DocumentationSearchCurrentBatchSize = c.GetInt("PRECISE_CODE_INTEL_DOCUMENTATION_SEARCH_CURRENT_BATCH_SIZE", "100", "The maximum number of unique search keys to clean up at a time.")

	c.MetricsConfig = executorqueue.InitMetricsConfig()
	c.MetricsConfig.Load()
}

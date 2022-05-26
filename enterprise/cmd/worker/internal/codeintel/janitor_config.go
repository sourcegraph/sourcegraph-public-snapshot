package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type janitorConfig struct {
	env.BaseConfig

	UploadTimeout                          time.Duration
	CleanupTaskInterval                    time.Duration
	CommitResolverTaskInterval             time.Duration
	RepositoryProcessDelay                 time.Duration
	RepositoryBatchSize                    int
	UploadProcessDelay                     time.Duration
	UploadBatchSize                        int
	PolicyBatchSize                        int
	CommitBatchSize                        int
	BranchesCacheMaxKeys                   int
	ConfigurationPolicyMembershipBatchSize int

	MetricsConfig *executorqueue.Config
}

var janitorConfigInst = &janitorConfig{}

func (c *janitorConfig) Load() {
	metricsConfig := executorqueue.InitMetricsConfig()
	metricsConfig.Load()
	c.MetricsConfig = metricsConfig

	c.CleanupTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic codeintel cleanup tasks.")
	c.CommitResolverTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_RESOLVER_TASK_INTERVAL", "10s", "The frequency with which to run the periodic commit resolver task.")
	c.RepositoryProcessDelay = c.GetInterval("PRECISE_CODE_INTEL_RETENTION_REPOSITORY_PROCESS_DELAY", "24h", "The minimum frequency that the same repository's uploads can be considered for expiration.")
	c.RepositoryBatchSize = c.GetInt("PRECISE_CODE_INTEL_RETENTION_REPOSITORY_BATCH_SIZE", "100", "The number of repositories to consider for expiration at a time.")
	c.UploadProcessDelay = c.GetInterval("PRECISE_CODE_INTEL_RETENTION_UPLOAD_PROCESS_DELAY", "24h", "The minimum frequency that the same upload record can be considered for expiration.")
	c.UploadBatchSize = c.GetInt("PRECISE_CODE_INTEL_RETENTION_UPLOAD_BATCH_SIZE", "100", "The number of uploads to consider for expiration at a time.")
	c.PolicyBatchSize = c.GetInt("PRECISE_CODE_INTEL_RETENTION_POLICY_BATCH_SIZE", "100", "The number of policies to consider for expiration at a time.")
	c.CommitBatchSize = c.GetInt("PRECISE_CODE_INTEL_RETENTION_COMMIT_BATCH_SIZE", "100", "The number of commits to process per upload at a time.")
	c.BranchesCacheMaxKeys = c.GetInt("PRECISE_CODE_INTEL_RETENTION_BRANCHES_CACHE_MAX_KEYS", "10000", "The number of maximum keys used to cache the set of branches visible from a commit.")
	c.ConfigurationPolicyMembershipBatchSize = c.GetInt("PRECISE_CODE_INTEL_CONFIGURATION_POLICY_MEMBERSHIP_BATCH_SIZE", "100", "The maximum number of policy configurations to update repository membership for at a time.")
}

func (c *janitorConfig) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.MetricsConfig.Validate())
	return errs
}

package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type janitorConfig struct {
	env.BaseConfig

	DataTTL                                 time.Duration
	UploadTimeout                           time.Duration
	CleanupTaskInterval                     time.Duration
	CommitResolverTaskInterval              time.Duration
	CommitResolverMinimumTimeSinceLastCheck time.Duration
	CommitResolverBatchSize                 int
}

var janitorConfigInst = &janitorConfig{}

func (c *janitorConfig) Load() {
	c.DataTTL = c.GetInterval("PRECISE_CODE_INTEL_DATA_TTL", "720h", "The maximum time an non-critical index can live in the database.")
	c.UploadTimeout = c.GetInterval("PRECISE_CODE_INTEL_UPLOAD_TIMEOUT", "24h", "The maximum time an upload can be in the 'uploading' state.")
	c.CleanupTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic codeintel cleanup tasks.")
	c.CommitResolverTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_RESOLVER_TASK_INTERVAL", "10s", "The frequency with which to run the periodic commit resolver task.")
	c.CommitResolverMinimumTimeSinceLastCheck = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_RESOLVER_MINIMUM_TIME_SINCE_LAST_CHECK", "24h", "The minimum time the commit resolver will re-check an upload or index record.")
	c.CommitResolverBatchSize = c.GetInt("PRECISE_CODE_INTEL_COMMIT_RESOLVER_BATCH_SIZE", "100", "The maximum number of unique commits to resolve at a time.")
}

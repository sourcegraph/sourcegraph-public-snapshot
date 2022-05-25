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
	AuditLogMaxAge                 time.Duration
	CommitResolverMaximumCommitLag time.Duration
	UploadTimeout                  time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_UPLOAD_JANITOR_INTERVAL", "1s", "How frequently to run the updater janitor routine.")
	c.MinimumTimeSinceLastCheck = c.GetInterval("CODEINTEL_UPLOAD_MINIMUM_TIME_SINCE_LAST_CHECK", "24h", "The minimum time the commit resolver will re-check an upload or index record.")
	c.CommitResolverBatchSize = c.GetInt("CODEINTEL_UPLOAD_COMMIT_RESOLVER_BATCH_SIZE", "100", "The maximum number of unique commits to resolve at a time.")
	c.AuditLogMaxAge = c.GetInterval("CODEINTEL_UPLOAD_AUDIT_LOG_MAX_AGE", "720h", "The maximum time a code intel audit log record can remain on the database.")
	// TODO(numbers): Update name to CODEINTEL_UPLOAD_COMMIT_RESOLVER_MAXIMUM_COMMIT_LAG. Need to give customer notice first.
	c.CommitResolverMaximumCommitLag = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_RESOLVER_MAXIMUM_COMMIT_LAG", "0s", "The maximum acceptable delay between accepting an upload and its commit becoming resolvable. Be cautious about setting this to a large value, as uploads for unresolvable commits will be retried periodically during this interval.")
	c.UploadTimeout = c.GetInterval("PRECISE_CODE_INTEL_UPLOAD_TIMEOUT", "24h", "The maximum time an upload can be in the 'uploading' state.")
}

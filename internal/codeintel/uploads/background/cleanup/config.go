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
	minimumTimeSinceLastCheckName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_MINIMUM_TIME_SINCE_LAST_CHECK", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_MINIMUM_TIME_SINCE_LAST_CHECK")
	commitResolverBatchSizeName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_COMMIT_RESOLVER_BATCH_SIZE", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_BATCH_SIZE")
	auditLogMaxAgeName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_AUDIT_LOG_MAX_AGE", "PRECISE_CODE_INTEL_AUDIT_LOG_MAX_AGE")
	commitResolverMaximumCommitLagName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_COMMIT_RESOLVER_MAXIMUM_COMMIT_LAG", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_MAXIMUM_COMMIT_LAG")
	uploadTimeoutName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_UPLOAD_TIMEOUT", "PRECISE_CODE_INTEL_UPLOAD_TIMEOUT")

	c.Interval = c.GetInterval("CODEINTEL_UPLOADS_CLEANUP_INTERVAL", "1m", "How frequently to run the updater janitor routine.")
	c.MinimumTimeSinceLastCheck = c.GetInterval(minimumTimeSinceLastCheckName, "24h", "The minimum time the commit resolver will re-check an upload or index record.")
	c.CommitResolverBatchSize = c.GetInt(commitResolverBatchSizeName, "100", "The maximum number of unique commits to resolve at a time.")
	c.AuditLogMaxAge = c.GetInterval(auditLogMaxAgeName, "720h", "The maximum time a code intel audit log record can remain on the database.")
	c.CommitResolverMaximumCommitLag = c.GetInterval(commitResolverMaximumCommitLagName, "0s", "The maximum acceptable delay between accepting an upload and its commit becoming resolvable. Be cautious about setting this to a large value, as uploads for unresolvable commits will be retried periodically during this interval.")
	c.UploadTimeout = c.GetInterval(uploadTimeoutName, "24h", "The maximum time an upload can be in the 'uploading' state.")
}

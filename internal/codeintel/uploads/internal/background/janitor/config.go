package janitor

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval                        time.Duration
	AbandonedSchemaVersionsInterval time.Duration
	MinimumTimeSinceLastCheck       time.Duration
	CommitResolverBatchSize         int
	AuditLogMaxAge                  time.Duration
	UnreferencedDocumentBatchSize   int
	UnreferencedDocumentMaxAge      time.Duration
	CommitResolverMaximumCommitLag  time.Duration
	UploadTimeout                   time.Duration
	ReconcilerBatchSize             int
	FailedIndexBatchSize            int
	FailedIndexMaxAge               time.Duration
}

func (c *Config) Load() {
	minimumTimeSinceLastCheckName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_MINIMUM_TIME_SINCE_LAST_CHECK", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_MINIMUM_TIME_SINCE_LAST_CHECK")
	commitResolverBatchSizeName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_COMMIT_RESOLVER_BATCH_SIZE", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_BATCH_SIZE")
	auditLogMaxAgeName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_AUDIT_LOG_MAX_AGE", "PRECISE_CODE_INTEL_AUDIT_LOG_MAX_AGE")
	commitResolverMaximumCommitLagName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_COMMIT_RESOLVER_MAXIMUM_COMMIT_LAG", "PRECISE_CODE_INTEL_COMMIT_RESOLVER_MAXIMUM_COMMIT_LAG")
	uploadTimeoutName := env.ChooseFallbackVariableName("CODEINTEL_UPLOADS_UPLOAD_TIMEOUT", "PRECISE_CODE_INTEL_UPLOAD_TIMEOUT")

	c.Interval = c.GetInterval("CODEINTEL_UPLOADS_CLEANUP_INTERVAL", "1m", "How frequently to run the updater janitor routine.")
	c.AbandonedSchemaVersionsInterval = c.GetInterval("CODEINTEL_UPLOADS_ABANDONED_SCHEMA_VERSIONS_CLEANUP_INTERVAL", "24h", "How frequently to run the query to clean up *_schema_version records that are not tracked by foreign key.")
	c.MinimumTimeSinceLastCheck = c.GetInterval(minimumTimeSinceLastCheckName, "24h", "The minimum time the commit resolver will re-check an upload or index record.")
	c.CommitResolverBatchSize = c.GetInt(commitResolverBatchSizeName, "100", "The maximum number of unique commits to resolve at a time.")
	c.AuditLogMaxAge = c.GetInterval(auditLogMaxAgeName, "720h", "The maximum time a code intel audit log record can remain on the database.")
	c.UnreferencedDocumentBatchSize = c.GetInt("CODEINTEL_UPLOADS_UNREFERENCED_DOCUMENT_BATCH_SIZE", "100", "The number of unreferenced SCIP documents to consider for deletion at a time.")
	c.UnreferencedDocumentMaxAge = c.GetInterval("CODEINTEL_UPLOADS_UNREFERENCED_DOCUMENT_MAX_AGE", "24h", "The maximum time an unreferenced SCIP document should remain in the database.")
	c.CommitResolverMaximumCommitLag = c.GetInterval(commitResolverMaximumCommitLagName, "0s", "The maximum acceptable delay between accepting an upload and its commit becoming resolvable. Be cautious about setting this to a large value, as uploads for unresolvable commits will be retried periodically during this interval.")
	c.UploadTimeout = c.GetInterval(uploadTimeoutName, "24h", "The maximum time an upload can be in the 'uploading' state.")
	c.ReconcilerBatchSize = c.GetInt("CODEINTEL_UPLOADS_RECONCILER_BATCH_SIZE", "1000", "The number of uploads to reconcile in one cleanup routine invocation.")
	c.FailedIndexBatchSize = c.GetInt("CODEINTEL_AUTOINDEXING_FAILED_INDEX_BATCH_SIZE", "1000", "The number of old, failed index records to delete at once.")
	c.FailedIndexMaxAge = c.GetInterval("CODEINTEL_AUTOINDEXING_FAILED_INDEX_MAX_AGE", "730h", "The maximum age a non-relevant failed index record will remain queryable.")
}

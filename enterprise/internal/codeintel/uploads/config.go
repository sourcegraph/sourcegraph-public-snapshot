package uploads

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type backfillConfig struct {
	env.BaseConfig

	Interval  time.Duration
	BatchSize int
}

var ConfigCommittedAtBackfillInst = &backfillConfig{}

func (c *backfillConfig) Load() {
	c.Interval = c.GetInterval("CODEINTEL_UPLOAD_BACKFILLER_INTERVAL", "10s", "The frequency with which to run periodic codeintel backfiller tasks.")
	c.BatchSize = c.GetInt("CODEINTEL_UPLOAD_BACKFILLER_BATCH_SIZE", "100", "The number of upload to populate an unset `commited_at` field per batch.")
}

type janitorConfig struct {
	env.BaseConfig

	Interval                       time.Duration
	MinimumTimeSinceLastCheck      time.Duration
	CommitResolverBatchSize        int
	AuditLogMaxAge                 time.Duration
	CommitResolverMaximumCommitLag time.Duration
	UploadTimeout                  time.Duration
	ReconcilerBatchSize            int
}

var ConfigJanitorInst = &janitorConfig{}

func (c *janitorConfig) Load() {
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
	c.ReconcilerBatchSize = c.GetInt("CODEINTEL_UPLOADS_RECONCILER_BATCH_SIZE", "1000", "The number of uploads to reconcile in one cleanup routine invocation.")
}

type commitGraphConfig struct {
	env.BaseConfig

	Interval                      time.Duration
	MaxAgeForNonStaleBranches     time.Duration
	MaxAgeForNonStaleTags         time.Duration
	CommitGraphUpdateTaskInterval time.Duration
}

var ConfigCommitGraphInst = &commitGraphConfig{}

func (c *commitGraphConfig) Load() {
	maxAgeForNonStaleBranches := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_COMMITGRAPH_MAX_AGE_FOR_NON_STALE_BRANCHES", "PRECISE_CODE_INTEL_MAX_AGE_FOR_NON_STALE_BRANCHES")
	maxAgeForNonStaleTags := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_COMMITGRAPH_MAX_AGE_FOR_NON_STALE_TAGS", "PRECISE_CODE_INTEL_MAX_AGE_FOR_NON_STALE_TAGS")
	commitGraphUpdateTaskInterval := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_COMMITGRAPH_UPDATE_TASK_INTERVAL", "PRECISE_CODE_INTEL_COMMIT_GRAPH_UPDATE_TASK_INTERVAL")

	c.Interval = c.GetInterval("CODEINTEL_UPLOAD_COMMITGRAPH_UPDATER_INTERVAL", "1s", "How frequently to run the upload commitgraph updater routine.")
	c.MaxAgeForNonStaleBranches = c.GetInterval(maxAgeForNonStaleBranches, "2160h", "The age after which a branch should be considered stale. Code intelligence indexes will be evicted from stale branches.")      // about 3 months
	c.MaxAgeForNonStaleTags = c.GetInterval(maxAgeForNonStaleTags, "8760h", "The age after which a tagged commit should be considered stale. Code intelligence indexes will be evicted from stale tagged commits.") // about 1 year
	c.CommitGraphUpdateTaskInterval = c.GetInterval(commitGraphUpdateTaskInterval, "10s", "The frequency with which to run periodic codeintel commit graph update tasks.")
}

type expirationConfig struct {
	env.BaseConfig

	CommitBatchSize        int
	ExpirerInterval        time.Duration
	PolicyBatchSize        int
	RepositoryBatchSize    int
	RepositoryProcessDelay time.Duration
	UploadBatchSize        int
	UploadProcessDelay     time.Duration
}

var ConfigExpirationInst = &expirationConfig{}

func (c *expirationConfig) Load() {
	commitBatchSize := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_COMMIT_BATCH_SIZE", "PRECISE_CODE_INTEL_RETENTION_COMMIT_BATCH_SIZE")
	policyBatchSize := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_POLICY_BATCH_SIZE", "PRECISE_CODE_INTEL_RETENTION_POLICY_BATCH_SIZE")
	repositoryBatchSize := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_REPOSITORY_BATCH_SIZE", "PRECISE_CODE_INTEL_RETENTION_REPOSITORY_BATCH_SIZE")
	repositoryProcessDelay := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_REPOSITORY_PROCESS_DELAY", "PRECISE_CODE_INTEL_RETENTION_REPOSITORY_PROCESS_DELAY")
	uploadBatchSize := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_UPLOAD_BATCH_SIZE", "PRECISE_CODE_INTEL_RETENTION_UPLOAD_BATCH_SIZE")
	uploadProcessDelay := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_UPLOAD_PROCESS_DELAY", "PRECISE_CODE_INTEL_RETENTION_UPLOAD_PROCESS_DELAY")

	c.CommitBatchSize = c.GetInt(commitBatchSize, "100", "The number of commits to process per upload at a time.")
	c.ExpirerInterval = c.GetInterval("CODEINTEL_UPLOAD_EXPIRER_INTERVAL", "1s", "How frequently to run the upload expirer routine.")
	c.PolicyBatchSize = c.GetInt(policyBatchSize, "100", "The number of policies to consider for expiration at a time.")
	c.RepositoryBatchSize = c.GetInt(repositoryBatchSize, "100", "The number of repositories to consider for expiration at a time.")
	c.RepositoryProcessDelay = c.GetInterval(repositoryProcessDelay, "24h", "The minimum frequency that the same repository's uploads can be considered for expiration.")
	c.UploadBatchSize = c.GetInt(uploadBatchSize, "100", "The number of uploads to consider for expiration at a time.")
	c.UploadProcessDelay = c.GetInterval(uploadProcessDelay, "24h", "The minimum frequency that the same upload record can be considered for expiration.")
}

type exportConfig struct {
	env.BaseConfig

	RankingInterval    time.Duration
	NumRankingRoutines int
}

var ConfigExportInst = &exportConfig{}

func (c *exportConfig) Load() {
	c.RankingInterval = c.GetInterval("CODEINTEL_UPLOADS_RANKING_INTERVAL", "1s", "How frequently to serialize a batch of the code intel graph for ranking.")
	c.NumRankingRoutines = c.GetInt("CODEINTEL_UPLOADS_RANKING_NUM_ROUTINES", "4", "The number of concurrent ranking graph serializer routines to run per worker instance.")
}

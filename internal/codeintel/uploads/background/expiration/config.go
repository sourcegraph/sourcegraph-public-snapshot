package expiration

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	BranchesCacheMaxKeys   int
	CleanupTaskInterval    time.Duration
	CommitBatchSize        int
	Interval               time.Duration
	PolicyBatchSize        int
	RepositoryBatchSize    int
	RepositoryProcessDelay time.Duration
	UploadBatchSize        int
	UploadProcessDelay     time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.BranchesCacheMaxKeys = c.GetInt("CODEINTEL_UPLOAD_EXPIRER_BRANCHES_CACHE_MAX_KEYS", "10000", "The number of maximum keys used to cache the set of branches visible from a commit.")
	c.CleanupTaskInterval = c.GetInterval("CODEINTEL_UPLOAD_EXPIRER_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic codeintel cleanup tasks.")
	c.CommitBatchSize = c.GetInt("CODEINTEL_UPLOAD_EXPIRER_COMMIT_BATCH_SIZE", "100", "The number of commits to process per upload at a time.")
	c.Interval = c.GetInterval("CODEINTEL_UPLOAD_EXPIRER_INTERVAL", "1s", "How frequently to run the upload expirer routine.")
	c.PolicyBatchSize = c.GetInt("CODEINTEL_UPLOAD_EXPIRER_POLICY_BATCH_SIZE", "100", "The number of policies to consider for expiration at a time.")
	c.RepositoryBatchSize = c.GetInt("CODEINTEL_UPLOAD_EXPIRER_REPOSITORY_BATCH_SIZE", "100", "The number of repositories to consider for expiration at a time.")
	c.RepositoryProcessDelay = c.GetInterval("CODEINTEL_UPLOAD_EXPIRER_REPOSITORY_PROCESS_DELAY", "24h", "The minimum frequency that the same repository's uploads can be considered for expiration.")
	c.UploadBatchSize = c.GetInt("CODEINTEL_UPLOAD_EXPIRER_UPLOAD_BATCH_SIZE", "100", "The number of uploads to consider for expiration at a time.")
	c.UploadProcessDelay = c.GetInterval("CODEINTEL_UPLOAD_EXPIRER_UPLOAD_PROCESS_DELAY", "24h", "The minimum frequency that the same upload record can be considered for expiration.")
}

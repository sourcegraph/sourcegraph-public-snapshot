package expirer

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	CommitBatchSize        int
	ExpirerInterval        time.Duration
	PolicyBatchSize        int
	RepositoryBatchSize    int
	RepositoryProcessDelay time.Duration
	UploadBatchSize        int
	UploadProcessDelay     time.Duration
}

func (c *Config) Load() {
	commitBatchSize := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_COMMIT_BATCH_SIZE", "PRECISE_CODE_INTEL_RETENTION_COMMIT_BATCH_SIZE")
	policyBatchSize := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_POLICY_BATCH_SIZE", "PRECISE_CODE_INTEL_RETENTION_POLICY_BATCH_SIZE")
	repositoryBatchSize := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_REPOSITORY_BATCH_SIZE", "PRECISE_CODE_INTEL_RETENTION_REPOSITORY_BATCH_SIZE")
	repositoryProcessDelay := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_REPOSITORY_PROCESS_DELAY", "PRECISE_CODE_INTEL_RETENTION_REPOSITORY_PROCESS_DELAY")
	uploadBatchSize := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_UPLOAD_BATCH_SIZE", "PRECISE_CODE_INTEL_RETENTION_UPLOAD_BATCH_SIZE")
	uploadProcessDelay := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_EXPIRER_UPLOAD_PROCESS_DELAY", "PRECISE_CODE_INTEL_RETENTION_UPLOAD_PROCESS_DELAY")

	c.CommitBatchSize = c.GetInt(commitBatchSize, "100", "The number of commits to process per upload at a time.")
	c.ExpirerInterval = c.GetInterval("CODEINTEL_UPLOAD_EXPIRER_INTERVAL", "30s", "How frequently to run the upload expirer routine.")
	c.PolicyBatchSize = c.GetInt(policyBatchSize, "100", "The number of policies to consider for expiration at a time.")
	c.RepositoryBatchSize = c.GetInt(repositoryBatchSize, "100", "The number of repositories to consider for expiration at a time.")
	c.RepositoryProcessDelay = c.GetInterval(repositoryProcessDelay, "24h", "The minimum frequency that the same repository's uploads can be considered for expiration.")
	c.UploadBatchSize = c.GetInt(uploadBatchSize, "100", "The number of uploads to consider for expiration at a time.")
	c.UploadProcessDelay = c.GetInterval(uploadProcessDelay, "24h", "The minimum frequency that the same upload record can be considered for expiration.")
}

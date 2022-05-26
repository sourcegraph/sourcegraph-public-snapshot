package scheduler

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	Interval               time.Duration
	RepositoryProcessDelay time.Duration
	RepositoryBatchSize    int
	PolicyBatchSize        int
}

var ConfigInst = &config{}

func (c *config) Load() {
	intervalName := env.ChooseFallbackVariableName(
		"CODEINTEL_AUTOINDEXING_SCHEDULER_INTERVAL",
		"PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL",
	)

	c.Interval = c.GetInterval(intervalName, "10m", "How frequently to run the autoindexer scheduling routine.")
	c.RepositoryProcessDelay = c.GetInterval("PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY", "24h", "The minimum frequency that the same repository can be considered for auto-index scheduling.")
	c.RepositoryBatchSize = c.GetInt("PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_BATCH_SIZE", "100", "The number of repositories to consider for auto-indexing scheduling at a time.")
	c.PolicyBatchSize = c.GetInt("PRECISE_CODE_INTEL_AUTO_INDEXING_POLICY_BATCH_SIZE", "100", "The number of policies to consider for auto-indexing scheduling at a time.")
}

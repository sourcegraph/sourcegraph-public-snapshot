package perforce

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval            time.Duration
	RepositoryBatchSize int
}

func (c *Config) Load() {
	c.Interval = c.GetInterval("SRC_PERFORCE_CHANGELIST_MAPPER_INTERVAL", "1m", "How frequently to run the Perforce changelist mapper routine.")
	c.RepositoryBatchSize = c.GetInt("SRC_PERFORCE_CHANGELIST_MAPPER_REPO_BATCH_SIZE", "100", "The number of repositories to load at once when processing for Perforce changelist mappings.")
}

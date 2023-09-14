package mapper

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval  time.Duration
	BatchSize int
}

func (c *Config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_RANKING_MAPPER_INTERVAL", "1s", "How frequently to run the ranking mapper.")
	c.BatchSize = c.GetInt("CODEINTEL_RANKING_MAPPER_BATCH_SIZE", "100", "How many definitions and references to map at once.")
}

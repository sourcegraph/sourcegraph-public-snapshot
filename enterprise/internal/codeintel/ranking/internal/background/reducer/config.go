package reducer

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
	c.Interval = c.GetInterval("CODEINTEL_RANKING_REDUCER_INTERVAL", "1s", "How frequently to run the ranking reducer.")
	c.BatchSize = c.GetInt("CODEINTEL_RANKING_REDUCER_BATCH_SIZE", "1000", "How many path counts to reduce at once.")
}

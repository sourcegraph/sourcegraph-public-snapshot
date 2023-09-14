package matcher

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	MatcherInterval time.Duration
	BatchSize       int
}

func (c *Config) Load() {
	c.MatcherInterval = c.GetInterval("CODEINTEL_SENTINEL_MATCHER_INTERVAL", "1s", "How frequently to match existing records against known vulnerabilities.")
	c.BatchSize = c.GetInt("CODEINTEL_SENTINEL_BATCH_SIZE", "100", "How many precise indexes to scan at once for vulnerabilities.")
}

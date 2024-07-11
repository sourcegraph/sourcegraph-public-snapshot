package janitor

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval time.Duration
}

func (c *Config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_RANKING_JANITOR_INTERVAL", "30s", "How frequently to run the ranking janitor.")
}

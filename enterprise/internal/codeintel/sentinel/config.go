package sentinel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type sentinelConfig struct {
	env.BaseConfig

	Interval time.Duration
}

var ConfigInst = &sentinelConfig{}

func (c *sentinelConfig) Load() {
	c.Interval = c.GetInterval("CODEINTEL_SENTINEL_INTERVAL", "1s", "How frequently to run background jobs.")
}

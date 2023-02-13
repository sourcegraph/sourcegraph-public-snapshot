package sentinel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type sentinelConfig struct {
	env.BaseConfig

	Interval time.Duration // TODO
}

var ConfigInst = &sentinelConfig{}

func (c *sentinelConfig) Load() {
	c.Interval = c.GetInterval("SENTINEL_INTERVAL", "1s", "TODO") // TODO
}

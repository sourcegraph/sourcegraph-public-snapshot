package cleanup

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	Interval time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_AUTOINDEXING_CLEANUP_INTERVAL", "1m", "How frequently to run the autoindexing janitor routine.")
}

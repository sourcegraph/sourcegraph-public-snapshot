package indexer

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
	c.Interval = c.GetInterval("CODEINTEL_RANKING_INDEXER_INTERVAL", "10s", "The frequency with which to run periodic codeintel rank indexing tasks.")
}

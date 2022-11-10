package loader

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	LoadInterval  time.Duration
	MergeInterval time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.LoadInterval = c.GetInterval("CODEINTEL_RANKING_LOADER_INTERVAL", "10s", "The frequency with which to run periodic codeintel rank loading tasks.")
	c.MergeInterval = c.GetInterval("CODEINTEL_RANKING_MERGER_INTERVAL", "1s", "The frequency with which to run periodic codeintel rank merging tasks.")
}

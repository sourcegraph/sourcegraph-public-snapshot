package resolver

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	Interval             time.Duration
	BatchSize            int
	MinimumCheckInterval time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_DEPENDENCIES_RESOLVER_INTERVAL", "1s", "How frequently to run the dependencies resolver routine.")
	c.BatchSize = c.GetInt("CODEINTEL_DEPENDENCIES_LOCKFILES_RESOLUTION_BATCH_SIZE", "100", "How many repository/commit pairs to resolve at a time.")
	c.MinimumCheckInterval = c.GetInterval("CODEINTEL_DEPENDENCIES_LOCKFILES_RESOLUTION_MINIMUM_CHECK_INTERVAL", "24h", "How frequently to re-resolve the same repository/commit pair.")
}

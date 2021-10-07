package config

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

// SharedConfig defines common items that are used by multiple queues.
type SharedConfig struct {
	env.BaseConfig

	FrontendPassword string
}

func (c *SharedConfig) Load() {
	c.FrontendPassword = c.GetOptional("EXECUTOR_FRONTEND_PASSWORD", "The password supplied to the frontend.")
}

package config

import (
	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

// SharedConfig defines common items that are used by multiple queues.
type SharedConfig struct {
	env.BaseConfig

	FrontendUsername string
	FrontendPassword string
}

func (c *SharedConfig) Load() {
	c.FrontendUsername = c.Get("EXECUTOR_FRONTEND_USERNAME", "", "The username supplied to the frontend.")
	c.FrontendPassword = c.Get("EXECUTOR_FRONTEND_PASSWORD", "", "The password supplied to the frontend.")
}

func (c *SharedConfig) Validate() error {
	if c.FrontendUsername == "" {
		return errors.Errorf("invalid value for EXECUTOR_FRONTEND_USERNAME: no value supplied")
	}
	if c.FrontendPassword == "" {
		return errors.Errorf("invalid value for EXECUTOR_FRONTEND_PASSWORD: no value supplied")
	}
	return c.BaseConfig.Validate()
}

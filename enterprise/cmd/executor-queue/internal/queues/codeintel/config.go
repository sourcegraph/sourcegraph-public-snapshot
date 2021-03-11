package codeintel

import "github.com/sourcegraph/sourcegraph/internal/env"

type Config struct {
	env.BaseConfig

	FrontendURL      string
	FrontendUsername string
	FrontendPassword string
}

func (c *Config) Load() {
	c.FrontendURL = c.Get("EXECUTOR_FRONTEND_URL", "", "The external URL of the sourcegraph instance.")
	c.FrontendUsername = c.Get("EXECUTOR_FRONTEND_USERNAME", "", "The username supplied to the frontend.")
	c.FrontendPassword = c.Get("EXECUTOR_FRONTEND_PASSWORD", "", "The password supplied to the frontend.")
}

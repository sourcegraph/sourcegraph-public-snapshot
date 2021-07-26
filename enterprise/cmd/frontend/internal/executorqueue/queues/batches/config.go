package batches

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/config"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Shared *config.SharedConfig
}

func (c *Config) Load() {}

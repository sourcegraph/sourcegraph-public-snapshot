package batches

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type janitorConfig struct {
	env.BaseConfig

	MetricsConfig *executorqueue.Config
}

var janitorConfigInst = &janitorConfig{}

func (c *janitorConfig) Load() {
	c.MetricsConfig = executorqueue.InitMetricsConfig()
	c.MetricsConfig.Load()
}

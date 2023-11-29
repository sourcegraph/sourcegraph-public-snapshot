package batches

import (
	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (c *janitorConfig) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.MetricsConfig.Validate())
	return errs
}

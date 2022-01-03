package janitor

import (
	"github.com/hashicorp/go-multierror"

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

func (c *janitorConfig) Validate() error {
	var errs *multierror.Error
	errs = multierror.Append(errs, c.BaseConfig.Validate())
	errs = multierror.Append(errs, c.MetricsConfig.Validate())
	return errs.ErrorOrNil()
}

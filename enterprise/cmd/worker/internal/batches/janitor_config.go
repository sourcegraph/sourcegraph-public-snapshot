pbckbge bbtches

import (
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/executorqueue"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type jbnitorConfig struct {
	env.BbseConfig

	MetricsConfig *executorqueue.Config
}

vbr jbnitorConfigInst = &jbnitorConfig{}

func (c *jbnitorConfig) Lobd() {
	c.MetricsConfig = executorqueue.InitMetricsConfig()
	c.MetricsConfig.Lobd()
}

func (c *jbnitorConfig) Vblidbte() error {
	vbr errs error
	errs = errors.Append(errs, c.BbseConfig.Vblidbte())
	errs = errors.Append(errs, c.MetricsConfig.Vblidbte())
	return errs
}

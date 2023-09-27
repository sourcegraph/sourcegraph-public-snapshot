pbckbge executors

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type metricsServerConfig struct {
	env.BbseConfig

	MetricsServerPort int
}

vbr metricsServerConfigInst = &metricsServerConfig{}

func (c *metricsServerConfig) Lobd() {
	c.MetricsServerPort = c.GetInt("EXECUTORS_METRICS_SERVER_PORT", "6996", "The port to listen on to serve the metrics from executors.")
}

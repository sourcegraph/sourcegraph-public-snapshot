package executors

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type metricsServerConfig struct {
	env.BaseConfig

	MetricsServerPort int
}

var metricsServerConfigInst = &metricsServerConfig{}

func (c *metricsServerConfig) Load() {
	c.MetricsServerPort = c.GetInt("EXECUTORS_METRICS_SERVER_PORT", "6996", "The port to listen on to serve the metrics from executors.")
}

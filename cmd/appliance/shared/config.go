package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	metrics metricsConfig
	grpc    grpcConfig
}

func (c *Config) Load() {
	c.metrics.addr = ":8080"
	c.metrics.secure = false
	c.grpc.addr = ":9000"
}

func (c *Config) Validate() error {
	var errs error
	return errs
}

type metricsConfig struct {
	addr   string
	secure bool
}

type grpcConfig struct {
	addr string
}

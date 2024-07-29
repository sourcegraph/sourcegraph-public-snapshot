package shared

import (
	"net"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

func LoadConfig() *Config {
	var config Config
	config.Load()
	return &config
}

type Config struct {
	env.BaseConfig

	ListenAddress                   string
	ReconcilerInterval              time.Duration
	ExhaustiveRequestLoggingEnabled bool
}

func (c *Config) Load() {
	c.ListenAddress = c.GetOptional("TENANTMANAGER_ADDR", "The address under which the tenantmanager API listens. Can include a port.")
	// Fall back to a reasonable default.
	if c.ListenAddress == "" {
		port := "3187"
		host := ""
		if env.InsecureDev {
			host = "127.0.0.1"
		}
		c.ListenAddress = net.JoinHostPort(host, port)
	}

	c.ReconcilerInterval = c.GetInterval("SRC_TENANT_RECONCILER_INTERVAL", "1m", "Interval between reconciler runs")
	c.ExhaustiveRequestLoggingEnabled = c.GetBool("SRC_TENANTMANAGER_EXHAUSTIVE_LOGGING_ENABLED", "false", "Enable exhaustive request logging in tenantmanager")
}

package processor

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval time.Duration

	// Note: set by pci-worker initialization
	WorkerConcurrency    int
	WorkerBudget         int64
	WorkerPollInterval   time.Duration
	MaximumRuntimePerJob time.Duration
}

func (c *Config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_UPLOAD_RESETTER_INTERVAL", "5s", "The frequency to reset lost upload jobs.")
}

package processor

import (
	"time"
)

type Config struct {
	// Note: set by pci-worker initialization
	WorkerConcurrency    int
	WorkerBudget         int64
	WorkerPollInterval   time.Duration
	MaximumRuntimePerJob time.Duration
}

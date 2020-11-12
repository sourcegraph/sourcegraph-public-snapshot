package main

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	UploadStoreConfig  *uploadstore.Config
	WorkerPollInterval time.Duration
	WorkerConcurrency  int
	WorkerBudget       int64
}

func (c *Config) Load() {
	uploadStoreConfig := &uploadstore.Config{}
	uploadStoreConfig.Load()
	c.UploadStoreConfig = uploadStoreConfig

	c.WorkerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_WORKER_POLL_INTERVAL", "1s", "Interval between queries to the upload queue.")
	c.WorkerConcurrency = c.GetInt("PRECISE_CODE_INTEL_WORKER_CONCURRENCY", "1", "The maximum number of indexes that can be processed concurrently.")
	c.WorkerBudget = int64(c.GetInt("PRECISE_CODE_INTEL_WORKER_BUDGET", "0", "The amount of compressed input data (in bytes) a worker can process concurrently. Zero acts as an infinite budget."))
}

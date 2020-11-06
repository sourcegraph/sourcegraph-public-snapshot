package main

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

const Port = 3188

type Config struct {
	env.BaseConfig

	UploadStoreConfig *uploadstore.Config

	BundleManagerURL      string
	WorkerPollInterval    time.Duration
	WorkerConcurrency     int
	WorkerBudget          int64
	ResetInterval         time.Duration
	CommitUpdaterInterval time.Duration
}

func (c *Config) Load() {
	uploadStoreConfig := &uploadstore.Config{}
	uploadStoreConfig.Load()
	c.UploadStoreConfig = uploadStoreConfig

	c.BundleManagerURL = c.Get("PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL", "", "HTTP address for internal LSIF bundle manager server.")
	c.WorkerPollInterval = c.GetInterval("PRECISE_CODE_INTEL_WORKER_POLL_INTERVAL", "1s", "Interval between queries to the upload queue.")
	c.WorkerConcurrency = c.GetInt("PRECISE_CODE_INTEL_WORKER_CONCURRENCY", "1", "The maximum number of indexes that can be processed concurrently.")
	c.WorkerBudget = int64(c.GetInt("PRECISE_CODE_INTEL_WORKER_BUDGET", "0", "The amount of compressed input data (in bytes) a worker can process concurrently. Zero acts as an infinite budget."))
	c.ResetInterval = c.GetInterval("PRECISE_CODE_INTEL_RESET_INTERVAL", "1m", "How often to reset stalled uploads.")
	c.CommitUpdaterInterval = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_UPDATER_INTERVAL", "5s", "How often to update commits for dirty repositories.")
}

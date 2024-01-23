package shared

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	env.BaseConfig

	WorkerPollInterval    time.Duration
	WorkerConcurrency     int
	WorkerBudget          int64
	MaximumRuntimePerJob  time.Duration
	SCIPUploadStoreConfig *lsifuploadstore.Config
	CliPath               string
}

func (c *Config) Load() {
	c.SCIPUploadStoreConfig = &lsifuploadstore.Config{}
	c.SCIPUploadStoreConfig.Load()

	c.WorkerPollInterval = c.GetInterval("SYNTACTIC_CODE_INTEL_WORKER_POLL_INTERVAL", "1s", "Interval between queries to the upload queue.")
	c.WorkerConcurrency = c.GetInt("SYNTACTIC_CODE_INTEL_WORKER_CONCURRENCY", "1", "The maximum number of indexes that can be processed concurrently.")
	c.WorkerBudget = int64(c.GetInt("SYNTACTIC_CODE_INTEL_WORKER_BUDGET", "0", "The amount of compressed input data (in bytes) a worker can process concurrently. Zero acts as an infinite budget."))
	c.MaximumRuntimePerJob = c.GetInterval("SYNTACTIC_CODE_INTEL_WORKER_MAXIMUM_RUNTIME_PER_JOB", "25m", "The maximum time a single LSIF processing job can take.")
	c.CliPath = c.Get("SCIP_TREESITTER_COMMAND", "scip-treesitter", "TODO: fill in description")
}

func (c *Config) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.SCIPUploadStoreConfig.Validate())
	return errs
}

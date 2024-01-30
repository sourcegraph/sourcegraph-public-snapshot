package shared

import (
	"net"
	"strconv"
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
	ListenAddress         string
}

const DefaultPort = 3188

func (c *Config) Load() {
	c.SCIPUploadStoreConfig = &lsifuploadstore.Config{}
	c.SCIPUploadStoreConfig.Load()

	c.WorkerPollInterval = c.GetInterval("SYNTACTIC_CODE_INTEL_WORKER_POLL_INTERVAL", "1s", "Interval between queries to the repository queue")
	c.WorkerConcurrency = c.GetInt("SYNTACTIC_CODE_INTEL_WORKER_CONCURRENCY", "1", "The maximum number of repositories that can be processed concurrently.")
	c.WorkerBudget = int64(c.GetInt("SYNTACTIC_CODE_INTEL_WORKER_BUDGET", "0", "The amount of compressed input data (in bytes) a worker can process concurrently. Zero acts as an infinite budget."))
	c.MaximumRuntimePerJob = c.GetInterval("SYNTACTIC_CODE_INTEL_WORKER_MAXIMUM_RUNTIME_PER_JOB", "25m", "The maximum time a single repository indexing job can take")

	c.CliPath = c.Get("SCIP_TREESITTER_PATH", "scip-treesitter", "TODO: fill in description")

	c.ListenAddress = c.GetOptional("SYNTACTIC_CODE_INTEL_WORKER_ADDR", "The address under which the syntactic codeintel worker API listens. Can include a port.")
	// Fall back to a reasonable default.
	if c.ListenAddress == "" {
		port := strconv.Itoa(DefaultPort)
		host := ""
		if env.InsecureDev {
			host = "127.0.0.1"
		}
		c.ListenAddress = net.JoinHostPort(host, port)
	}
}

func (c *Config) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.SCIPUploadStoreConfig.Validate())
	return errs
}

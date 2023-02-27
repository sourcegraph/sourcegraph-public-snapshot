package permissions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type config struct {
	env.BaseConfig

	WorkerPollInterval  time.Duration
	WorkerConcurrency   int
	WorkerRetryInterval time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.WorkerPollInterval = c.GetInterval("BITBUCKET_PROJECT_PERMISSIONS_WORKER_POLL_INTERVAL", "1s", "How frequently to query the job queue")
	c.WorkerConcurrency = c.GetInt("BITBUCKET_PROJECT_PERMISSIONS_WORKER_CONCURRENCY", "1", "The maximum number of projects that can be processed concurrently")
	c.WorkerRetryInterval = c.GetInterval("BITBUCKET_PROJECT_PERMISSIONS_WORKER_RETRY_INTERVAL", "30s", "The minimum number of time to wait before retrying a failed job")
}

func (c *config) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	if c.WorkerPollInterval < 0 {
		errs = errors.Append(errs, errors.New("BITBUCKET_PROJECT_PERMISSIONS_WORKER_POLL_INTERVAL must be greater than or equal to 0"))
	}
	if c.WorkerConcurrency < 1 {
		errs = errors.Append(errs, errors.New("BITBUCKET_PROJECT_PERMISSIONS_WORKER_CONCURRENCY must be greater than 0"))
	}

	return errs
}

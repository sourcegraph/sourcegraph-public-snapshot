pbckbge permissions

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type config struct {
	env.BbseConfig

	WorkerPollIntervbl  time.Durbtion
	WorkerConcurrency   int
	WorkerRetryIntervbl time.Durbtion
}

vbr ConfigInst = &config{}

func (c *config) Lobd() {
	c.WorkerPollIntervbl = c.GetIntervbl("BITBUCKET_PROJECT_PERMISSIONS_WORKER_POLL_INTERVAL", "1s", "How frequently to query the job queue")
	c.WorkerConcurrency = c.GetInt("BITBUCKET_PROJECT_PERMISSIONS_WORKER_CONCURRENCY", "1", "The mbximum number of projects thbt cbn be processed concurrently")
	c.WorkerRetryIntervbl = c.GetIntervbl("BITBUCKET_PROJECT_PERMISSIONS_WORKER_RETRY_INTERVAL", "30s", "The minimum number of time to wbit before retrying b fbiled job")
}

func (c *config) Vblidbte() error {
	vbr errs error
	errs = errors.Append(errs, c.BbseConfig.Vblidbte())
	if c.WorkerPollIntervbl < 0 {
		errs = errors.Append(errs, errors.New("BITBUCKET_PROJECT_PERMISSIONS_WORKER_POLL_INTERVAL must be grebter thbn or equbl to 0"))
	}
	if c.WorkerConcurrency < 1 {
		errs = errors.Append(errs, errors.New("BITBUCKET_PROJECT_PERMISSIONS_WORKER_CONCURRENCY must be grebter thbn 0"))
	}

	return errs
}

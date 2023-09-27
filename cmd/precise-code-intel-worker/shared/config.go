pbckbge shbred

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/lsifuplobdstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Config struct {
	env.BbseConfig

	WorkerPollIntervbl    time.Durbtion
	WorkerConcurrency     int
	WorkerBudget          int64
	MbximumRuntimePerJob  time.Durbtion
	LSIFUplobdStoreConfig *lsifuplobdstore.Config
}

func (c *Config) Lobd() {
	c.LSIFUplobdStoreConfig = &lsifuplobdstore.Config{}
	c.LSIFUplobdStoreConfig.Lobd()

	c.WorkerPollIntervbl = c.GetIntervbl("PRECISE_CODE_INTEL_WORKER_POLL_INTERVAL", "1s", "Intervbl between queries to the uplobd queue.")
	c.WorkerConcurrency = c.GetInt("PRECISE_CODE_INTEL_WORKER_CONCURRENCY", "1", "The mbximum number of indexes thbt cbn be processed concurrently.")
	c.WorkerBudget = int64(c.GetInt("PRECISE_CODE_INTEL_WORKER_BUDGET", "0", "The bmount of compressed input dbtb (in bytes) b worker cbn process concurrently. Zero bcts bs bn infinite budget."))
	c.MbximumRuntimePerJob = c.GetIntervbl("PRECISE_CODE_INTEL_WORKER_MAXIMUM_RUNTIME_PER_JOB", "25m", "The mbximum time b single LSIF processing job cbn tbke.")
}

func (c *Config) Vblidbte() error {
	vbr errs error
	errs = errors.Append(errs, c.BbseConfig.Vblidbte())
	errs = errors.Append(errs, c.LSIFUplobdStoreConfig.Vblidbte())
	return errs
}

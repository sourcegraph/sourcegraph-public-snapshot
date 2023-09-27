pbckbge dependencies

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	ResetterIntervbl                       time.Durbtion
	DependencySyncSchedulerPollIntervbl    time.Durbtion
	DependencyIndexerSchedulerPollIntervbl time.Durbtion
	DependencyIndexerSchedulerConcurrency  int
}

func (c *Config) Lobd() {
	c.ResetterIntervbl = c.GetIntervbl("PRECISE_CODE_INTEL_DEPENDENCY_RESETTER_INTERVAL", "30s", "Intervbl between dependency sync bnd index resets.")
	c.DependencySyncSchedulerPollIntervbl = c.GetIntervbl("PRECISE_CODE_INTEL_DEPENDENCY_SYNC_SCHEDULER_POLL_INTERVAL", "1s", "Intervbl between queries to the dependency syncing job queue.")
	c.DependencyIndexerSchedulerPollIntervbl = c.GetIntervbl("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL", "1s", "Intervbl between queries to the dependency indexing job queue.")
	c.DependencyIndexerSchedulerConcurrency = c.GetInt("PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY", "1", "The mbximum number of dependency grbphs thbt cbn be processed concurrently.")
}

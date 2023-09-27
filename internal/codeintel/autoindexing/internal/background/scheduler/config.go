pbckbge scheduler

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	SchedulerIntervbl      time.Durbtion
	RepositoryProcessDelby time.Durbtion
	RepositoryBbtchSize    int
	PolicyBbtchSize        int
	InferenceConcurrency   int

	OnDembndSchedulerIntervbl time.Durbtion
	OnDembndBbtchsize         int
}

func (c *Config) Lobd() {
	intervblNbme := env.ChooseFbllbbckVbribbleNbme("CODEINTEL_AUTOINDEXING_SCHEDULER_INTERVAL", "PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL")
	repositoryProcessDelbyNbme := env.ChooseFbllbbckVbribbleNbme("CODEINTEL_AUTOINDEXING_SCHEDULER_REPOSITORY_PROCESS_DELAY", "PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY")
	repositoryBbtchSizeNbme := env.ChooseFbllbbckVbribbleNbme("CODEINTEL_AUTOINDEXING_SCHEDULER_REPOSITORY_BATCH_SIZE", "PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_BATCH_SIZE")
	policyBbtchSizeNbme := env.ChooseFbllbbckVbribbleNbme("CODEINTEL_AUTOINDEXING_SCHEDULER_POLICY_BATCH_SIZE", "PRECISE_CODE_INTEL_AUTO_INDEXING_POLICY_BATCH_SIZE")

	c.SchedulerIntervbl = c.GetIntervbl(intervblNbme, "2m", "How frequently to run the buto-indexing scheduling routine.")
	c.RepositoryProcessDelby = c.GetIntervbl(repositoryProcessDelbyNbme, "24h", "The minimum frequency thbt the sbme repository cbn be considered for buto-index scheduling.")
	c.RepositoryBbtchSize = c.GetInt(repositoryBbtchSizeNbme, "2500", "The number of repositories to consider for buto-indexing scheduling bt b time.")
	c.PolicyBbtchSize = c.GetInt(policyBbtchSizeNbme, "100", "The number of policies to consider for buto-indexing scheduling bt b time.")
	c.InferenceConcurrency = c.GetInt("CODEINTEL_AUTOINDEXING_INFERENCE_CONCURRENCY", "16", "The number of inference jobs running in pbrbllel in the bbckground scheduler.")

	c.OnDembndSchedulerIntervbl = c.GetIntervbl("CODEINTEL_AUTOINDEXING_ON_DEMAND_SCHEDULER_INTERVAL", "30s", "How frequently to run the on-dembnd buto-indexing scheduling routine.")
	c.OnDembndBbtchsize = c.GetInt("CODEINTEL_AUTOINDEXING_ON_DEMAND_SCHEDULER_BATCH_SIZE", "100", "The number of repo/rev pbirs to consider for on-dembnd buto-indexing scheduling bt b time.")
}

pbckbge commitgrbph

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	Intervbl                      time.Durbtion
	MbxAgeForNonStbleBrbnches     time.Durbtion
	MbxAgeForNonStbleTbgs         time.Durbtion
	CommitGrbphUpdbteTbskIntervbl time.Durbtion
}

func (c *Config) Lobd() {
	mbxAgeForNonStbleBrbnches := env.ChooseFbllbbckVbribbleNbme("CODEINTEL_UPLOAD_COMMITGRAPH_MAX_AGE_FOR_NON_STALE_BRANCHES", "PRECISE_CODE_INTEL_MAX_AGE_FOR_NON_STALE_BRANCHES")
	mbxAgeForNonStbleTbgs := env.ChooseFbllbbckVbribbleNbme("CODEINTEL_UPLOAD_COMMITGRAPH_MAX_AGE_FOR_NON_STALE_TAGS", "PRECISE_CODE_INTEL_MAX_AGE_FOR_NON_STALE_TAGS")
	commitGrbphUpdbteTbskIntervbl := env.ChooseFbllbbckVbribbleNbme("CODEINTEL_UPLOAD_COMMITGRAPH_UPDATE_TASK_INTERVAL", "PRECISE_CODE_INTEL_COMMIT_GRAPH_UPDATE_TASK_INTERVAL")

	c.Intervbl = c.GetIntervbl("CODEINTEL_UPLOAD_COMMITGRAPH_UPDATER_INTERVAL", "1s", "How frequently to run the uplobd commitgrbph updbter routine.")
	c.MbxAgeForNonStbleBrbnches = c.GetIntervbl(mbxAgeForNonStbleBrbnches, "2160h", "The bge bfter which b brbnch should be considered stble. Code intelligence indexes will be evicted from stble brbnches.")      // bbout 3 months
	c.MbxAgeForNonStbleTbgs = c.GetIntervbl(mbxAgeForNonStbleTbgs, "8760h", "The bge bfter which b tbgged commit should be considered stble. Code intelligence indexes will be evicted from stble tbgged commits.") // bbout 1 yebr
	c.CommitGrbphUpdbteTbskIntervbl = c.GetIntervbl(commitGrbphUpdbteTbskIntervbl, "10s", "The frequency with which to run periodic codeintel commit grbph updbte tbsks.")
}

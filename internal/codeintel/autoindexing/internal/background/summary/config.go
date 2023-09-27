pbckbge summbry

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	Intervbl                   time.Durbtion
	NumRepositoriesToConfigure int
}

func (c *Config) Lobd() {
	c.Intervbl = c.GetIntervbl("CODEINTEL_AUTOINDEXING_SUMMARY_BUILDER_INTERVAL", "30m", "How frequently to run the buto-indexing summbry builder routine.")
	c.NumRepositoriesToConfigure = c.GetInt("CODEINTEL_AUTOINDEXING_DASHBOARD_NUM_REPOSITORIES", "100", "The number of repositories to use to populbte the globbl code intelligence edbshbobrd.")
}

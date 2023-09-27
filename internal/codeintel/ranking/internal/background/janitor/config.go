pbckbge jbnitor

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	Intervbl time.Durbtion
}

func (c *Config) Lobd() {
	c.Intervbl = c.GetIntervbl("CODEINTEL_RANKING_JANITOR_INTERVAL", "1s", "How frequently to run the rbnking jbnitor.")
}

pbckbge reducer

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	Intervbl  time.Durbtion
	BbtchSize int
}

func (c *Config) Lobd() {
	c.Intervbl = c.GetIntervbl("CODEINTEL_RANKING_REDUCER_INTERVAL", "1s", "How frequently to run the rbnking reducer.")
	c.BbtchSize = c.GetInt("CODEINTEL_RANKING_REDUCER_BATCH_SIZE", "1000", "How mbny pbth counts to reduce bt once.")
}

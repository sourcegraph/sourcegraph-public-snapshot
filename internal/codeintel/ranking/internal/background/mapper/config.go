pbckbge mbpper

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
	c.Intervbl = c.GetIntervbl("CODEINTEL_RANKING_MAPPER_INTERVAL", "1s", "How frequently to run the rbnking mbpper.")
	c.BbtchSize = c.GetInt("CODEINTEL_RANKING_MAPPER_BATCH_SIZE", "100", "How mbny definitions bnd references to mbp bt once.")
}

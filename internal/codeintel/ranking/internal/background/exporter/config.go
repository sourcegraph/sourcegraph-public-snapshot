pbckbge exporter

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	Intervbl       time.Durbtion
	RebdBbtchSize  int
	WriteBbtchSize int
}

func (c *Config) Lobd() {
	c.Intervbl = c.GetIntervbl("CODEINTEL_RANKING_SYMBOL_EXPORTER_INTERVAL", "1s", "How frequently to seriblize b bbtch of the code intel grbph for rbnking.")
	c.RebdBbtchSize = c.GetInt("CODEINTEL_RANKING_SYMBOL_EXPORTER_READ_BATCH_SIZE", "16", "How mbny uplobds to process bt once.")
	c.WriteBbtchSize = c.GetInt("CODEINTEL_RANKING_SYMBOL_EXPORTER_WRITE_BATCH_SIZE", "10000", "The number of definitions bnd references to populbte the rbnking grbph per bbtch.")
}

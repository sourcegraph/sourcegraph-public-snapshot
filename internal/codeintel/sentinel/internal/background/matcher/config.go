pbckbge mbtcher

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	MbtcherIntervbl time.Durbtion
	BbtchSize       int
}

func (c *Config) Lobd() {
	c.MbtcherIntervbl = c.GetIntervbl("CODEINTEL_SENTINEL_MATCHER_INTERVAL", "1s", "How frequently to mbtch existing records bgbinst known vulnerbbilities.")
	c.BbtchSize = c.GetInt("CODEINTEL_SENTINEL_BATCH_SIZE", "100", "How mbny precise indexes to scbn bt once for vulnerbbilities.")
}

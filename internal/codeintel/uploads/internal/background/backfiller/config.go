pbckbge bbckfiller

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
	c.Intervbl = c.GetIntervbl("CODEINTEL_UPLOAD_BACKFILLER_INTERVAL", "10s", "The frequency with which to run periodic codeintel bbckfiller tbsks.")
	c.BbtchSize = c.GetInt("CODEINTEL_UPLOAD_BACKFILLER_BATCH_SIZE", "100", "The number of uplobd to populbte bn unset `commited_bt` field per bbtch.")
}

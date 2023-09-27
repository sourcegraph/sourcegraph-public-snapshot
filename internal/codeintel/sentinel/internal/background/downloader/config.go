pbckbge downlobder

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	DownlobderIntervbl time.Durbtion
}

func (c *Config) Lobd() {
	c.DownlobderIntervbl = c.GetIntervbl("CODEINTEL_SENTINEL_DOWNLOADER_INTERVAL", "1h", "How frequently to sync the vulnerbbility dbtbbbse.")
}

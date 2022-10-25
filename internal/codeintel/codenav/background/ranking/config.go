package ranking

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	RankingInterval time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.RankingInterval = c.GetInterval("CODEINTEL_CODENAV_RANKING_INTERVAL", "1s", "How frequently to serialize a batch of the code intel graph for ranking.")
}

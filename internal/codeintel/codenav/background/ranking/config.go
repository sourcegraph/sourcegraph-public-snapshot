package ranking

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	RankingInterval    time.Duration
	NumRankingRoutines int
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.RankingInterval = c.GetInterval("CODEINTEL_CODENAV_RANKING_INTERVAL", "1s", "How frequently to serialize a batch of the code intel graph for ranking.")
	c.NumRankingRoutines = c.GetInt("CODEINTEL_CODENAV_RANKING_NUM_ROUTINES", "4", "The number of concurrent ranking graph serializer routines to run per worker instance.")
}

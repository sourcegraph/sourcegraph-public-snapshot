package sentinel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type sentinelConfig struct {
	env.BaseConfig

	DownloaderInterval time.Duration
	MatcherInterval    time.Duration
	BatchSize          int
}

var ConfigInst = &sentinelConfig{}

func (c *sentinelConfig) Load() {
	c.DownloaderInterval = c.GetInterval("CODEINTEL_SENTINEL_DOWNLOADER_INTERVAL", "1h", "How frequently to sync the vulnerability database.")
	c.MatcherInterval = c.GetInterval("CODEINTEL_SENTINEL_MATCHER_INTERVAL", "1s", "How frequently to match existing records against known vulnerabilities.")
	c.BatchSize = c.GetInt("CODEINTEL_SENTINEL_BATCH_SIZE", "100", "How many precise indexes to scan at once for vulnerabilities.")
}

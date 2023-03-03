package sentinel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type sentinelConfig struct {
	env.BaseConfig

	DownloaderInterval time.Duration
	MatcherInterval    time.Duration
}

var ConfigInst = &sentinelConfig{}

func (c *sentinelConfig) Load() {
	c.DownloaderInterval = c.GetInterval("CODEINTEL_SENTINEL_DOWNLOADER_INTERVAL", "1h", "How frequently to sync the vulnerability database.")
	c.MatcherInterval = c.GetInterval("CODEINTEL_SENTINEL_MATCHER_INTERVAL", "1s", "How frequently to match existing records against known vulnerabilities.")
}

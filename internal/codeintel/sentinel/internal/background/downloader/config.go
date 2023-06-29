package downloader

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	DownloaderInterval time.Duration
}

func (c *Config) Load() {
	c.DownloaderInterval = c.GetInterval("CODEINTEL_SENTINEL_DOWNLOADER_INTERVAL", "1h", "How frequently to sync the vulnerability database.")
}

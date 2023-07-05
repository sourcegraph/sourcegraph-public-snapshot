package backfiller

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval  time.Duration
	BatchSize int
}

func (c *Config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_UPLOAD_BACKFILLER_INTERVAL", "10s", "The frequency with which to run periodic codeintel backfiller tasks.")
	c.BatchSize = c.GetInt("CODEINTEL_UPLOAD_BACKFILLER_BATCH_SIZE", "100", "The number of upload to populate an unset `commited_at` field per batch.")
}

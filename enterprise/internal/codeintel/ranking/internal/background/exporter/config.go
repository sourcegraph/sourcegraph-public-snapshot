package exporter

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval       time.Duration
	ReadBatchSize  int
	WriteBatchSize int
}

func (c *Config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_RANKING_SYMBOL_EXPORTER_INTERVAL", "1s", "How frequently to serialize a batch of the code intel graph for ranking.")
	c.ReadBatchSize = c.GetInt("CODEINTEL_RANKING_SYMBOL_EXPORTER_READ_BATCH_SIZE", "16", "How many uploads to process at once.")
	c.WriteBatchSize = c.GetInt("CODEINTEL_RANKING_SYMBOL_EXPORTER_WRITE_BATCH_SIZE", "10000", "The number of definitions and references to populate the ranking graph per batch.")
}

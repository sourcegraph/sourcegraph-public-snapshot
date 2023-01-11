package ranking

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type indexerConfig struct {
	env.BaseConfig

	Interval time.Duration
}

var IndexerConfigInst = &indexerConfig{}

func (c *indexerConfig) Load() {
	c.Interval = c.GetInterval("CODEINTEL_RANKING_INDEXER_INTERVAL", "10s", "The frequency with which to run periodic codeintel rank indexing tasks.")
}

type loaderConfig struct {
	env.BaseConfig

	LoadInterval  time.Duration
	MergeInterval time.Duration
}

var LoaderConfigInst = &loaderConfig{}

func (c *loaderConfig) Load() {
	c.LoadInterval = c.GetInterval("CODEINTEL_RANKING_LOADER_INTERVAL", "10s", "The frequency with which to run periodic codeintel rank loading tasks.")
	c.MergeInterval = c.GetInterval("CODEINTEL_RANKING_MERGER_INTERVAL", "1s", "The frequency with which to run periodic codeintel rank merging tasks.")
}

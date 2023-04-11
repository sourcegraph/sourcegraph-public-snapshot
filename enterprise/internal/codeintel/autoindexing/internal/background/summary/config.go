package summary

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval                   time.Duration
	NumRepositoriesToConfigure int
}

func (c *Config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_AUTOINDEXING_SUMMARY_BUILDER_INTERVAL", "30m", "How frequently to run the auto-indexing summary builder routine.")
	c.NumRepositoriesToConfigure = c.GetInt("CODEINTEL_AUTOINDEXING_DASHBOARD_NUM_REPOSITORIES", "100", "The number of repositories to use to populate the global code intelligence edashboard.")
}

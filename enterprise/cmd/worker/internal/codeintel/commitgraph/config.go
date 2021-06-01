package commitgraph

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	CommitGraphUpdateTaskInterval time.Duration
}

var config = &Config{}

func (c *Config) Load() {
	c.CommitGraphUpdateTaskInterval = c.GetInterval(
		"PRECISE_CODE_INTEL_COMMIT_GRAPH_UPDATE_TASK_INTERVAL",
		"10s",
		"The frequency with which to run periodic codeintel commit graph update tasks.",
	)
}

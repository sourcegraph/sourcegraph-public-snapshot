package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type commitGraphConfig struct {
	env.BaseConfig

	CommitGraphUpdateTaskInterval time.Duration
}

var commitGraphConfigInst = &commitGraphConfig{}

func (c *commitGraphConfig) Load() {
	c.CommitGraphUpdateTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_GRAPH_UPDATE_TASK_INTERVAL", "10s", "The frequency with which to run periodic codeintel commit graph update tasks.")
}

package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type commitGraphConfig struct {
	env.BaseConfig

	MaxAgeForNonStaleBranches     time.Duration
	MaxAgeForNonStaleTags         time.Duration
	CommitGraphUpdateTaskInterval time.Duration
}

var commitGraphConfigInst = &commitGraphConfig{}

func (c *commitGraphConfig) Load() {
	c.MaxAgeForNonStaleBranches = c.GetInterval("PRECISE_CODE_INTEL_MAX_AGE_FOR_NON_STALE_BRANCHES", "2160h", "The age after which a branch should be considered stale. Code intelligence indexes will be evicted from stale branches.")      // about 3 months
	c.MaxAgeForNonStaleTags = c.GetInterval("PRECISE_CODE_INTEL_MAX_AGE_FOR_NON_STALE_TAGS", "8760h", "The age after which a tagged commit should be considered stale. Code intelligence indexes will be evicted from stale tagged commits.") // about 1 year
	c.CommitGraphUpdateTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_COMMIT_GRAPH_UPDATE_TASK_INTERVAL", "10s", "The frequency with which to run periodic codeintel commit graph update tasks.")
}

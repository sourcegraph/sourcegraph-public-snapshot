package commitgraph

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval                      time.Duration
	MaxAgeForNonStaleBranches     time.Duration
	MaxAgeForNonStaleTags         time.Duration
	CommitGraphUpdateTaskInterval time.Duration
}

func (c *Config) Load() {
	maxAgeForNonStaleBranches := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_COMMITGRAPH_MAX_AGE_FOR_NON_STALE_BRANCHES", "PRECISE_CODE_INTEL_MAX_AGE_FOR_NON_STALE_BRANCHES")
	maxAgeForNonStaleTags := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_COMMITGRAPH_MAX_AGE_FOR_NON_STALE_TAGS", "PRECISE_CODE_INTEL_MAX_AGE_FOR_NON_STALE_TAGS")
	commitGraphUpdateTaskInterval := env.ChooseFallbackVariableName("CODEINTEL_UPLOAD_COMMITGRAPH_UPDATE_TASK_INTERVAL", "PRECISE_CODE_INTEL_COMMIT_GRAPH_UPDATE_TASK_INTERVAL")

	c.Interval = c.GetInterval("CODEINTEL_UPLOAD_COMMITGRAPH_UPDATER_INTERVAL", "1s", "How frequently to run the upload commitgraph updater routine.")
	c.MaxAgeForNonStaleBranches = c.GetInterval(maxAgeForNonStaleBranches, "2160h", "The age after which a branch should be considered stale. Code intelligence indexes will be evicted from stale branches.")      // about 3 months
	c.MaxAgeForNonStaleTags = c.GetInterval(maxAgeForNonStaleTags, "8760h", "The age after which a tagged commit should be considered stale. Code intelligence indexes will be evicted from stale tagged commits.") // about 1 year
	c.CommitGraphUpdateTaskInterval = c.GetInterval(commitGraphUpdateTaskInterval, "10s", "The frequency with which to run periodic codeintel commit graph update tasks.")
}

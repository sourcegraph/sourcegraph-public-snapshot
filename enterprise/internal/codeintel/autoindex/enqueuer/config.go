package enqueuer

import (
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	MaximumRepositoriesInspectedPerSecond    rate.Limit
	MaximumRepositoriesUpdatedPerSecond      rate.Limit
	MaximumIndexJobsPerInferredConfiguration int
}

func (c *Config) Load() {
	c.MaximumRepositoriesInspectedPerSecond = toRate(c.GetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_REPOSITORIES_INSPECTED_PER_SECOND", "0", "The maximum number of repositories inspected for auto-indexing per second. Set to zero to disable limit."))
	c.MaximumRepositoriesUpdatedPerSecond = toRate(c.GetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_REPOSITORIES_UPDATED_PER_SECOND", "0", "The maximum number of repositories cloned or fetched for auto-indexing per second. Set to zero to disable limit."))
	c.MaximumIndexJobsPerInferredConfiguration = c.GetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_INDEX_JOBS_PER_INFERRED_CONFIGURATION", "25", "Repositories with a number of inferred auto-index jobs exceeding this threshold will be auto-indexed.")
}

func toRate(value int) rate.Limit {
	if value == 0 {
		return rate.Inf
	}

	return rate.Limit(value)
}

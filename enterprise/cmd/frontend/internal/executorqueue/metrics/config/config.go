package config

import (
	"encoding/json"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Allocations      map[string]QueueAllocation
	EnvironmentLabel string
}

func (c *Config) Load() {
	c.EnvironmentLabel = c.GetOptional("EXECUTOR_METRIC_ENVIRONMENT_LABEL", "A label to pass to the custom metric to distinguish environments.")

	var err error
	if c.Allocations, err = parseAllocations(c.GetOptional("EXECUTOR_ALLOCATIONS", "Allocation map to distribute workloads across different clouds.")); err != nil {
		c.AddError(err)
	}
}

func parseAllocations(allocations string) (map[string]QueueAllocation, error) {
	m := map[string]map[string]float64{}
	if allocations != "" {
		if err := json.Unmarshal([]byte(allocations), &m); err != nil {
			return nil, errors.Wrap(err, "parsing EXECUTOR_ALLOCATIONS")
		}
	}

	return normalizeAllocations(
		m,
		os.Getenv("EXECUTOR_METRIC_AWS_NAMESPACE") != "",
		os.Getenv("EXECUTOR_METRIC_GCP_PROJECT_ID") != "",
	)
}

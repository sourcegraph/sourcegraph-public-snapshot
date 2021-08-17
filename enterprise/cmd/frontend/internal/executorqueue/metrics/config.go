package metrics

import (
	"encoding/json"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	EnvironmentLabel string
	Allocations      allocationConfigs
}

type allocationConfig map[string]float64
type allocationConfigs map[string]allocationConfig

func (ac allocationConfig) IsConfigured() bool {
	if len(ac) == 0 {
		return false
	}

	for _, val := range ac {
		if val != 0.0 {
			return true
		}
	}
	return false
}

func (c *Config) Load() {
	c.EnvironmentLabel = c.GetOptional("EXECUTOR_METRIC_ENVIRONMENT_LABEL", "A label to pass to the custom metric to distinguish environments.")
	allocations := c.GetOptional("EXECUTOR_ALLOCATIONS", "Allocation map to distribute workloads across different clouds.")
	if allocations != "" {
		if err := json.Unmarshal([]byte(allocations), &c.Allocations); err != nil {
			panic(errors.Wrap(err, "parsing EXECUTOR_ALLOCATIONS"))
		}
	}
}

var metricsConfig = &Config{}

func init() {
	metricsConfig.Load()
}

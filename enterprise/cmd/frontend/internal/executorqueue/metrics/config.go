package metrics

import (
	"encoding/json"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	EnvironmentLabel string
	Allocations      map[string]QueueAllocation
	AWSConfig        awsConfig
	GCPConfig        gcpConfig
}

var (
	awsConfigured = os.Getenv("EXECUTOR_METRIC_AWS_NAMESPACE") != ""
	gcpConfigured = os.Getenv("EXECUTOR_METRIC_GCP_PROJECT_ID") != ""
)

func (c *Config) Load() {
	c.EnvironmentLabel = c.Get("EXECUTOR_METRIC_ENVIRONMENT_LABEL", "dev", "A label to pass to the custom metric to distinguish environments.")

	var err error
	if c.Allocations, err = parseAllocations(c.GetOptional("EXECUTOR_ALLOCATIONS", "Allocation map to distribute workloads across different clouds.")); err != nil {
		c.AddError(err)
	}

	if awsConfigured {
		c.AWSConfig.load(&c.BaseConfig)
	}
	if gcpConfigured {
		c.GCPConfig.load(&c.BaseConfig)
	}
}

func parseAllocations(allocations string) (map[string]QueueAllocation, error) {
	m := map[string]map[string]float64{}
	if allocations != "" {
		if err := json.Unmarshal([]byte(allocations), &m); err != nil {
			return nil, errors.Wrap(err, "parsing EXECUTOR_ALLOCATIONS")
		}
	}

	return normalizeAllocations(m, awsConfigured, gcpConfigured)
}

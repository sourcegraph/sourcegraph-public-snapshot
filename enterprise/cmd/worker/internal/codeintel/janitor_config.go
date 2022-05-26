package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type janitorConfig struct {
	env.BaseConfig

	CleanupTaskInterval                    time.Duration
	ConfigurationPolicyMembershipBatchSize int

	MetricsConfig *executorqueue.Config
}

var janitorConfigInst = &janitorConfig{}

func (c *janitorConfig) Load() {
	metricsConfig := executorqueue.InitMetricsConfig()
	metricsConfig.Load()
	c.MetricsConfig = metricsConfig

	c.CleanupTaskInterval = c.GetInterval("PRECISE_CODE_INTEL_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic codeintel cleanup tasks.")
	c.ConfigurationPolicyMembershipBatchSize = c.GetInt("PRECISE_CODE_INTEL_CONFIGURATION_POLICY_MEMBERSHIP_BATCH_SIZE", "100", "The maximum number of policy configurations to update repository membership for at a time.")
}

func (c *janitorConfig) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.MetricsConfig.Validate())
	return errs
}

package repository_matcher

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Interval                               time.Duration
	ConfigurationPolicyMembershipBatchSize int
}

func (c *Config) Load() {
	configurationPolicyMembershipBatchSize := env.ChooseFallbackVariableName("CODEINTEL_POLICIES_REPO_MATCHER_CONFIGURATION_POLICY_MEMBERSHIP_BATCH_SIZE", "PRECISE_CODE_INTEL_CONFIGURATION_POLICY_MEMBERSHIP_BATCH_SIZE")

	c.Interval = c.GetInterval("CODEINTEL_POLICIES_REPO_MATCHER_INTERVAL", "1m", "How frequently to run the policies repository matcher routine.")
	c.ConfigurationPolicyMembershipBatchSize = c.GetInt(configurationPolicyMembershipBatchSize, "100", "The maximum number of policy configurations to update repository membership for at a time.")
}

package aws

import "github.com/sourcegraph/sourcegraph/internal/env"

type AWSConfig struct {
	env.BaseConfig

	MetricNamespace string
}

func (c *AWSConfig) Load() {
	c.MetricNamespace = c.GetOptional("EXECUTOR_METRIC_AWS_NAMESPACE", "The namespace to which to export the custom metric for scaling executors.")
}

var awsConfig = &AWSConfig{}

func init() {
	awsConfig.Load()
}

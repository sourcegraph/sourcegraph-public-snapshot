package loggingmetric


type LoggingMetricBucketOptions struct {
	// explicit_buckets block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#explicit_buckets LoggingMetric#explicit_buckets}
	ExplicitBuckets *LoggingMetricBucketOptionsExplicitBuckets `field:"optional" json:"explicitBuckets" yaml:"explicitBuckets"`
	// exponential_buckets block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#exponential_buckets LoggingMetric#exponential_buckets}
	ExponentialBuckets *LoggingMetricBucketOptionsExponentialBuckets `field:"optional" json:"exponentialBuckets" yaml:"exponentialBuckets"`
	// linear_buckets block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#linear_buckets LoggingMetric#linear_buckets}
	LinearBuckets *LoggingMetricBucketOptionsLinearBuckets `field:"optional" json:"linearBuckets" yaml:"linearBuckets"`
}


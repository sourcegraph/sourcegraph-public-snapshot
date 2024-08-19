package loggingmetric


type LoggingMetricBucketOptionsLinearBuckets struct {
	// Must be greater than 0.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#num_finite_buckets LoggingMetric#num_finite_buckets}
	NumFiniteBuckets *float64 `field:"required" json:"numFiniteBuckets" yaml:"numFiniteBuckets"`
	// Lower bound of the first bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#offset LoggingMetric#offset}
	Offset *float64 `field:"required" json:"offset" yaml:"offset"`
	// Must be greater than 0.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#width LoggingMetric#width}
	Width *float64 `field:"required" json:"width" yaml:"width"`
}


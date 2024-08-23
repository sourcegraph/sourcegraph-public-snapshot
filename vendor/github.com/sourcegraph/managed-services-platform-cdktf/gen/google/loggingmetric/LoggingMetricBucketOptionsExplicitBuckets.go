package loggingmetric


type LoggingMetricBucketOptionsExplicitBuckets struct {
	// The values must be monotonically increasing.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#bounds LoggingMetric#bounds}
	Bounds *[]*float64 `field:"required" json:"bounds" yaml:"bounds"`
}


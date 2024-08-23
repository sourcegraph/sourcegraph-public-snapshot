package loggingmetric


type LoggingMetricMetricDescriptor struct {
	// Whether the metric records instantaneous values, changes to a value, etc.
	//
	// Some combinations of metricKind and valueType might not be supported.
	// For counter metrics, set this to DELTA. Possible values: ["DELTA", "GAUGE", "CUMULATIVE"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#metric_kind LoggingMetric#metric_kind}
	MetricKind *string `field:"required" json:"metricKind" yaml:"metricKind"`
	// Whether the measurement is an integer, a floating-point number, etc.
	//
	// Some combinations of metricKind and valueType might not be supported.
	// For counter metrics, set this to INT64. Possible values: ["BOOL", "INT64", "DOUBLE", "STRING", "DISTRIBUTION", "MONEY"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#value_type LoggingMetric#value_type}
	ValueType *string `field:"required" json:"valueType" yaml:"valueType"`
	// A concise name for the metric, which can be displayed in user interfaces.
	//
	// Use sentence case
	// without an ending period, for example "Request count". This field is optional but it is
	// recommended to be set for any metrics associated with user-visible concepts, such as Quota.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#display_name LoggingMetric#display_name}
	DisplayName *string `field:"optional" json:"displayName" yaml:"displayName"`
	// labels block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#labels LoggingMetric#labels}
	Labels interface{} `field:"optional" json:"labels" yaml:"labels"`
	// The unit in which the metric value is reported.
	//
	// It is only applicable if the valueType is
	// 'INT64', 'DOUBLE', or 'DISTRIBUTION'. The supported units are a subset of
	// [The Unified Code for Units of Measure](http://unitsofmeasure.org/ucum.html) standard
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/logging_metric#unit LoggingMetric#unit}
	Unit *string `field:"optional" json:"unit" yaml:"unit"`
}


package computesubnetwork


type ComputeSubnetworkLogConfig struct {
	// Can only be specified if VPC flow logging for this subnetwork is enabled.
	//
	// Toggles the aggregation interval for collecting flow logs. Increasing the
	// interval time will reduce the amount of generated flow logs for long
	// lasting connections. Default is an interval of 5 seconds per connection. Default value: "INTERVAL_5_SEC" Possible values: ["INTERVAL_5_SEC", "INTERVAL_30_SEC", "INTERVAL_1_MIN", "INTERVAL_5_MIN", "INTERVAL_10_MIN", "INTERVAL_15_MIN"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_subnetwork#aggregation_interval ComputeSubnetwork#aggregation_interval}
	AggregationInterval *string `field:"optional" json:"aggregationInterval" yaml:"aggregationInterval"`
	// Export filter used to define which VPC flow logs should be logged, as as CEL expression.
	//
	// See
	// https://cloud.google.com/vpc/docs/flow-logs#filtering for details on how to format this field.
	// The default value is 'true', which evaluates to include everything.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_subnetwork#filter_expr ComputeSubnetwork#filter_expr}
	FilterExpr *string `field:"optional" json:"filterExpr" yaml:"filterExpr"`
	// Can only be specified if VPC flow logging for this subnetwork is enabled.
	//
	// The value of the field must be in [0, 1]. Set the sampling rate of VPC
	// flow logs within the subnetwork where 1.0 means all collected logs are
	// reported and 0.0 means no logs are reported. Default is 0.5 which means
	// half of all collected logs are reported.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_subnetwork#flow_sampling ComputeSubnetwork#flow_sampling}
	FlowSampling *float64 `field:"optional" json:"flowSampling" yaml:"flowSampling"`
	// Can only be specified if VPC flow logging for this subnetwork is enabled.
	//
	// Configures whether metadata fields should be added to the reported VPC
	// flow logs. Default value: "INCLUDE_ALL_METADATA" Possible values: ["EXCLUDE_ALL_METADATA", "INCLUDE_ALL_METADATA", "CUSTOM_METADATA"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_subnetwork#metadata ComputeSubnetwork#metadata}
	Metadata *string `field:"optional" json:"metadata" yaml:"metadata"`
	// List of metadata fields that should be added to reported logs.
	//
	// Can only be specified if VPC flow logs for this subnetwork is enabled and "metadata" is set to CUSTOM_METADATA.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_subnetwork#metadata_fields ComputeSubnetwork#metadata_fields}
	MetadataFields *[]*string `field:"optional" json:"metadataFields" yaml:"metadataFields"`
}


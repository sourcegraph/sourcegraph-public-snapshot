package bigquerytable


type BigqueryTableRangePartitioningRange struct {
	// End of the range partitioning, exclusive.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#end BigqueryTable#end}
	End *float64 `field:"required" json:"end" yaml:"end"`
	// The width of each range within the partition.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#interval BigqueryTable#interval}
	Interval *float64 `field:"required" json:"interval" yaml:"interval"`
	// Start of the range partitioning, inclusive.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#start BigqueryTable#start}
	Start *float64 `field:"required" json:"start" yaml:"start"`
}


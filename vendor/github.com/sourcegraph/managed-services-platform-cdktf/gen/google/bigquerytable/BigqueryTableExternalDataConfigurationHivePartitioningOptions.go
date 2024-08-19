package bigquerytable


type BigqueryTableExternalDataConfigurationHivePartitioningOptions struct {
	// When set, what mode of hive partitioning to use when reading data.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#mode BigqueryTable#mode}
	Mode *string `field:"optional" json:"mode" yaml:"mode"`
	// If set to true, queries over this table require a partition filter that can be used for partition elimination to be specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#require_partition_filter BigqueryTable#require_partition_filter}
	RequirePartitionFilter interface{} `field:"optional" json:"requirePartitionFilter" yaml:"requirePartitionFilter"`
	// When hive partition detection is requested, a common for all source uris must be required.
	//
	// The prefix must end immediately before the partition key encoding begins.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#source_uri_prefix BigqueryTable#source_uri_prefix}
	SourceUriPrefix *string `field:"optional" json:"sourceUriPrefix" yaml:"sourceUriPrefix"`
}


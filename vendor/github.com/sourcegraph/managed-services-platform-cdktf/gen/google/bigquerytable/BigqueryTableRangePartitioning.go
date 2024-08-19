package bigquerytable


type BigqueryTableRangePartitioning struct {
	// The field used to determine how to create a range-based partition.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#field BigqueryTable#field}
	Field *string `field:"required" json:"field" yaml:"field"`
	// range block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#range BigqueryTable#range}
	Range *BigqueryTableRangePartitioningRange `field:"required" json:"range" yaml:"range"`
}


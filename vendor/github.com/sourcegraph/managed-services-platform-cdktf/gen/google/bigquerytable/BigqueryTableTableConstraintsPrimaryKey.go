package bigquerytable


type BigqueryTableTableConstraintsPrimaryKey struct {
	// The columns that are composed of the primary key constraint.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#columns BigqueryTable#columns}
	Columns *[]*string `field:"required" json:"columns" yaml:"columns"`
}


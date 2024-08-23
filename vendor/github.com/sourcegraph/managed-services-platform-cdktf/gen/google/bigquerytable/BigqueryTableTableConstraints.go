package bigquerytable


type BigqueryTableTableConstraints struct {
	// foreign_keys block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#foreign_keys BigqueryTable#foreign_keys}
	ForeignKeys interface{} `field:"optional" json:"foreignKeys" yaml:"foreignKeys"`
	// primary_key block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#primary_key BigqueryTable#primary_key}
	PrimaryKey *BigqueryTableTableConstraintsPrimaryKey `field:"optional" json:"primaryKey" yaml:"primaryKey"`
}


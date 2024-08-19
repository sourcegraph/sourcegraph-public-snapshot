package bigquerytable


type BigqueryTableTableConstraintsForeignKeysColumnReferences struct {
	// The column in the primary key that are referenced by the referencingColumn.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#referenced_column BigqueryTable#referenced_column}
	ReferencedColumn *string `field:"required" json:"referencedColumn" yaml:"referencedColumn"`
	// The column that composes the foreign key.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#referencing_column BigqueryTable#referencing_column}
	ReferencingColumn *string `field:"required" json:"referencingColumn" yaml:"referencingColumn"`
}


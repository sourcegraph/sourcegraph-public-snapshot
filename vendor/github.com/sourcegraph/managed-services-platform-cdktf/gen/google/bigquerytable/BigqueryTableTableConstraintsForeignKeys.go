package bigquerytable


type BigqueryTableTableConstraintsForeignKeys struct {
	// column_references block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#column_references BigqueryTable#column_references}
	ColumnReferences *BigqueryTableTableConstraintsForeignKeysColumnReferences `field:"required" json:"columnReferences" yaml:"columnReferences"`
	// referenced_table block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#referenced_table BigqueryTable#referenced_table}
	ReferencedTable *BigqueryTableTableConstraintsForeignKeysReferencedTable `field:"required" json:"referencedTable" yaml:"referencedTable"`
	// Set only if the foreign key constraint is named.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#name BigqueryTable#name}
	Name *string `field:"optional" json:"name" yaml:"name"`
}


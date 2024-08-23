package bigquerytable


type BigqueryTableExternalDataConfigurationGoogleSheetsOptions struct {
	// Range of a sheet to query from.
	//
	// Only used when non-empty. At least one of range or skip_leading_rows must be set. Typical format: "sheet_name!top_left_cell_id:bottom_right_cell_id" For example: "sheet1!A1:B20"
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#range BigqueryTable#range}
	Range *string `field:"optional" json:"range" yaml:"range"`
	// The number of rows at the top of the sheet that BigQuery will skip when reading the data.
	//
	// At least one of range or skip_leading_rows must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#skip_leading_rows BigqueryTable#skip_leading_rows}
	SkipLeadingRows *float64 `field:"optional" json:"skipLeadingRows" yaml:"skipLeadingRows"`
}


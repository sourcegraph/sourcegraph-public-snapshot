package bigquerytable


type BigqueryTableMaterializedView struct {
	// A query whose result is persisted.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#query BigqueryTable#query}
	Query *string `field:"required" json:"query" yaml:"query"`
	// Allow non incremental materialized view definition. The default value is false.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#allow_non_incremental_definition BigqueryTable#allow_non_incremental_definition}
	AllowNonIncrementalDefinition interface{} `field:"optional" json:"allowNonIncrementalDefinition" yaml:"allowNonIncrementalDefinition"`
	// Specifies if BigQuery should automatically refresh materialized view when the base table is updated. The default is true.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#enable_refresh BigqueryTable#enable_refresh}
	EnableRefresh interface{} `field:"optional" json:"enableRefresh" yaml:"enableRefresh"`
	// Specifies maximum frequency at which this materialized view will be refreshed. The default is 1800000.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#refresh_interval_ms BigqueryTable#refresh_interval_ms}
	RefreshIntervalMs *float64 `field:"optional" json:"refreshIntervalMs" yaml:"refreshIntervalMs"`
}


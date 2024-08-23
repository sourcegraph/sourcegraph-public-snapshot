package bigquerytable


type BigqueryTableTableReplicationInfo struct {
	// The ID of the source dataset.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#source_dataset_id BigqueryTable#source_dataset_id}
	SourceDatasetId *string `field:"required" json:"sourceDatasetId" yaml:"sourceDatasetId"`
	// The ID of the source project.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#source_project_id BigqueryTable#source_project_id}
	SourceProjectId *string `field:"required" json:"sourceProjectId" yaml:"sourceProjectId"`
	// The ID of the source materialized view.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#source_table_id BigqueryTable#source_table_id}
	SourceTableId *string `field:"required" json:"sourceTableId" yaml:"sourceTableId"`
	// The interval at which the source materialized view is polled for updates. The default is 300000.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#replication_interval_ms BigqueryTable#replication_interval_ms}
	ReplicationIntervalMs *float64 `field:"optional" json:"replicationIntervalMs" yaml:"replicationIntervalMs"`
}


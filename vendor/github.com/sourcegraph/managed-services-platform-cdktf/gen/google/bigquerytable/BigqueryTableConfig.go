package bigquerytable

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type BigqueryTableConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// The dataset ID to create the table in. Changing this forces a new resource to be created.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#dataset_id BigqueryTable#dataset_id}
	DatasetId *string `field:"required" json:"datasetId" yaml:"datasetId"`
	// A unique ID for the resource. Changing this forces a new resource to be created.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#table_id BigqueryTable#table_id}
	TableId *string `field:"required" json:"tableId" yaml:"tableId"`
	// Specifies column names to use for data clustering.
	//
	// Up to four top-level columns are allowed, and should be specified in descending priority order.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#clustering BigqueryTable#clustering}
	Clustering *[]*string `field:"optional" json:"clustering" yaml:"clustering"`
	// Whether or not to allow Terraform to destroy the instance.
	//
	// Unless this field is set to false in Terraform state, a terraform destroy or terraform apply that would delete the instance will fail.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#deletion_protection BigqueryTable#deletion_protection}
	DeletionProtection interface{} `field:"optional" json:"deletionProtection" yaml:"deletionProtection"`
	// The field description.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#description BigqueryTable#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// encryption_configuration block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#encryption_configuration BigqueryTable#encryption_configuration}
	EncryptionConfiguration *BigqueryTableEncryptionConfiguration `field:"optional" json:"encryptionConfiguration" yaml:"encryptionConfiguration"`
	// The time when this table expires, in milliseconds since the epoch.
	//
	// If not present, the table will persist indefinitely. Expired tables will be deleted and their storage reclaimed.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#expiration_time BigqueryTable#expiration_time}
	ExpirationTime *float64 `field:"optional" json:"expirationTime" yaml:"expirationTime"`
	// external_data_configuration block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#external_data_configuration BigqueryTable#external_data_configuration}
	ExternalDataConfiguration *BigqueryTableExternalDataConfiguration `field:"optional" json:"externalDataConfiguration" yaml:"externalDataConfiguration"`
	// A descriptive name for the table.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#friendly_name BigqueryTable#friendly_name}
	FriendlyName *string `field:"optional" json:"friendlyName" yaml:"friendlyName"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#id BigqueryTable#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// A mapping of labels to assign to the resource.
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#labels BigqueryTable#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// materialized_view block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#materialized_view BigqueryTable#materialized_view}
	MaterializedView *BigqueryTableMaterializedView `field:"optional" json:"materializedView" yaml:"materializedView"`
	// The maximum staleness of data that could be returned when the table (or stale MV) is queried.
	//
	// Staleness encoded as a string encoding of [SQL IntervalValue type](https://cloud.google.com/bigquery/docs/reference/standard-sql/data-types#interval_type).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#max_staleness BigqueryTable#max_staleness}
	MaxStaleness *string `field:"optional" json:"maxStaleness" yaml:"maxStaleness"`
	// The ID of the project in which the resource belongs.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#project BigqueryTable#project}
	Project *string `field:"optional" json:"project" yaml:"project"`
	// range_partitioning block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#range_partitioning BigqueryTable#range_partitioning}
	RangePartitioning *BigqueryTableRangePartitioning `field:"optional" json:"rangePartitioning" yaml:"rangePartitioning"`
	// If set to true, queries over this table require a partition filter that can be used for partition elimination to be specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#require_partition_filter BigqueryTable#require_partition_filter}
	RequirePartitionFilter interface{} `field:"optional" json:"requirePartitionFilter" yaml:"requirePartitionFilter"`
	// A JSON schema for the table.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#schema BigqueryTable#schema}
	Schema *string `field:"optional" json:"schema" yaml:"schema"`
	// table_constraints block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#table_constraints BigqueryTable#table_constraints}
	TableConstraints *BigqueryTableTableConstraints `field:"optional" json:"tableConstraints" yaml:"tableConstraints"`
	// table_replication_info block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#table_replication_info BigqueryTable#table_replication_info}
	TableReplicationInfo *BigqueryTableTableReplicationInfo `field:"optional" json:"tableReplicationInfo" yaml:"tableReplicationInfo"`
	// time_partitioning block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#time_partitioning BigqueryTable#time_partitioning}
	TimePartitioning *BigqueryTableTimePartitioning `field:"optional" json:"timePartitioning" yaml:"timePartitioning"`
	// view block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table#view BigqueryTable#view}
	View *BigqueryTableView `field:"optional" json:"view" yaml:"view"`
}


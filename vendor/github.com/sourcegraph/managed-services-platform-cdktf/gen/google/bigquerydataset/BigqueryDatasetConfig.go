package bigquerydataset

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type BigqueryDatasetConfig struct {
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
	// A unique ID for this dataset, without the project name.
	//
	// The ID
	// must contain only letters (a-z, A-Z), numbers (0-9), or
	// underscores (_). The maximum length is 1,024 characters.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#dataset_id BigqueryDataset#dataset_id}
	DatasetId *string `field:"required" json:"datasetId" yaml:"datasetId"`
	// access block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#access BigqueryDataset#access}
	Access interface{} `field:"optional" json:"access" yaml:"access"`
	// Defines the default collation specification of future tables created in the dataset.
	//
	// If a table is created in this dataset without table-level
	// default collation, then the table inherits the dataset default collation,
	// which is applied to the string fields that do not have explicit collation
	// specified. A change to this field affects only tables created afterwards,
	// and does not alter the existing tables.
	//
	// The following values are supported:
	// - 'und:ci': undetermined locale, case insensitive.
	// - '': empty string. Default to case-sensitive behavior.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#default_collation BigqueryDataset#default_collation}
	DefaultCollation *string `field:"optional" json:"defaultCollation" yaml:"defaultCollation"`
	// default_encryption_configuration block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#default_encryption_configuration BigqueryDataset#default_encryption_configuration}
	DefaultEncryptionConfiguration *BigqueryDatasetDefaultEncryptionConfiguration `field:"optional" json:"defaultEncryptionConfiguration" yaml:"defaultEncryptionConfiguration"`
	// The default partition expiration for all partitioned tables in the dataset, in milliseconds.
	//
	// Once this property is set, all newly-created partitioned tables in
	// the dataset will have an 'expirationMs' property in the 'timePartitioning'
	// settings set to this value, and changing the value will only
	// affect new tables, not existing ones. The storage in a partition will
	// have an expiration time of its partition time plus this value.
	// Setting this property overrides the use of 'defaultTableExpirationMs'
	// for partitioned tables: only one of 'defaultTableExpirationMs' and
	// 'defaultPartitionExpirationMs' will be used for any new partitioned
	// table. If you provide an explicit 'timePartitioning.expirationMs' when
	// creating or updating a partitioned table, that value takes precedence
	// over the default partition expiration time indicated by this property.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#default_partition_expiration_ms BigqueryDataset#default_partition_expiration_ms}
	DefaultPartitionExpirationMs *float64 `field:"optional" json:"defaultPartitionExpirationMs" yaml:"defaultPartitionExpirationMs"`
	// The default lifetime of all tables in the dataset, in milliseconds. The minimum value is 3600000 milliseconds (one hour).
	//
	// Once this property is set, all newly-created tables in the dataset
	// will have an 'expirationTime' property set to the creation time plus
	// the value in this property, and changing the value will only affect
	// new tables, not existing ones. When the 'expirationTime' for a given
	// table is reached, that table will be deleted automatically.
	// If a table's 'expirationTime' is modified or removed before the
	// table expires, or if you provide an explicit 'expirationTime' when
	// creating a table, that value takes precedence over the default
	// expiration time indicated by this property.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#default_table_expiration_ms BigqueryDataset#default_table_expiration_ms}
	DefaultTableExpirationMs *float64 `field:"optional" json:"defaultTableExpirationMs" yaml:"defaultTableExpirationMs"`
	// If set to 'true', delete all the tables in the dataset when destroying the resource;
	//
	// otherwise,
	// destroying the resource will fail if tables are present.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#delete_contents_on_destroy BigqueryDataset#delete_contents_on_destroy}
	DeleteContentsOnDestroy interface{} `field:"optional" json:"deleteContentsOnDestroy" yaml:"deleteContentsOnDestroy"`
	// A user-friendly description of the dataset.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#description BigqueryDataset#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// external_dataset_reference block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#external_dataset_reference BigqueryDataset#external_dataset_reference}
	ExternalDatasetReference *BigqueryDatasetExternalDatasetReference `field:"optional" json:"externalDatasetReference" yaml:"externalDatasetReference"`
	// A descriptive name for the dataset.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#friendly_name BigqueryDataset#friendly_name}
	FriendlyName *string `field:"optional" json:"friendlyName" yaml:"friendlyName"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#id BigqueryDataset#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// TRUE if the dataset and its table names are case-insensitive, otherwise FALSE.
	//
	// By default, this is FALSE, which means the dataset and its table names are
	// case-sensitive. This field does not affect routine references.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#is_case_insensitive BigqueryDataset#is_case_insensitive}
	IsCaseInsensitive interface{} `field:"optional" json:"isCaseInsensitive" yaml:"isCaseInsensitive"`
	// The labels associated with this dataset. You can use these to organize and group your datasets.
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#labels BigqueryDataset#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// The geographic location where the dataset should reside. See [official docs](https://cloud.google.com/bigquery/docs/dataset-locations).
	//
	// There are two types of locations, regional or multi-regional. A regional
	// location is a specific geographic place, such as Tokyo, and a multi-regional
	// location is a large geographic area, such as the United States, that
	// contains at least two geographic places.
	//
	//
	// The default value is multi-regional location 'US'.
	// Changing this forces a new resource to be created.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#location BigqueryDataset#location}
	Location *string `field:"optional" json:"location" yaml:"location"`
	// Defines the time travel window in hours.
	//
	// The value can be from 48 to 168 hours (2 to 7 days).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#max_time_travel_hours BigqueryDataset#max_time_travel_hours}
	MaxTimeTravelHours *string `field:"optional" json:"maxTimeTravelHours" yaml:"maxTimeTravelHours"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#project BigqueryDataset#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// Specifies the storage billing model for the dataset.
	//
	// Set this flag value to LOGICAL to use logical bytes for storage billing,
	// or to PHYSICAL to use physical bytes instead.
	//
	// LOGICAL is the default if this flag isn't specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#storage_billing_model BigqueryDataset#storage_billing_model}
	StorageBillingModel *string `field:"optional" json:"storageBillingModel" yaml:"storageBillingModel"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_dataset#timeouts BigqueryDataset#timeouts}
	Timeouts *BigqueryDatasetTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}


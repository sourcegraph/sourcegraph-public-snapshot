package sqldatabaseinstance

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type SqlDatabaseInstanceConfig struct {
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
	// The MySQL, PostgreSQL or SQL Server (beta) version to use.
	//
	// Supported values include MYSQL_5_6, MYSQL_5_7, MYSQL_8_0, POSTGRES_9_6, POSTGRES_10, POSTGRES_11, POSTGRES_12, POSTGRES_13, POSTGRES_14, POSTGRES_15, SQLSERVER_2017_STANDARD, SQLSERVER_2017_ENTERPRISE, SQLSERVER_2017_EXPRESS, SQLSERVER_2017_WEB. Database Version Policies includes an up-to-date reference of supported versions.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#database_version SqlDatabaseInstance#database_version}
	DatabaseVersion *string `field:"required" json:"databaseVersion" yaml:"databaseVersion"`
	// clone block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#clone SqlDatabaseInstance#clone}
	Clone *SqlDatabaseInstanceClone `field:"optional" json:"clone" yaml:"clone"`
	// Used to block Terraform from deleting a SQL Instance. Defaults to true.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#deletion_protection SqlDatabaseInstance#deletion_protection}
	DeletionProtection interface{} `field:"optional" json:"deletionProtection" yaml:"deletionProtection"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#encryption_key_name SqlDatabaseInstance#encryption_key_name}.
	EncryptionKeyName *string `field:"optional" json:"encryptionKeyName" yaml:"encryptionKeyName"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#id SqlDatabaseInstance#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The type of the instance. The valid values are:- 'SQL_INSTANCE_TYPE_UNSPECIFIED', 'CLOUD_SQL_INSTANCE', 'ON_PREMISES_INSTANCE' and 'READ_REPLICA_INSTANCE'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#instance_type SqlDatabaseInstance#instance_type}
	InstanceType *string `field:"optional" json:"instanceType" yaml:"instanceType"`
	// Maintenance version.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#maintenance_version SqlDatabaseInstance#maintenance_version}
	MaintenanceVersion *string `field:"optional" json:"maintenanceVersion" yaml:"maintenanceVersion"`
	// The name of the instance that will act as the master in the replication setup.
	//
	// Note, this requires the master to have binary_log_enabled set, as well as existing backups.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#master_instance_name SqlDatabaseInstance#master_instance_name}
	MasterInstanceName *string `field:"optional" json:"masterInstanceName" yaml:"masterInstanceName"`
	// The name of the instance.
	//
	// If the name is left blank, Terraform will randomly generate one when the instance is first created. This is done because after a name is used, it cannot be reused for up to one week.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#name SqlDatabaseInstance#name}
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The ID of the project in which the resource belongs.
	//
	// If it is not provided, the provider project is used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#project SqlDatabaseInstance#project}
	Project *string `field:"optional" json:"project" yaml:"project"`
	// The region the instance will sit in.
	//
	// Note, Cloud SQL is not available in all regions. A valid region must be provided to use this resource. If a region is not provided in the resource definition, the provider region will be used instead, but this will be an apply-time error for instances if the provider region is not supported with Cloud SQL. If you choose not to provide the region argument for this resource, make sure you understand this.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#region SqlDatabaseInstance#region}
	Region *string `field:"optional" json:"region" yaml:"region"`
	// replica_configuration block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#replica_configuration SqlDatabaseInstance#replica_configuration}
	ReplicaConfiguration *SqlDatabaseInstanceReplicaConfiguration `field:"optional" json:"replicaConfiguration" yaml:"replicaConfiguration"`
	// restore_backup_context block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#restore_backup_context SqlDatabaseInstance#restore_backup_context}
	RestoreBackupContext *SqlDatabaseInstanceRestoreBackupContext `field:"optional" json:"restoreBackupContext" yaml:"restoreBackupContext"`
	// Initial root password. Required for MS SQL Server.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#root_password SqlDatabaseInstance#root_password}
	RootPassword *string `field:"optional" json:"rootPassword" yaml:"rootPassword"`
	// settings block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#settings SqlDatabaseInstance#settings}
	Settings *SqlDatabaseInstanceSettings `field:"optional" json:"settings" yaml:"settings"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#timeouts SqlDatabaseInstance#timeouts}
	Timeouts *SqlDatabaseInstanceTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}


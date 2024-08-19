// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type DataTerraformRemoteStatePgConfig struct {
	// Experimental.
	Defaults *map[string]interface{} `field:"optional" json:"defaults" yaml:"defaults"`
	// Experimental.
	Workspace *string `field:"optional" json:"workspace" yaml:"workspace"`
	// Postgres connection string;
	//
	// a postgres:// URL.
	// The PG_CONN_STR and standard libpq environment variables can also be used to indicate how to connect to the PostgreSQL database.
	// Experimental.
	ConnStr *string `field:"required" json:"connStr" yaml:"connStr"`
	// Name of the automatically-managed Postgres schema, default to terraform_remote_state.
	//
	// Can also be set using the PG_SCHEMA_NAME environment variable.
	// Experimental.
	SchemaName *string `field:"optional" json:"schemaName" yaml:"schemaName"`
	// If set to true, the Postgres index must already exist.
	//
	// Can also be set using the PG_SKIP_INDEX_CREATION environment variable.
	// Terraform won't try to create the index, this is useful when it has already been created by a database administrator.
	// Experimental.
	SkipIndexCreation *bool `field:"optional" json:"skipIndexCreation" yaml:"skipIndexCreation"`
	// If set to true, the Postgres schema must already exist.
	//
	// Can also be set using the PG_SKIP_SCHEMA_CREATION environment variable.
	// Terraform won't try to create the schema, this is useful when it has already been created by a database administrator.
	// Experimental.
	SkipSchemaCreation *bool `field:"optional" json:"skipSchemaCreation" yaml:"skipSchemaCreation"`
	// If set to true, the Postgres table must already exist.
	//
	// Can also be set using the PG_SKIP_TABLE_CREATION environment variable.
	// Terraform won't try to create the table, this is useful when it has already been created by a database administrator.
	// Experimental.
	SkipTableCreation *bool `field:"optional" json:"skipTableCreation" yaml:"skipTableCreation"`
}


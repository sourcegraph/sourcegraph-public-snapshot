package sqldatabaseinstance


type SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettings struct {
	// Number of backups to retain.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#retained_backups SqlDatabaseInstance#retained_backups}
	RetainedBackups *float64 `field:"required" json:"retainedBackups" yaml:"retainedBackups"`
	// The unit that 'retainedBackups' represents. Defaults to COUNT.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#retention_unit SqlDatabaseInstance#retention_unit}
	RetentionUnit *string `field:"optional" json:"retentionUnit" yaml:"retentionUnit"`
}


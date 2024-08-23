package sqldatabaseinstance


type SqlDatabaseInstanceSettingsBackupConfiguration struct {
	// backup_retention_settings block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#backup_retention_settings SqlDatabaseInstance#backup_retention_settings}
	BackupRetentionSettings *SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettings `field:"optional" json:"backupRetentionSettings" yaml:"backupRetentionSettings"`
	// True if binary logging is enabled.
	//
	// If settings.backup_configuration.enabled is false, this must be as well. Can only be used with MySQL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#binary_log_enabled SqlDatabaseInstance#binary_log_enabled}
	BinaryLogEnabled interface{} `field:"optional" json:"binaryLogEnabled" yaml:"binaryLogEnabled"`
	// True if backup configuration is enabled.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#enabled SqlDatabaseInstance#enabled}
	Enabled interface{} `field:"optional" json:"enabled" yaml:"enabled"`
	// Location of the backup configuration.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#location SqlDatabaseInstance#location}
	Location *string `field:"optional" json:"location" yaml:"location"`
	// True if Point-in-time recovery is enabled.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#point_in_time_recovery_enabled SqlDatabaseInstance#point_in_time_recovery_enabled}
	PointInTimeRecoveryEnabled interface{} `field:"optional" json:"pointInTimeRecoveryEnabled" yaml:"pointInTimeRecoveryEnabled"`
	// HH:MM format time indicating when backup configuration starts.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#start_time SqlDatabaseInstance#start_time}
	StartTime *string `field:"optional" json:"startTime" yaml:"startTime"`
	// The number of days of transaction logs we retain for point in time restore, from 1-7.
	//
	// (For PostgreSQL Enterprise Plus instances, from 1 to 35.)
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#transaction_log_retention_days SqlDatabaseInstance#transaction_log_retention_days}
	TransactionLogRetentionDays *float64 `field:"optional" json:"transactionLogRetentionDays" yaml:"transactionLogRetentionDays"`
}


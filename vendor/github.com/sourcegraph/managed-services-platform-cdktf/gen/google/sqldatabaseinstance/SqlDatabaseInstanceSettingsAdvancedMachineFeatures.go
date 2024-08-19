package sqldatabaseinstance


type SqlDatabaseInstanceSettingsAdvancedMachineFeatures struct {
	// The number of threads per physical core. Can be 1 or 2.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#threads_per_core SqlDatabaseInstance#threads_per_core}
	ThreadsPerCore *float64 `field:"optional" json:"threadsPerCore" yaml:"threadsPerCore"`
}


package sqldatabaseinstance


type SqlDatabaseInstanceSettingsDataCacheConfig struct {
	// Whether data cache is enabled for the instance.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#data_cache_enabled SqlDatabaseInstance#data_cache_enabled}
	DataCacheEnabled interface{} `field:"optional" json:"dataCacheEnabled" yaml:"dataCacheEnabled"`
}


package sqldatabaseinstance


type SqlDatabaseInstanceSettingsActiveDirectoryConfig struct {
	// Domain name of the Active Directory for SQL Server (e.g., mydomain.com).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#domain SqlDatabaseInstance#domain}
	Domain *string `field:"required" json:"domain" yaml:"domain"`
}


package sqldatabaseinstance


type SqlDatabaseInstanceSettingsIpConfigurationAuthorizedNetworks struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#value SqlDatabaseInstance#value}.
	Value *string `field:"required" json:"value" yaml:"value"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#expiration_time SqlDatabaseInstance#expiration_time}.
	ExpirationTime *string `field:"optional" json:"expirationTime" yaml:"expirationTime"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#name SqlDatabaseInstance#name}.
	Name *string `field:"optional" json:"name" yaml:"name"`
}


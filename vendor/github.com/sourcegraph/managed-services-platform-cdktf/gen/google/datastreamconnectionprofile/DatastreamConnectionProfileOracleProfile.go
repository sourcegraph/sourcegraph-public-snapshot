package datastreamconnectionprofile


type DatastreamConnectionProfileOracleProfile struct {
	// Database for the Oracle connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#database_service DatastreamConnectionProfile#database_service}
	DatabaseService *string `field:"required" json:"databaseService" yaml:"databaseService"`
	// Hostname for the Oracle connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#hostname DatastreamConnectionProfile#hostname}
	Hostname *string `field:"required" json:"hostname" yaml:"hostname"`
	// Password for the Oracle connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#password DatastreamConnectionProfile#password}
	Password *string `field:"required" json:"password" yaml:"password"`
	// Username for the Oracle connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#username DatastreamConnectionProfile#username}
	Username *string `field:"required" json:"username" yaml:"username"`
	// Connection string attributes.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#connection_attributes DatastreamConnectionProfile#connection_attributes}
	ConnectionAttributes *map[string]*string `field:"optional" json:"connectionAttributes" yaml:"connectionAttributes"`
	// Port for the Oracle connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#port DatastreamConnectionProfile#port}
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}


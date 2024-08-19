package datastreamconnectionprofile


type DatastreamConnectionProfilePostgresqlProfile struct {
	// Database for the PostgreSQL connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#database DatastreamConnectionProfile#database}
	Database *string `field:"required" json:"database" yaml:"database"`
	// Hostname for the PostgreSQL connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#hostname DatastreamConnectionProfile#hostname}
	Hostname *string `field:"required" json:"hostname" yaml:"hostname"`
	// Password for the PostgreSQL connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#password DatastreamConnectionProfile#password}
	Password *string `field:"required" json:"password" yaml:"password"`
	// Username for the PostgreSQL connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#username DatastreamConnectionProfile#username}
	Username *string `field:"required" json:"username" yaml:"username"`
	// Port for the PostgreSQL connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#port DatastreamConnectionProfile#port}
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}


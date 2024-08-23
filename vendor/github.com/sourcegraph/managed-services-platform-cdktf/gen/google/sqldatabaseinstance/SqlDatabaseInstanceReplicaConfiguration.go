package sqldatabaseinstance


type SqlDatabaseInstanceReplicaConfiguration struct {
	// PEM representation of the trusted CA's x509 certificate.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#ca_certificate SqlDatabaseInstance#ca_certificate}
	CaCertificate *string `field:"optional" json:"caCertificate" yaml:"caCertificate"`
	// PEM representation of the replica's x509 certificate.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#client_certificate SqlDatabaseInstance#client_certificate}
	ClientCertificate *string `field:"optional" json:"clientCertificate" yaml:"clientCertificate"`
	// PEM representation of the replica's private key. The corresponding public key in encoded in the client_certificate.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#client_key SqlDatabaseInstance#client_key}
	ClientKey *string `field:"optional" json:"clientKey" yaml:"clientKey"`
	// The number of seconds between connect retries. MySQL's default is 60 seconds.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#connect_retry_interval SqlDatabaseInstance#connect_retry_interval}
	ConnectRetryInterval *float64 `field:"optional" json:"connectRetryInterval" yaml:"connectRetryInterval"`
	// Path to a SQL file in Google Cloud Storage from which replica instances are created. Format is gs://bucket/filename.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#dump_file_path SqlDatabaseInstance#dump_file_path}
	DumpFilePath *string `field:"optional" json:"dumpFilePath" yaml:"dumpFilePath"`
	// Specifies if the replica is the failover target.
	//
	// If the field is set to true the replica will be designated as a failover replica. If the master instance fails, the replica instance will be promoted as the new master instance. Not supported for Postgres
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#failover_target SqlDatabaseInstance#failover_target}
	FailoverTarget interface{} `field:"optional" json:"failoverTarget" yaml:"failoverTarget"`
	// Time in ms between replication heartbeats.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#master_heartbeat_period SqlDatabaseInstance#master_heartbeat_period}
	MasterHeartbeatPeriod *float64 `field:"optional" json:"masterHeartbeatPeriod" yaml:"masterHeartbeatPeriod"`
	// Password for the replication connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#password SqlDatabaseInstance#password}
	Password *string `field:"optional" json:"password" yaml:"password"`
	// Permissible ciphers for use in SSL encryption.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#ssl_cipher SqlDatabaseInstance#ssl_cipher}
	SslCipher *string `field:"optional" json:"sslCipher" yaml:"sslCipher"`
	// Username for replication connection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#username SqlDatabaseInstance#username}
	Username *string `field:"optional" json:"username" yaml:"username"`
	// True if the master's common name value is checked during the SSL handshake.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance#verify_server_certificate SqlDatabaseInstance#verify_server_certificate}
	VerifyServerCertificate interface{} `field:"optional" json:"verifyServerCertificate" yaml:"verifyServerCertificate"`
}


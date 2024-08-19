package datastreamconnectionprofile


type DatastreamConnectionProfileForwardSshConnectivity struct {
	// Hostname for the SSH tunnel.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#hostname DatastreamConnectionProfile#hostname}
	Hostname *string `field:"required" json:"hostname" yaml:"hostname"`
	// Username for the SSH tunnel.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#username DatastreamConnectionProfile#username}
	Username *string `field:"required" json:"username" yaml:"username"`
	// SSH password.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#password DatastreamConnectionProfile#password}
	Password *string `field:"optional" json:"password" yaml:"password"`
	// Port for the SSH tunnel.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#port DatastreamConnectionProfile#port}
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// SSH private key.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#private_key DatastreamConnectionProfile#private_key}
	PrivateKey *string `field:"optional" json:"privateKey" yaml:"privateKey"`
}


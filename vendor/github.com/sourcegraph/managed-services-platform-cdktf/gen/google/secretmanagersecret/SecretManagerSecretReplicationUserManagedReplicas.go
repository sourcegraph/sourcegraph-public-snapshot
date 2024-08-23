package secretmanagersecret


type SecretManagerSecretReplicationUserManagedReplicas struct {
	// The canonical IDs of the location to replicate data. For example: "us-east1".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#location SecretManagerSecret#location}
	Location *string `field:"required" json:"location" yaml:"location"`
	// customer_managed_encryption block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#customer_managed_encryption SecretManagerSecret#customer_managed_encryption}
	CustomerManagedEncryption *SecretManagerSecretReplicationUserManagedReplicasCustomerManagedEncryption `field:"optional" json:"customerManagedEncryption" yaml:"customerManagedEncryption"`
}


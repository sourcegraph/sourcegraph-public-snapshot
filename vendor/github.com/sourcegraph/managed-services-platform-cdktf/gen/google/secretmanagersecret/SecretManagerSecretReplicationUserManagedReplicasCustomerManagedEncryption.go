package secretmanagersecret


type SecretManagerSecretReplicationUserManagedReplicasCustomerManagedEncryption struct {
	// Describes the Cloud KMS encryption key that will be used to protect destination secret.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#kms_key_name SecretManagerSecret#kms_key_name}
	KmsKeyName *string `field:"required" json:"kmsKeyName" yaml:"kmsKeyName"`
}


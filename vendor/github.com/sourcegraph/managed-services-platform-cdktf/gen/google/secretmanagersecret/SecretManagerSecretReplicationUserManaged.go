package secretmanagersecret


type SecretManagerSecretReplicationUserManaged struct {
	// replicas block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#replicas SecretManagerSecret#replicas}
	Replicas interface{} `field:"required" json:"replicas" yaml:"replicas"`
}


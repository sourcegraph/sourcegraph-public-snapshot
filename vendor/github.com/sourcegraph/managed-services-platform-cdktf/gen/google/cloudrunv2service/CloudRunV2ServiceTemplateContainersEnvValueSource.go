package cloudrunv2service


type CloudRunV2ServiceTemplateContainersEnvValueSource struct {
	// secret_key_ref block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#secret_key_ref CloudRunV2Service#secret_key_ref}
	SecretKeyRef *CloudRunV2ServiceTemplateContainersEnvValueSourceSecretKeyRef `field:"optional" json:"secretKeyRef" yaml:"secretKeyRef"`
}


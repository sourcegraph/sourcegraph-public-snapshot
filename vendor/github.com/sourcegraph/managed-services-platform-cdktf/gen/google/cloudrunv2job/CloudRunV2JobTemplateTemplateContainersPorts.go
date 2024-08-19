package cloudrunv2job


type CloudRunV2JobTemplateTemplateContainersPorts struct {
	// Port number the container listens on. This must be a valid TCP port number, 0 < containerPort < 65536.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#container_port CloudRunV2Job#container_port}
	ContainerPort *float64 `field:"optional" json:"containerPort" yaml:"containerPort"`
	// If specified, used to specify which protocol to use. Allowed values are "http1" and "h2c".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#name CloudRunV2Job#name}
	Name *string `field:"optional" json:"name" yaml:"name"`
}


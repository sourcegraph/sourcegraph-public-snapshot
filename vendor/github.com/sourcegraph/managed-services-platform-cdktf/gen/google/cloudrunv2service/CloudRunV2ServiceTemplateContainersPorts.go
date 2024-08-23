package cloudrunv2service


type CloudRunV2ServiceTemplateContainersPorts struct {
	// Port number the container listens on. This must be a valid TCP port number, 0 < containerPort < 65536.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#container_port CloudRunV2Service#container_port}
	ContainerPort *float64 `field:"optional" json:"containerPort" yaml:"containerPort"`
	// If specified, used to specify which protocol to use. Allowed values are "http1" and "h2c".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#name CloudRunV2Service#name}
	Name *string `field:"optional" json:"name" yaml:"name"`
}


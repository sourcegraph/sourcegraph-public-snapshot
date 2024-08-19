package cloudrunv2service


type CloudRunV2ServiceTemplateContainersStartupProbeTcpSocket struct {
	// Port number to access on the container.
	//
	// Must be in the range 1 to 65535.
	// If not specified, defaults to the same value as container.ports[0].containerPort.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#port CloudRunV2Service#port}
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}


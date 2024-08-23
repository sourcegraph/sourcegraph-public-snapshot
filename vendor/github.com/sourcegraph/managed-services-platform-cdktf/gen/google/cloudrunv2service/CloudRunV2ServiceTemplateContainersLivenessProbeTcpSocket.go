package cloudrunv2service


type CloudRunV2ServiceTemplateContainersLivenessProbeTcpSocket struct {
	// Port number to access on the container.
	//
	// Must be in the range 1 to 65535.
	// If not specified, defaults to the exposed port of the container, which
	// is the value of container.ports[0].containerPort.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#port CloudRunV2Service#port}
	Port *float64 `field:"required" json:"port" yaml:"port"`
}


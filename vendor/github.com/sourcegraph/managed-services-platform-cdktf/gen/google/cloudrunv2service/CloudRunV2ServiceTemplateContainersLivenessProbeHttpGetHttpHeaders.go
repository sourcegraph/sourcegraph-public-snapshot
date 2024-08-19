package cloudrunv2service


type CloudRunV2ServiceTemplateContainersLivenessProbeHttpGetHttpHeaders struct {
	// The header field name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#name CloudRunV2Service#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The header field value.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#value CloudRunV2Service#value}
	Value *string `field:"optional" json:"value" yaml:"value"`
}


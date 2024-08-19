package clouddeploytarget


type ClouddeployTargetCustomTarget struct {
	// Required. The name of the CustomTargetType. Format must be `projects/{project}/locations/{location}/customTargetTypes/{custom_target_type}`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_target#custom_target_type ClouddeployTarget#custom_target_type}
	CustomTargetType *string `field:"required" json:"customTargetType" yaml:"customTargetType"`
}


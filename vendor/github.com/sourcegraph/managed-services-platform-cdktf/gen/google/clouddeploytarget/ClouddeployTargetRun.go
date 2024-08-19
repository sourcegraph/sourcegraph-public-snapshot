package clouddeploytarget


type ClouddeployTargetRun struct {
	// Required. The location where the Cloud Run Service should be located. Format is `projects/{project}/locations/{location}`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_target#location ClouddeployTarget#location}
	Location *string `field:"required" json:"location" yaml:"location"`
}


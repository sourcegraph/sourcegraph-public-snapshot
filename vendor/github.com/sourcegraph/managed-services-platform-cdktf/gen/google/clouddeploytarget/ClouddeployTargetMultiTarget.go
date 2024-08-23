package clouddeploytarget


type ClouddeployTargetMultiTarget struct {
	// Required. The target_ids of this multiTarget.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_target#target_ids ClouddeployTarget#target_ids}
	TargetIds *[]*string `field:"required" json:"targetIds" yaml:"targetIds"`
}


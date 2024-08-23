package clouddeploycustomtargettype


type ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepo struct {
	// Cloud Build 2nd gen repository in the format of 'projects/<project>/locations/<location>/connections/<connection>/repositories/<repository>'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#repository ClouddeployCustomTargetType#repository}
	Repository *string `field:"required" json:"repository" yaml:"repository"`
	// Relative path from the repository root to the Skaffold file.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#path ClouddeployCustomTargetType#path}
	Path *string `field:"optional" json:"path" yaml:"path"`
	// Branch or tag to use when cloning the repository.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#ref ClouddeployCustomTargetType#ref}
	Ref *string `field:"optional" json:"ref" yaml:"ref"`
}


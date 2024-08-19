package clouddeploycustomtargettype


type ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGit struct {
	// Git repository the package should be cloned from.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#repo ClouddeployCustomTargetType#repo}
	Repo *string `field:"required" json:"repo" yaml:"repo"`
	// Relative path from the repository root to the Skaffold file.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#path ClouddeployCustomTargetType#path}
	Path *string `field:"optional" json:"path" yaml:"path"`
	// Git ref the package should be cloned from.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#ref ClouddeployCustomTargetType#ref}
	Ref *string `field:"optional" json:"ref" yaml:"ref"`
}


package clouddeploycustomtargettype


type ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModules struct {
	// The Skaffold Config modules to use from the specified source.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#configs ClouddeployCustomTargetType#configs}
	Configs *[]*string `field:"optional" json:"configs" yaml:"configs"`
	// git block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#git ClouddeployCustomTargetType#git}
	Git *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGit `field:"optional" json:"git" yaml:"git"`
	// google_cloud_build_repo block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#google_cloud_build_repo ClouddeployCustomTargetType#google_cloud_build_repo}
	GoogleCloudBuildRepo *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepo `field:"optional" json:"googleCloudBuildRepo" yaml:"googleCloudBuildRepo"`
	// google_cloud_storage block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#google_cloud_storage ClouddeployCustomTargetType#google_cloud_storage}
	GoogleCloudStorage *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorage `field:"optional" json:"googleCloudStorage" yaml:"googleCloudStorage"`
}


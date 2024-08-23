package clouddeploycustomtargettype


type ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorage struct {
	// Cloud Storage source paths to copy recursively.
	//
	// For example, providing 'gs://my-bucket/dir/configs/*' will result in Skaffold copying all files within the 'dir/configs' directory in the bucket 'my-bucket'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#source ClouddeployCustomTargetType#source}
	Source *string `field:"required" json:"source" yaml:"source"`
	// Relative path from the source to the Skaffold file.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#path ClouddeployCustomTargetType#path}
	Path *string `field:"optional" json:"path" yaml:"path"`
}


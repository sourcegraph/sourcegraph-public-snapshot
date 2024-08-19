package clouddeploycustomtargettype


type ClouddeployCustomTargetTypeCustomActions struct {
	// The Skaffold custom action responsible for deploy operations.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#deploy_action ClouddeployCustomTargetType#deploy_action}
	DeployAction *string `field:"required" json:"deployAction" yaml:"deployAction"`
	// include_skaffold_modules block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#include_skaffold_modules ClouddeployCustomTargetType#include_skaffold_modules}
	IncludeSkaffoldModules interface{} `field:"optional" json:"includeSkaffoldModules" yaml:"includeSkaffoldModules"`
	// The Skaffold custom action responsible for render operations.
	//
	// If not provided then Cloud Deploy will perform the render operations via 'skaffold render'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_custom_target_type#render_action ClouddeployCustomTargetType#render_action}
	RenderAction *string `field:"optional" json:"renderAction" yaml:"renderAction"`
}


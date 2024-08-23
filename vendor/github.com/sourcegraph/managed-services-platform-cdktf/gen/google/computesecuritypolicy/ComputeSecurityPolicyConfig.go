package computesecuritypolicy

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ComputeSecurityPolicyConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// The name of the security policy.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#name ComputeSecurityPolicy#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// adaptive_protection_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#adaptive_protection_config ComputeSecurityPolicy#adaptive_protection_config}
	AdaptiveProtectionConfig *ComputeSecurityPolicyAdaptiveProtectionConfig `field:"optional" json:"adaptiveProtectionConfig" yaml:"adaptiveProtectionConfig"`
	// advanced_options_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#advanced_options_config ComputeSecurityPolicy#advanced_options_config}
	AdvancedOptionsConfig *ComputeSecurityPolicyAdvancedOptionsConfig `field:"optional" json:"advancedOptionsConfig" yaml:"advancedOptionsConfig"`
	// An optional description of this security policy. Max size is 2048.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#description ComputeSecurityPolicy#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#id ComputeSecurityPolicy#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The project in which the resource belongs. If it is not provided, the provider project is used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#project ComputeSecurityPolicy#project}
	Project *string `field:"optional" json:"project" yaml:"project"`
	// recaptcha_options_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#recaptcha_options_config ComputeSecurityPolicy#recaptcha_options_config}
	RecaptchaOptionsConfig *ComputeSecurityPolicyRecaptchaOptionsConfig `field:"optional" json:"recaptchaOptionsConfig" yaml:"recaptchaOptionsConfig"`
	// rule block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#rule ComputeSecurityPolicy#rule}
	Rule interface{} `field:"optional" json:"rule" yaml:"rule"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#timeouts ComputeSecurityPolicy#timeouts}
	Timeouts *ComputeSecurityPolicyTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// The type indicates the intended use of the security policy.
	//
	// CLOUD_ARMOR - Cloud Armor backend security policies can be configured to filter incoming HTTP requests targeting backend services. They filter requests before they hit the origin servers. CLOUD_ARMOR_EDGE - Cloud Armor edge security policies can be configured to filter incoming HTTP requests targeting backend services (including Cloud CDN-enabled) as well as backend buckets (Cloud Storage). They filter requests before the request is served from Google's cache.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#type ComputeSecurityPolicy#type}
	Type *string `field:"optional" json:"type" yaml:"type"`
}


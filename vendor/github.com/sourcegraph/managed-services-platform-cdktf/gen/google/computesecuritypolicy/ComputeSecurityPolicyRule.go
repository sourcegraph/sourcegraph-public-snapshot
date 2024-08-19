package computesecuritypolicy


type ComputeSecurityPolicyRule struct {
	// Action to take when match matches the request.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#action ComputeSecurityPolicy#action}
	Action *string `field:"required" json:"action" yaml:"action"`
	// match block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#match ComputeSecurityPolicy#match}
	Match *ComputeSecurityPolicyRuleMatch `field:"required" json:"match" yaml:"match"`
	// An unique positive integer indicating the priority of evaluation for a rule.
	//
	// Rules are evaluated from highest priority (lowest numerically) to lowest priority (highest numerically) in order.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#priority ComputeSecurityPolicy#priority}
	Priority *float64 `field:"required" json:"priority" yaml:"priority"`
	// An optional description of this rule. Max size is 64.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#description ComputeSecurityPolicy#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// header_action block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#header_action ComputeSecurityPolicy#header_action}
	HeaderAction *ComputeSecurityPolicyRuleHeaderAction `field:"optional" json:"headerAction" yaml:"headerAction"`
	// When set to true, the action specified above is not enforced.
	//
	// Stackdriver logs for requests that trigger a preview action are annotated as such.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#preview ComputeSecurityPolicy#preview}
	Preview interface{} `field:"optional" json:"preview" yaml:"preview"`
	// rate_limit_options block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#rate_limit_options ComputeSecurityPolicy#rate_limit_options}
	RateLimitOptions *ComputeSecurityPolicyRuleRateLimitOptions `field:"optional" json:"rateLimitOptions" yaml:"rateLimitOptions"`
	// redirect_options block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#redirect_options ComputeSecurityPolicy#redirect_options}
	RedirectOptions *ComputeSecurityPolicyRuleRedirectOptions `field:"optional" json:"redirectOptions" yaml:"redirectOptions"`
}


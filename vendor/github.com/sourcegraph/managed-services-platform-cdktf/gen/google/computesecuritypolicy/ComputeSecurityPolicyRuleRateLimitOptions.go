package computesecuritypolicy


type ComputeSecurityPolicyRuleRateLimitOptions struct {
	// Action to take for requests that are under the configured rate limit threshold. Valid option is "allow" only.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#conform_action ComputeSecurityPolicy#conform_action}
	ConformAction *string `field:"required" json:"conformAction" yaml:"conformAction"`
	// Action to take for requests that are above the configured rate limit threshold, to either deny with a specified HTTP response code, or redirect to a different endpoint.
	//
	// Valid options are "deny()" where valid values for status are 403, 404, 429, and 502, and "redirect" where the redirect parameters come from exceedRedirectOptions below.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#exceed_action ComputeSecurityPolicy#exceed_action}
	ExceedAction *string `field:"required" json:"exceedAction" yaml:"exceedAction"`
	// rate_limit_threshold block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#rate_limit_threshold ComputeSecurityPolicy#rate_limit_threshold}
	RateLimitThreshold *ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThreshold `field:"required" json:"rateLimitThreshold" yaml:"rateLimitThreshold"`
	// Can only be specified if the action for the rule is "rate_based_ban".
	//
	// If specified, determines the time (in seconds) the traffic will continue to be banned by the rate limit after the rate falls below the threshold.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#ban_duration_sec ComputeSecurityPolicy#ban_duration_sec}
	BanDurationSec *float64 `field:"optional" json:"banDurationSec" yaml:"banDurationSec"`
	// ban_threshold block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#ban_threshold ComputeSecurityPolicy#ban_threshold}
	BanThreshold *ComputeSecurityPolicyRuleRateLimitOptionsBanThreshold `field:"optional" json:"banThreshold" yaml:"banThreshold"`
	// Determines the key to enforce the rateLimitThreshold on.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#enforce_on_key ComputeSecurityPolicy#enforce_on_key}
	EnforceOnKey *string `field:"optional" json:"enforceOnKey" yaml:"enforceOnKey"`
	// Rate limit key name applicable only for the following key types: HTTP_HEADER -- Name of the HTTP header whose value is taken as the key value.
	//
	// HTTP_COOKIE -- Name of the HTTP cookie whose value is taken as the key value.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#enforce_on_key_name ComputeSecurityPolicy#enforce_on_key_name}
	EnforceOnKeyName *string `field:"optional" json:"enforceOnKeyName" yaml:"enforceOnKeyName"`
	// exceed_redirect_options block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#exceed_redirect_options ComputeSecurityPolicy#exceed_redirect_options}
	ExceedRedirectOptions *ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptions `field:"optional" json:"exceedRedirectOptions" yaml:"exceedRedirectOptions"`
}


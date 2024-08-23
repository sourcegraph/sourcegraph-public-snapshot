package computesecuritypolicy


type ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThreshold struct {
	// Number of HTTP(S) requests for calculating the threshold.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#count ComputeSecurityPolicy#count}
	Count *float64 `field:"required" json:"count" yaml:"count"`
	// Interval over which the threshold is computed.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#interval_sec ComputeSecurityPolicy#interval_sec}
	IntervalSec *float64 `field:"required" json:"intervalSec" yaml:"intervalSec"`
}


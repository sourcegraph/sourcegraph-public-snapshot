package computesecuritypolicy


type ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfig struct {
	// A list of custom Content-Type header values to apply the JSON parsing.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#content_types ComputeSecurityPolicy#content_types}
	ContentTypes *[]*string `field:"required" json:"contentTypes" yaml:"contentTypes"`
}


package computesecuritypolicy


type ComputeSecurityPolicyRuleHeaderActionRequestHeadersToAdds struct {
	// The name of the header to set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#header_name ComputeSecurityPolicy#header_name}
	HeaderName *string `field:"required" json:"headerName" yaml:"headerName"`
	// The value to set the named header to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#header_value ComputeSecurityPolicy#header_value}
	HeaderValue *string `field:"optional" json:"headerValue" yaml:"headerValue"`
}


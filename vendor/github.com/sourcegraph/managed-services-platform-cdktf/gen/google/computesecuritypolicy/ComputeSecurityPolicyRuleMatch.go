package computesecuritypolicy


type ComputeSecurityPolicyRuleMatch struct {
	// config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#config ComputeSecurityPolicy#config}
	Config *ComputeSecurityPolicyRuleMatchConfig `field:"optional" json:"config" yaml:"config"`
	// expr block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#expr ComputeSecurityPolicy#expr}
	Expr *ComputeSecurityPolicyRuleMatchExpr `field:"optional" json:"expr" yaml:"expr"`
	// Predefined rule expression.
	//
	// If this field is specified, config must also be specified. Available options:   SRC_IPS_V1: Must specify the corresponding src_ip_ranges field in config.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#versioned_expr ComputeSecurityPolicy#versioned_expr}
	VersionedExpr *string `field:"optional" json:"versionedExpr" yaml:"versionedExpr"`
}


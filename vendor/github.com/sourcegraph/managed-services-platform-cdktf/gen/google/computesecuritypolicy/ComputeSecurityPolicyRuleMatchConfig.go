package computesecuritypolicy


type ComputeSecurityPolicyRuleMatchConfig struct {
	// Set of IP addresses or ranges (IPV4 or IPV6) in CIDR notation to match against inbound traffic.
	//
	// There is a limit of 10 IP ranges per rule. A value of '*' matches all IPs (can be used to override the default behavior).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_security_policy#src_ip_ranges ComputeSecurityPolicy#src_ip_ranges}
	SrcIpRanges *[]*string `field:"required" json:"srcIpRanges" yaml:"srcIpRanges"`
}


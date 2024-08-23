package computefirewall


type ComputeFirewallLogConfig struct {
	// This field denotes whether to include or exclude metadata for firewall logs. Possible values: ["EXCLUDE_ALL_METADATA", "INCLUDE_ALL_METADATA"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#metadata ComputeFirewall#metadata}
	Metadata *string `field:"required" json:"metadata" yaml:"metadata"`
}


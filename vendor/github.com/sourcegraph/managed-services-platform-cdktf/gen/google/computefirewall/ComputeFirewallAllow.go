package computefirewall


type ComputeFirewallAllow struct {
	// The IP protocol to which this rule applies.
	//
	// The protocol type is
	// required when creating a firewall rule. This value can either be
	// one of the following well known protocol strings (tcp, udp,
	// icmp, esp, ah, sctp, ipip, all), or the IP protocol number.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#protocol ComputeFirewall#protocol}
	Protocol *string `field:"required" json:"protocol" yaml:"protocol"`
	// An optional list of ports to which this rule applies.
	//
	// This field
	// is only applicable for UDP or TCP protocol. Each entry must be
	// either an integer or a range. If not specified, this rule
	// applies to connections through any port.
	//
	// Example inputs include: ["22"], ["80","443"], and
	// ["12345-12349"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#ports ComputeFirewall#ports}
	Ports *[]*string `field:"optional" json:"ports" yaml:"ports"`
}


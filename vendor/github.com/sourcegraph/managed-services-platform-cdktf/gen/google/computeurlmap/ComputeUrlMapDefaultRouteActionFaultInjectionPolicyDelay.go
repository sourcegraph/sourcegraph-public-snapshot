package computeurlmap


type ComputeUrlMapDefaultRouteActionFaultInjectionPolicyDelay struct {
	// fixed_delay block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#fixed_delay ComputeUrlMap#fixed_delay}
	FixedDelay *ComputeUrlMapDefaultRouteActionFaultInjectionPolicyDelayFixedDelay `field:"optional" json:"fixedDelay" yaml:"fixedDelay"`
	// The percentage of traffic (connections/operations/requests) on which delay will be introduced as part of fault injection.
	//
	// The value must be between 0.0 and 100.0 inclusive.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#percentage ComputeUrlMap#percentage}
	Percentage *float64 `field:"optional" json:"percentage" yaml:"percentage"`
}


package computeglobalforwardingrule


type ComputeGlobalForwardingRuleTimeouts struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#create ComputeGlobalForwardingRule#create}.
	Create *string `field:"optional" json:"create" yaml:"create"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#delete ComputeGlobalForwardingRule#delete}.
	Delete *string `field:"optional" json:"delete" yaml:"delete"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#update ComputeGlobalForwardingRule#update}.
	Update *string `field:"optional" json:"update" yaml:"update"`
}


package computeglobalforwardingrule


type ComputeGlobalForwardingRuleServiceDirectoryRegistrations struct {
	// Service Directory namespace to register the forwarding rule under.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#namespace ComputeGlobalForwardingRule#namespace}
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
	// [Optional] Service Directory region to register this global forwarding rule under.
	//
	// Default to "us-central1". Only used for PSC for Google APIs. All PSC for
	// Google APIs Forwarding Rules on the same network should use the same Service
	// Directory region.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#service_directory_region ComputeGlobalForwardingRule#service_directory_region}
	ServiceDirectoryRegion *string `field:"optional" json:"serviceDirectoryRegion" yaml:"serviceDirectoryRegion"`
}


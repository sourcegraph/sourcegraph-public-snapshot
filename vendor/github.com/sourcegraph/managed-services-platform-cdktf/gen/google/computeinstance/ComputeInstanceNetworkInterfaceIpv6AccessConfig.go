package computeinstance


type ComputeInstanceNetworkInterfaceIpv6AccessConfig struct {
	// The service-level to be provided for IPv6 traffic when the subnet has an external subnet.
	//
	// Only PREMIUM tier is valid for IPv6
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#network_tier ComputeInstance#network_tier}
	NetworkTier *string `field:"required" json:"networkTier" yaml:"networkTier"`
	// The first IPv6 address of the external IPv6 range associated with this instance, prefix length is stored in externalIpv6PrefixLength in ipv6AccessConfig.
	//
	// To use a static external IP address, it must be unused and in the same region as the instance's zone. If not specified, Google Cloud will automatically assign an external IPv6 address from the instance's subnetwork.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#external_ipv6 ComputeInstance#external_ipv6}
	ExternalIpv6 *string `field:"optional" json:"externalIpv6" yaml:"externalIpv6"`
	// The prefix length of the external IPv6 range.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#external_ipv6_prefix_length ComputeInstance#external_ipv6_prefix_length}
	ExternalIpv6PrefixLength *string `field:"optional" json:"externalIpv6PrefixLength" yaml:"externalIpv6PrefixLength"`
	// The name of this access configuration. In ipv6AccessConfigs, the recommended name is External IPv6.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#name ComputeInstance#name}
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The domain name to be used when creating DNSv6 records for the external IPv6 ranges.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#public_ptr_domain_name ComputeInstance#public_ptr_domain_name}
	PublicPtrDomainName *string `field:"optional" json:"publicPtrDomainName" yaml:"publicPtrDomainName"`
}


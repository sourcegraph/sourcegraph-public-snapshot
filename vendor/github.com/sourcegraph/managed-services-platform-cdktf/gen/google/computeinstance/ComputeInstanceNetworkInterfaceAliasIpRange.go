package computeinstance


type ComputeInstanceNetworkInterfaceAliasIpRange struct {
	// The IP CIDR range represented by this alias IP range.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#ip_cidr_range ComputeInstance#ip_cidr_range}
	IpCidrRange *string `field:"required" json:"ipCidrRange" yaml:"ipCidrRange"`
	// The subnetwork secondary range name specifying the secondary range from which to allocate the IP CIDR range for this alias IP range.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#subnetwork_range_name ComputeInstance#subnetwork_range_name}
	SubnetworkRangeName *string `field:"optional" json:"subnetworkRangeName" yaml:"subnetworkRangeName"`
}


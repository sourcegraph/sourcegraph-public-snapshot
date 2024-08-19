package computeinstance


type ComputeInstanceNetworkInterface struct {
	// access_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#access_config ComputeInstance#access_config}
	AccessConfig interface{} `field:"optional" json:"accessConfig" yaml:"accessConfig"`
	// alias_ip_range block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#alias_ip_range ComputeInstance#alias_ip_range}
	AliasIpRange interface{} `field:"optional" json:"aliasIpRange" yaml:"aliasIpRange"`
	// The prefix length of the primary internal IPv6 range.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#internal_ipv6_prefix_length ComputeInstance#internal_ipv6_prefix_length}
	InternalIpv6PrefixLength *float64 `field:"optional" json:"internalIpv6PrefixLength" yaml:"internalIpv6PrefixLength"`
	// ipv6_access_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#ipv6_access_config ComputeInstance#ipv6_access_config}
	Ipv6AccessConfig interface{} `field:"optional" json:"ipv6AccessConfig" yaml:"ipv6AccessConfig"`
	// An IPv6 internal network address for this network interface.
	//
	// If not specified, Google Cloud will automatically assign an internal IPv6 address from the instance's subnetwork.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#ipv6_address ComputeInstance#ipv6_address}
	Ipv6Address *string `field:"optional" json:"ipv6Address" yaml:"ipv6Address"`
	// The name or self_link of the network attached to this interface.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#network ComputeInstance#network}
	Network *string `field:"optional" json:"network" yaml:"network"`
	// The private IP address assigned to the instance.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#network_ip ComputeInstance#network_ip}
	NetworkIp *string `field:"optional" json:"networkIp" yaml:"networkIp"`
	// The type of vNIC to be used on this interface. Possible values:GVNIC, VIRTIO_NET.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#nic_type ComputeInstance#nic_type}
	NicType *string `field:"optional" json:"nicType" yaml:"nicType"`
	// The networking queue count that's specified by users for the network interface.
	//
	// Both Rx and Tx queues will be set to this number. It will be empty if not specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#queue_count ComputeInstance#queue_count}
	QueueCount *float64 `field:"optional" json:"queueCount" yaml:"queueCount"`
	// The stack type for this network interface to identify whether the IPv6 feature is enabled or not.
	//
	// If not specified, IPV4_ONLY will be used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#stack_type ComputeInstance#stack_type}
	StackType *string `field:"optional" json:"stackType" yaml:"stackType"`
	// The name or self_link of the subnetwork attached to this interface.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#subnetwork ComputeInstance#subnetwork}
	Subnetwork *string `field:"optional" json:"subnetwork" yaml:"subnetwork"`
	// The project in which the subnetwork belongs.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance#subnetwork_project ComputeInstance#subnetwork_project}
	SubnetworkProject *string `field:"optional" json:"subnetworkProject" yaml:"subnetworkProject"`
}


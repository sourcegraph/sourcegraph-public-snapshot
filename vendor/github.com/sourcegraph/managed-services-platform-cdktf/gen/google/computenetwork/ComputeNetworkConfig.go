package computenetwork

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ComputeNetworkConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// Name of the resource.
	//
	// Provided by the client when the resource is
	// created. The name must be 1-63 characters long, and comply with
	// RFC1035. Specifically, the name must be 1-63 characters long and match
	// the regular expression '[a-z]([-a-z0-9]*[a-z0-9])?' which means the
	// first character must be a lowercase letter, and all following
	// characters must be a dash, lowercase letter, or digit, except the last
	// character, which cannot be a dash.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#name ComputeNetwork#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// When set to 'true', the network is created in "auto subnet mode" and it will create a subnet for each region automatically across the '10.128.0.0/9' address range.
	//
	// When set to 'false', the network is created in "custom subnet mode" so
	// the user can explicitly connect subnetwork resources.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#auto_create_subnetworks ComputeNetwork#auto_create_subnetworks}
	AutoCreateSubnetworks interface{} `field:"optional" json:"autoCreateSubnetworks" yaml:"autoCreateSubnetworks"`
	// If set to 'true', default routes ('0.0.0.0/0') will be deleted immediately after network creation. Defaults to 'false'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#delete_default_routes_on_create ComputeNetwork#delete_default_routes_on_create}
	DeleteDefaultRoutesOnCreate interface{} `field:"optional" json:"deleteDefaultRoutesOnCreate" yaml:"deleteDefaultRoutesOnCreate"`
	// An optional description of this resource. The resource must be recreated to modify this field.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#description ComputeNetwork#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Enable ULA internal ipv6 on this network. Enabling this feature will assign a /48 from google defined ULA prefix fd20::/20.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#enable_ula_internal_ipv6 ComputeNetwork#enable_ula_internal_ipv6}
	EnableUlaInternalIpv6 interface{} `field:"optional" json:"enableUlaInternalIpv6" yaml:"enableUlaInternalIpv6"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#id ComputeNetwork#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// When enabling ula internal ipv6, caller optionally can specify the /48 range they want from the google defined ULA prefix fd20::/20.
	//
	// The input must be a
	// valid /48 ULA IPv6 address and must be within the fd20::/20. Operation will
	// fail if the speficied /48 is already in used by another resource.
	// If the field is not speficied, then a /48 range will be randomly allocated from fd20::/20 and returned via this field.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#internal_ipv6_range ComputeNetwork#internal_ipv6_range}
	InternalIpv6Range *string `field:"optional" json:"internalIpv6Range" yaml:"internalIpv6Range"`
	// Maximum Transmission Unit in bytes.
	//
	// The default value is 1460 bytes.
	// The minimum value for this field is 1300 and the maximum value is 8896 bytes (jumbo frames).
	// Note that packets larger than 1500 bytes (standard Ethernet) can be subject to TCP-MSS clamping or dropped
	// with an ICMP 'Fragmentation-Needed' message if the packets are routed to the Internet or other VPCs
	// with varying MTUs.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#mtu ComputeNetwork#mtu}
	Mtu *float64 `field:"optional" json:"mtu" yaml:"mtu"`
	// Set the order that Firewall Rules and Firewall Policies are evaluated. Default value: "AFTER_CLASSIC_FIREWALL" Possible values: ["BEFORE_CLASSIC_FIREWALL", "AFTER_CLASSIC_FIREWALL"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#network_firewall_policy_enforcement_order ComputeNetwork#network_firewall_policy_enforcement_order}
	NetworkFirewallPolicyEnforcementOrder *string `field:"optional" json:"networkFirewallPolicyEnforcementOrder" yaml:"networkFirewallPolicyEnforcementOrder"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#project ComputeNetwork#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// The network-wide routing mode to use.
	//
	// If set to 'REGIONAL', this
	// network's cloud routers will only advertise routes with subnetworks
	// of this network in the same region as the router. If set to 'GLOBAL',
	// this network's cloud routers will advertise routes with all
	// subnetworks of this network, across regions. Possible values: ["REGIONAL", "GLOBAL"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#routing_mode ComputeNetwork#routing_mode}
	RoutingMode *string `field:"optional" json:"routingMode" yaml:"routingMode"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network#timeouts ComputeNetwork#timeouts}
	Timeouts *ComputeNetworkTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}


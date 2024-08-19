package computeglobalforwardingrule

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ComputeGlobalForwardingRuleConfig struct {
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
	// Name of the resource;
	//
	// provided by the client when the resource is created.
	// The name must be 1-63 characters long, and comply with
	// [RFC1035](https://www.ietf.org/rfc/rfc1035.txt).
	//
	// Specifically, the name must be 1-63 characters long and match the regular
	// expression '[a-z]([-a-z0-9]*[a-z0-9])?' which means the first
	// character must be a lowercase letter, and all following characters must
	// be a dash, lowercase letter, or digit, except the last character, which
	// cannot be a dash.
	//
	// For Private Service Connect forwarding rules that forward traffic to Google
	// APIs, the forwarding rule name must be a 1-20 characters string with
	// lowercase letters and numbers and must start with a letter.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#name ComputeGlobalForwardingRule#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The URL of the target resource to receive the matched traffic.
	//
	// For
	// regional forwarding rules, this target must be in the same region as the
	// forwarding rule. For global forwarding rules, this target must be a global
	// load balancing resource.
	//
	// The forwarded traffic must be of a type appropriate to the target object.
	// For load balancers, see the "Target" column in [Port specifications](https://cloud.google.com/load-balancing/docs/forwarding-rule-concepts#ip_address_specifications).
	// For Private Service Connect forwarding rules that forward traffic to Google APIs, provide the name of a supported Google API bundle:
	//  'vpc-sc' - [ APIs that support VPC Service Controls](https://cloud.google.com/vpc-service-controls/docs/supported-products).
	//  'all-apis' - [All supported Google APIs](https://cloud.google.com/vpc/docs/private-service-connect#supported-apis).
	//
	//
	// For Private Service Connect forwarding rules that forward traffic to managed services, the target must be a service attachment.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#target ComputeGlobalForwardingRule#target}
	Target *string `field:"required" json:"target" yaml:"target"`
	// An optional description of this resource. Provide this property when you create the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#description ComputeGlobalForwardingRule#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#id ComputeGlobalForwardingRule#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// IP address for which this forwarding rule accepts traffic.
	//
	// When a client
	// sends traffic to this IP address, the forwarding rule directs the traffic
	// to the referenced 'target'.
	//
	// While creating a forwarding rule, specifying an 'IPAddress' is
	// required under the following circumstances:
	//
	// When the 'target' is set to 'targetGrpcProxy' and
	// 'validateForProxyless' is set to 'true', the
	// 'IPAddress' should be set to '0.0.0.0'.
	// When the 'target' is a Private Service Connect Google APIs
	// bundle, you must specify an 'IPAddress'.
	//
	//
	// Otherwise, you can optionally specify an IP address that references an
	// existing static (reserved) IP address resource. When omitted, Google Cloud
	// assigns an ephemeral IP address.
	//
	// Use one of the following formats to specify an IP address while creating a
	// forwarding rule:
	//
	// IP address number, as in '100.1.2.3'
	// IPv6 address range, as in '2600:1234::/96'
	// Full resource URL, as in
	// 'https://www.googleapis.com/compute/v1/projects/project_id/regions/region/addresses/address-name'
	// Partial URL or by name, as in:
	// 'projects/project_id/regions/region/addresses/address-name'
	// 'regions/region/addresses/address-name'
	// 'global/addresses/address-name'
	// 'address-name'
	//
	//
	// The forwarding rule's 'target',
	// and in most cases, also the 'loadBalancingScheme', determine the
	// type of IP address that you can use. For detailed information, see
	// [IP address
	// specifications](https://cloud.google.com/load-balancing/docs/forwarding-rule-concepts#ip_address_specifications).
	//
	// When reading an 'IPAddress', the API always returns the IP
	// address number.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#ip_address ComputeGlobalForwardingRule#ip_address}
	IpAddress *string `field:"optional" json:"ipAddress" yaml:"ipAddress"`
	// The IP protocol to which this rule applies.
	//
	// For protocol forwarding, valid
	// options are 'TCP', 'UDP', 'ESP',
	// 'AH', 'SCTP', 'ICMP' and
	// 'L3_DEFAULT'.
	//
	// The valid IP protocols are different for different load balancing products
	// as described in [Load balancing
	// features](https://cloud.google.com/load-balancing/docs/features#protocols_from_the_load_balancer_to_the_backends). Possible values: ["TCP", "UDP", "ESP", "AH", "SCTP", "ICMP"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#ip_protocol ComputeGlobalForwardingRule#ip_protocol}
	IpProtocol *string `field:"optional" json:"ipProtocol" yaml:"ipProtocol"`
	// The IP Version that will be used by this global forwarding rule. Possible values: ["IPV4", "IPV6"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#ip_version ComputeGlobalForwardingRule#ip_version}
	IpVersion *string `field:"optional" json:"ipVersion" yaml:"ipVersion"`
	// Labels to apply to this forwarding rule.  A list of key->value pairs.
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#labels ComputeGlobalForwardingRule#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// Specifies the forwarding rule type.
	//
	// For more information about forwarding rules, refer to
	// [Forwarding rule concepts](https://cloud.google.com/load-balancing/docs/forwarding-rule-concepts). Default value: "EXTERNAL" Possible values: ["EXTERNAL", "EXTERNAL_MANAGED", "INTERNAL_MANAGED", "INTERNAL_SELF_MANAGED"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#load_balancing_scheme ComputeGlobalForwardingRule#load_balancing_scheme}
	LoadBalancingScheme *string `field:"optional" json:"loadBalancingScheme" yaml:"loadBalancingScheme"`
	// metadata_filters block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#metadata_filters ComputeGlobalForwardingRule#metadata_filters}
	MetadataFilters interface{} `field:"optional" json:"metadataFilters" yaml:"metadataFilters"`
	// This field is not used for external load balancing.
	//
	// For Internal TCP/UDP Load Balancing, this field identifies the network that
	// the load balanced IP should belong to for this Forwarding Rule.
	// If the subnetwork is specified, the network of the subnetwork will be used.
	// If neither subnetwork nor this field is specified, the default network will
	// be used.
	//
	// For Private Service Connect forwarding rules that forward traffic to Google
	// APIs, a network must be provided.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#network ComputeGlobalForwardingRule#network}
	Network *string `field:"optional" json:"network" yaml:"network"`
	// This is used in PSC consumer ForwardingRule to control whether it should try to auto-generate a DNS zone or not.
	//
	// Non-PSC forwarding rules do not use this field.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#no_automate_dns_zone ComputeGlobalForwardingRule#no_automate_dns_zone}
	NoAutomateDnsZone interface{} `field:"optional" json:"noAutomateDnsZone" yaml:"noAutomateDnsZone"`
	// The 'portRange' field has the following limitations: It requires that the forwarding rule 'IPProtocol' be TCP, UDP, or SCTP, and It's applicable only to the following products: external passthrough Network Load Balancers, internal and external proxy Network Load Balancers, internal and external Application Load Balancers, external protocol forwarding, and Classic VPN.
	//
	// Some products have restrictions on what ports can be used. See
	// [port specifications](https://cloud.google.com/load-balancing/docs/forwarding-rule-concepts#port_specifications)
	// for details.
	//
	// For external forwarding rules, two or more forwarding rules cannot use the
	// same '[IPAddress, IPProtocol]' pair, and cannot have overlapping
	// 'portRange's.
	//
	// For internal forwarding rules within the same VPC network, two or more
	// forwarding rules cannot use the same '[IPAddress, IPProtocol]' pair, and
	// cannot have overlapping 'portRange's.
	PortRange *string `field:"optional" json:"portRange" yaml:"portRange"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#project ComputeGlobalForwardingRule#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// service_directory_registrations block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#service_directory_registrations ComputeGlobalForwardingRule#service_directory_registrations}
	ServiceDirectoryRegistrations *ComputeGlobalForwardingRuleServiceDirectoryRegistrations `field:"optional" json:"serviceDirectoryRegistrations" yaml:"serviceDirectoryRegistrations"`
	// If not empty, this Forwarding Rule will only forward the traffic when the source IP address matches one of the IP addresses or CIDR ranges set here.
	//
	// Note that a Forwarding Rule can only have up to 64 source IP ranges, and this field can only be used with a regional Forwarding Rule whose scheme is EXTERNAL. Each sourceIpRange entry should be either an IP address (for example, 1.2.3.4) or a CIDR range (for example, 1.2.3.0/24).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#source_ip_ranges ComputeGlobalForwardingRule#source_ip_ranges}
	SourceIpRanges *[]*string `field:"optional" json:"sourceIpRanges" yaml:"sourceIpRanges"`
	// This field identifies the subnetwork that the load balanced IP should belong to for this Forwarding Rule, used in internal load balancing and network load balancing with IPv6.
	//
	// If the network specified is in auto subnet mode, this field is optional.
	// However, a subnetwork must be specified if the network is in custom subnet
	// mode or when creating external forwarding rule with IPv6.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#subnetwork ComputeGlobalForwardingRule#subnetwork}
	Subnetwork *string `field:"optional" json:"subnetwork" yaml:"subnetwork"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_global_forwarding_rule#timeouts ComputeGlobalForwardingRule#timeouts}
	Timeouts *ComputeGlobalForwardingRuleTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}


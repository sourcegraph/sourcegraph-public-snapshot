package computefirewall

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ComputeFirewallConfig struct {
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
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#name ComputeFirewall#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The name or self_link of the network to attach this firewall to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#network ComputeFirewall#network}
	Network *string `field:"required" json:"network" yaml:"network"`
	// allow block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#allow ComputeFirewall#allow}
	Allow interface{} `field:"optional" json:"allow" yaml:"allow"`
	// deny block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#deny ComputeFirewall#deny}
	Deny interface{} `field:"optional" json:"deny" yaml:"deny"`
	// An optional description of this resource. Provide this property when you create the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#description ComputeFirewall#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// If destination ranges are specified, the firewall will apply only to traffic that has destination IP address in these ranges.
	//
	// These ranges
	// must be expressed in CIDR format. IPv4 or IPv6 ranges are supported.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#destination_ranges ComputeFirewall#destination_ranges}
	DestinationRanges *[]*string `field:"optional" json:"destinationRanges" yaml:"destinationRanges"`
	// Direction of traffic to which this firewall applies;
	//
	// default is
	// INGRESS. Note: For INGRESS traffic, one of 'source_ranges',
	// 'source_tags' or 'source_service_accounts' is required. Possible values: ["INGRESS", "EGRESS"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#direction ComputeFirewall#direction}
	Direction *string `field:"optional" json:"direction" yaml:"direction"`
	// Denotes whether the firewall rule is disabled, i.e not applied to the network it is associated with. When set to true, the firewall rule is not enforced and the network behaves as if it did not exist. If this is unspecified, the firewall rule will be enabled.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#disabled ComputeFirewall#disabled}
	Disabled interface{} `field:"optional" json:"disabled" yaml:"disabled"`
	// This field denotes whether to enable logging for a particular firewall rule.
	//
	// If logging is enabled, logs will be exported to Stackdriver.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#enable_logging ComputeFirewall#enable_logging}
	EnableLogging interface{} `field:"optional" json:"enableLogging" yaml:"enableLogging"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#id ComputeFirewall#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// log_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#log_config ComputeFirewall#log_config}
	LogConfig *ComputeFirewallLogConfig `field:"optional" json:"logConfig" yaml:"logConfig"`
	// Priority for this rule.
	//
	// This is an integer between 0 and 65535, both
	// inclusive. When not specified, the value assumed is 1000. Relative
	// priorities determine precedence of conflicting rules. Lower value of
	// priority implies higher precedence (eg, a rule with priority 0 has
	// higher precedence than a rule with priority 1). DENY rules take
	// precedence over ALLOW rules having equal priority.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#priority ComputeFirewall#priority}
	Priority *float64 `field:"optional" json:"priority" yaml:"priority"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#project ComputeFirewall#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// If source ranges are specified, the firewall will apply only to traffic that has source IP address in these ranges.
	//
	// These ranges must
	// be expressed in CIDR format. One or both of sourceRanges and
	// sourceTags may be set. If both properties are set, the firewall will
	// apply to traffic that has source IP address within sourceRanges OR the
	// source IP that belongs to a tag listed in the sourceTags property. The
	// connection does not need to match both properties for the firewall to
	// apply. IPv4 or IPv6 ranges are supported. For INGRESS traffic, one of
	// 'source_ranges', 'source_tags' or 'source_service_accounts' is required.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#source_ranges ComputeFirewall#source_ranges}
	SourceRanges *[]*string `field:"optional" json:"sourceRanges" yaml:"sourceRanges"`
	// If source service accounts are specified, the firewall will apply only to traffic originating from an instance with a service account in this list.
	//
	// Source service accounts cannot be used to control traffic to an
	// instance's external IP address because service accounts are associated
	// with an instance, not an IP address. sourceRanges can be set at the
	// same time as sourceServiceAccounts. If both are set, the firewall will
	// apply to traffic that has source IP address within sourceRanges OR the
	// source IP belongs to an instance with service account listed in
	// sourceServiceAccount. The connection does not need to match both
	// properties for the firewall to apply. sourceServiceAccounts cannot be
	// used at the same time as sourceTags or targetTags. For INGRESS traffic,
	// one of 'source_ranges', 'source_tags' or 'source_service_accounts' is required.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#source_service_accounts ComputeFirewall#source_service_accounts}
	SourceServiceAccounts *[]*string `field:"optional" json:"sourceServiceAccounts" yaml:"sourceServiceAccounts"`
	// If source tags are specified, the firewall will apply only to traffic with source IP that belongs to a tag listed in source tags.
	//
	// Source
	// tags cannot be used to control traffic to an instance's external IP
	// address. Because tags are associated with an instance, not an IP
	// address. One or both of sourceRanges and sourceTags may be set. If
	// both properties are set, the firewall will apply to traffic that has
	// source IP address within sourceRanges OR the source IP that belongs to
	// a tag listed in the sourceTags property. The connection does not need
	// to match both properties for the firewall to apply. For INGRESS traffic,
	// one of 'source_ranges', 'source_tags' or 'source_service_accounts' is required.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#source_tags ComputeFirewall#source_tags}
	SourceTags *[]*string `field:"optional" json:"sourceTags" yaml:"sourceTags"`
	// A list of service accounts indicating sets of instances located in the network that may make network connections as specified in allowed[].
	//
	// targetServiceAccounts cannot be used at the same time as targetTags or
	// sourceTags. If neither targetServiceAccounts nor targetTags are
	// specified, the firewall rule applies to all instances on the specified
	// network.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#target_service_accounts ComputeFirewall#target_service_accounts}
	TargetServiceAccounts *[]*string `field:"optional" json:"targetServiceAccounts" yaml:"targetServiceAccounts"`
	// A list of instance tags indicating sets of instances located in the network that may make network connections as specified in allowed[].
	//
	// If no targetTags are specified, the firewall rule applies to all
	// instances on the specified network.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#target_tags ComputeFirewall#target_tags}
	TargetTags *[]*string `field:"optional" json:"targetTags" yaml:"targetTags"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_firewall#timeouts ComputeFirewall#timeouts}
	Timeouts *ComputeFirewallTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}


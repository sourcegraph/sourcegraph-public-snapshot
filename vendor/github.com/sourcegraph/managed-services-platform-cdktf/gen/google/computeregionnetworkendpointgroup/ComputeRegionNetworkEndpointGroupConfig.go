package computeregionnetworkendpointgroup

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ComputeRegionNetworkEndpointGroupConfig struct {
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
	// provided by the client when the resource is
	// created. The name must be 1-63 characters long, and comply with
	// RFC1035. Specifically, the name must be 1-63 characters long and match
	// the regular expression '[a-z]([-a-z0-9]*[a-z0-9])?' which means the
	// first character must be a lowercase letter, and all following
	// characters must be a dash, lowercase letter, or digit, except the last
	// character, which cannot be a dash.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#name ComputeRegionNetworkEndpointGroup#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// A reference to the region where the regional NEGs reside.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#region ComputeRegionNetworkEndpointGroup#region}
	Region *string `field:"required" json:"region" yaml:"region"`
	// app_engine block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#app_engine ComputeRegionNetworkEndpointGroup#app_engine}
	AppEngine *ComputeRegionNetworkEndpointGroupAppEngine `field:"optional" json:"appEngine" yaml:"appEngine"`
	// cloud_function block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#cloud_function ComputeRegionNetworkEndpointGroup#cloud_function}
	CloudFunction *ComputeRegionNetworkEndpointGroupCloudFunction `field:"optional" json:"cloudFunction" yaml:"cloudFunction"`
	// cloud_run block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#cloud_run ComputeRegionNetworkEndpointGroup#cloud_run}
	CloudRun *ComputeRegionNetworkEndpointGroupCloudRun `field:"optional" json:"cloudRun" yaml:"cloudRun"`
	// An optional description of this resource. Provide this property when you create the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#description ComputeRegionNetworkEndpointGroup#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#id ComputeRegionNetworkEndpointGroup#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// This field is only used for PSC and INTERNET NEGs.
	//
	// The URL of the network to which all network endpoints in the NEG belong. Uses
	// "default" project network if unspecified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#network ComputeRegionNetworkEndpointGroup#network}
	Network *string `field:"optional" json:"network" yaml:"network"`
	// Type of network endpoints in this network endpoint group.
	//
	// Defaults to SERVERLESS. Default value: "SERVERLESS" Possible values: ["SERVERLESS", "PRIVATE_SERVICE_CONNECT", "INTERNET_IP_PORT", "INTERNET_FQDN_PORT"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#network_endpoint_type ComputeRegionNetworkEndpointGroup#network_endpoint_type}
	NetworkEndpointType *string `field:"optional" json:"networkEndpointType" yaml:"networkEndpointType"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#project ComputeRegionNetworkEndpointGroup#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// This field is only used for PSC and INTERNET NEGs.
	//
	// The target service url used to set up private service connection to
	// a Google API or a PSC Producer Service Attachment.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#psc_target_service ComputeRegionNetworkEndpointGroup#psc_target_service}
	PscTargetService *string `field:"optional" json:"pscTargetService" yaml:"pscTargetService"`
	// This field is only used for PSC NEGs.
	//
	// Optional URL of the subnetwork to which all network endpoints in the NEG belong.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#subnetwork ComputeRegionNetworkEndpointGroup#subnetwork}
	Subnetwork *string `field:"optional" json:"subnetwork" yaml:"subnetwork"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#timeouts ComputeRegionNetworkEndpointGroup#timeouts}
	Timeouts *ComputeRegionNetworkEndpointGroupTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}


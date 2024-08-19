package datastreamprivateconnection

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type DatastreamPrivateConnectionConfig struct {
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
	// Display name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#display_name DatastreamPrivateConnection#display_name}
	DisplayName *string `field:"required" json:"displayName" yaml:"displayName"`
	// The name of the location this private connection is located in.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#location DatastreamPrivateConnection#location}
	Location *string `field:"required" json:"location" yaml:"location"`
	// The private connectivity identifier.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#private_connection_id DatastreamPrivateConnection#private_connection_id}
	PrivateConnectionId *string `field:"required" json:"privateConnectionId" yaml:"privateConnectionId"`
	// vpc_peering_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#vpc_peering_config DatastreamPrivateConnection#vpc_peering_config}
	VpcPeeringConfig *DatastreamPrivateConnectionVpcPeeringConfig `field:"required" json:"vpcPeeringConfig" yaml:"vpcPeeringConfig"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#id DatastreamPrivateConnection#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Labels.
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#labels DatastreamPrivateConnection#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#project DatastreamPrivateConnection#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#timeouts DatastreamPrivateConnection#timeouts}
	Timeouts *DatastreamPrivateConnectionTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}


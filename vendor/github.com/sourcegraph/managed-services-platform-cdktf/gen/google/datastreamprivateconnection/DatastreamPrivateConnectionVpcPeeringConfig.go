package datastreamprivateconnection


type DatastreamPrivateConnectionVpcPeeringConfig struct {
	// A free subnet for peering. (CIDR of /29).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#subnet DatastreamPrivateConnection#subnet}
	Subnet *string `field:"required" json:"subnet" yaml:"subnet"`
	// Fully qualified name of the VPC that Datastream will peer to. Format: projects/{project}/global/{networks}/{name}.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_private_connection#vpc DatastreamPrivateConnection#vpc}
	Vpc *string `field:"required" json:"vpc" yaml:"vpc"`
}


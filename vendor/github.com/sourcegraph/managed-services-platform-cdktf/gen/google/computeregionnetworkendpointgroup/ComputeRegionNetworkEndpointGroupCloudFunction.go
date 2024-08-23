package computeregionnetworkendpointgroup


type ComputeRegionNetworkEndpointGroupCloudFunction struct {
	// A user-defined name of the Cloud Function. The function name is case-sensitive and must be 1-63 characters long. Example value: "func1".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#function ComputeRegionNetworkEndpointGroup#function}
	Function *string `field:"optional" json:"function" yaml:"function"`
	// A template to parse function field from a request URL.
	//
	// URL mask allows
	// for routing to multiple Cloud Functions without having to create
	// multiple Network Endpoint Groups and backend services.
	//
	// For example, request URLs "mydomain.com/function1" and "mydomain.com/function2"
	// can be backed by the same Serverless NEG with URL mask "/". The URL mask
	// will parse them to { function = "function1" } and { function = "function2" } respectively.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#url_mask ComputeRegionNetworkEndpointGroup#url_mask}
	UrlMask *string `field:"optional" json:"urlMask" yaml:"urlMask"`
}


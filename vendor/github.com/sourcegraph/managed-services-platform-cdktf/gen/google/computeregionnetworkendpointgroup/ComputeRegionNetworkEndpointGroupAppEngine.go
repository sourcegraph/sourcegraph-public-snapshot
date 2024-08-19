package computeregionnetworkendpointgroup


type ComputeRegionNetworkEndpointGroupAppEngine struct {
	// Optional serving service. The service name must be 1-63 characters long, and comply with RFC1035. Example value: "default", "my-service".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#service ComputeRegionNetworkEndpointGroup#service}
	Service *string `field:"optional" json:"service" yaml:"service"`
	// A template to parse service and version fields from a request URL.
	//
	// URL mask allows for routing to multiple App Engine services without
	// having to create multiple Network Endpoint Groups and backend services.
	//
	// For example, the request URLs "foo1-dot-appname.appspot.com/v1" and
	// "foo1-dot-appname.appspot.com/v2" can be backed by the same Serverless NEG with
	// URL mask "-dot-appname.appspot.com/". The URL mask will parse
	// them to { service = "foo1", version = "v1" } and { service = "foo1", version = "v2" } respectively.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#url_mask ComputeRegionNetworkEndpointGroup#url_mask}
	UrlMask *string `field:"optional" json:"urlMask" yaml:"urlMask"`
	// Optional serving version. The version must be 1-63 characters long, and comply with RFC1035. Example value: "v1", "v2".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#version ComputeRegionNetworkEndpointGroup#version}
	Version *string `field:"optional" json:"version" yaml:"version"`
}


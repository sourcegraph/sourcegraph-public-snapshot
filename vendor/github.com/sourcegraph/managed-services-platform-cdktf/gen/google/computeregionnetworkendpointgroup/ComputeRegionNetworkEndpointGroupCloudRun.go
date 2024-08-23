package computeregionnetworkendpointgroup


type ComputeRegionNetworkEndpointGroupCloudRun struct {
	// Cloud Run service is the main resource of Cloud Run.
	//
	// The service must be 1-63 characters long, and comply with RFC1035.
	// Example value: "run-service".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#service ComputeRegionNetworkEndpointGroup#service}
	Service *string `field:"optional" json:"service" yaml:"service"`
	// Cloud Run tag represents the "named-revision" to provide additional fine-grained traffic routing information.
	//
	// The tag must be 1-63 characters long, and comply with RFC1035.
	// Example value: "revision-0010".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#tag ComputeRegionNetworkEndpointGroup#tag}
	Tag *string `field:"optional" json:"tag" yaml:"tag"`
	// A template to parse service and tag fields from a request URL.
	//
	// URL mask allows for routing to multiple Run services without having
	// to create multiple network endpoint groups and backend services.
	//
	// For example, request URLs "foo1.domain.com/bar1" and "foo1.domain.com/bar2"
	// an be backed by the same Serverless Network Endpoint Group (NEG) with
	// URL mask ".domain.com/". The URL mask will parse them to { service="bar1", tag="foo1" }
	// and { service="bar2", tag="foo2" } respectively.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_region_network_endpoint_group#url_mask ComputeRegionNetworkEndpointGroup#url_mask}
	UrlMask *string `field:"optional" json:"urlMask" yaml:"urlMask"`
}


package computeurlmap


type ComputeUrlMapDefaultRouteActionWeightedBackendServicesHeaderActionRequestHeadersToAdd struct {
	// The name of the header to add.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#header_name ComputeUrlMap#header_name}
	HeaderName *string `field:"optional" json:"headerName" yaml:"headerName"`
	// The value of the header to add.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#header_value ComputeUrlMap#header_value}
	HeaderValue *string `field:"optional" json:"headerValue" yaml:"headerValue"`
	// If false, headerValue is appended to any values that already exist for the header.
	//
	// If true, headerValue is set for the header, discarding any values that were set for that header.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#replace ComputeUrlMap#replace}
	Replace interface{} `field:"optional" json:"replace" yaml:"replace"`
}


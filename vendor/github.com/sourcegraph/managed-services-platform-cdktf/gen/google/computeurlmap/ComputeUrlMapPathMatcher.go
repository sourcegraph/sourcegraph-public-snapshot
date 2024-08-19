package computeurlmap


type ComputeUrlMapPathMatcher struct {
	// The name to which this PathMatcher is referred by the HostRule.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#name ComputeUrlMap#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// default_route_action block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#default_route_action ComputeUrlMap#default_route_action}
	DefaultRouteAction *ComputeUrlMapPathMatcherDefaultRouteAction `field:"optional" json:"defaultRouteAction" yaml:"defaultRouteAction"`
	// The backend service or backend bucket to use when none of the given paths match.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#default_service ComputeUrlMap#default_service}
	DefaultService *string `field:"optional" json:"defaultService" yaml:"defaultService"`
	// default_url_redirect block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#default_url_redirect ComputeUrlMap#default_url_redirect}
	DefaultUrlRedirect *ComputeUrlMapPathMatcherDefaultUrlRedirect `field:"optional" json:"defaultUrlRedirect" yaml:"defaultUrlRedirect"`
	// An optional description of this resource. Provide this property when you create the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#description ComputeUrlMap#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// header_action block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#header_action ComputeUrlMap#header_action}
	HeaderAction *ComputeUrlMapPathMatcherHeaderAction `field:"optional" json:"headerAction" yaml:"headerAction"`
	// path_rule block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#path_rule ComputeUrlMap#path_rule}
	PathRule interface{} `field:"optional" json:"pathRule" yaml:"pathRule"`
	// route_rules block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#route_rules ComputeUrlMap#route_rules}
	RouteRules interface{} `field:"optional" json:"routeRules" yaml:"routeRules"`
}


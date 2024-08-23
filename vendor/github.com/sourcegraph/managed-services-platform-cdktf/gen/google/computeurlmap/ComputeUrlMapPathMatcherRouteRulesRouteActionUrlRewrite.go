package computeurlmap


type ComputeUrlMapPathMatcherRouteRulesRouteActionUrlRewrite struct {
	// Prior to forwarding the request to the selected service, the request's host header is replaced with contents of hostRewrite.
	//
	// The value must be between 1 and
	// 255 characters.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#host_rewrite ComputeUrlMap#host_rewrite}
	HostRewrite *string `field:"optional" json:"hostRewrite" yaml:"hostRewrite"`
	// Prior to forwarding the request to the selected backend service, the matching portion of the request's path is replaced by pathPrefixRewrite.
	//
	// The value must
	// be between 1 and 1024 characters.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#path_prefix_rewrite ComputeUrlMap#path_prefix_rewrite}
	PathPrefixRewrite *string `field:"optional" json:"pathPrefixRewrite" yaml:"pathPrefixRewrite"`
	// Prior to forwarding the request to the selected origin, if the request matched a pathTemplateMatch, the matching portion of the request's path is replaced re-written using the pattern specified by pathTemplateRewrite.
	//
	// pathTemplateRewrite must be between 1 and 255 characters
	// (inclusive), must start with a '/', and must only use variables
	// captured by the route's pathTemplate matchers.
	//
	// pathTemplateRewrite may only be used when all of a route's
	// MatchRules specify pathTemplate.
	//
	// Only one of pathPrefixRewrite and pathTemplateRewrite may be
	// specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#path_template_rewrite ComputeUrlMap#path_template_rewrite}
	PathTemplateRewrite *string `field:"optional" json:"pathTemplateRewrite" yaml:"pathTemplateRewrite"`
}


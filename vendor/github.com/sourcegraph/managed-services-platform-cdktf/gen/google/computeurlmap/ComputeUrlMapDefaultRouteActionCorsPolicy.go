package computeurlmap


type ComputeUrlMapDefaultRouteActionCorsPolicy struct {
	// In response to a preflight request, setting this to true indicates that the actual request can include user credentials.
	//
	// This translates to the Access-Control-Allow-Credentials header.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#allow_credentials ComputeUrlMap#allow_credentials}
	AllowCredentials interface{} `field:"optional" json:"allowCredentials" yaml:"allowCredentials"`
	// Specifies the content for the Access-Control-Allow-Headers header.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#allow_headers ComputeUrlMap#allow_headers}
	AllowHeaders *[]*string `field:"optional" json:"allowHeaders" yaml:"allowHeaders"`
	// Specifies the content for the Access-Control-Allow-Methods header.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#allow_methods ComputeUrlMap#allow_methods}
	AllowMethods *[]*string `field:"optional" json:"allowMethods" yaml:"allowMethods"`
	// Specifies the regular expression patterns that match allowed origins.
	//
	// For regular expression grammar
	// please see en.cppreference.com/w/cpp/regex/ecmascript
	// An origin is allowed if it matches either an item in allowOrigins or an item in allowOriginRegexes.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#allow_origin_regexes ComputeUrlMap#allow_origin_regexes}
	AllowOriginRegexes *[]*string `field:"optional" json:"allowOriginRegexes" yaml:"allowOriginRegexes"`
	// Specifies the list of origins that will be allowed to do CORS requests.
	//
	// An origin is allowed if it matches either an item in allowOrigins or an item in allowOriginRegexes.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#allow_origins ComputeUrlMap#allow_origins}
	AllowOrigins *[]*string `field:"optional" json:"allowOrigins" yaml:"allowOrigins"`
	// If true, specifies the CORS policy is disabled.
	//
	// The default value is false, which indicates that the CORS policy is in effect.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#disabled ComputeUrlMap#disabled}
	Disabled interface{} `field:"optional" json:"disabled" yaml:"disabled"`
	// Specifies the content for the Access-Control-Expose-Headers header.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#expose_headers ComputeUrlMap#expose_headers}
	ExposeHeaders *[]*string `field:"optional" json:"exposeHeaders" yaml:"exposeHeaders"`
	// Specifies how long results of a preflight request can be cached in seconds. This translates to the Access-Control-Max-Age header.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#max_age ComputeUrlMap#max_age}
	MaxAge *float64 `field:"optional" json:"maxAge" yaml:"maxAge"`
}


package computeurlmap


type ComputeUrlMapPathMatcherRouteRulesMatchRulesHeaderMatches struct {
	// The name of the HTTP header to match.
	//
	// For matching against the HTTP request's
	// authority, use a headerMatch with the header name ":authority". For matching a
	// request's method, use the headerName ":method".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#header_name ComputeUrlMap#header_name}
	HeaderName *string `field:"required" json:"headerName" yaml:"headerName"`
	// The value should exactly match contents of exactMatch.
	//
	// Only one of exactMatch,
	// prefixMatch, suffixMatch, regexMatch, presentMatch or rangeMatch must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#exact_match ComputeUrlMap#exact_match}
	ExactMatch *string `field:"optional" json:"exactMatch" yaml:"exactMatch"`
	// If set to false, the headerMatch is considered a match if the match criteria above are met.
	//
	// If set to true, the headerMatch is considered a match if the
	// match criteria above are NOT met. Defaults to false.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#invert_match ComputeUrlMap#invert_match}
	InvertMatch interface{} `field:"optional" json:"invertMatch" yaml:"invertMatch"`
	// The value of the header must start with the contents of prefixMatch.
	//
	// Only one of
	// exactMatch, prefixMatch, suffixMatch, regexMatch, presentMatch or rangeMatch
	// must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#prefix_match ComputeUrlMap#prefix_match}
	PrefixMatch *string `field:"optional" json:"prefixMatch" yaml:"prefixMatch"`
	// A header with the contents of headerName must exist.
	//
	// The match takes place
	// whether or not the request's header has a value or not. Only one of exactMatch,
	// prefixMatch, suffixMatch, regexMatch, presentMatch or rangeMatch must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#present_match ComputeUrlMap#present_match}
	PresentMatch interface{} `field:"optional" json:"presentMatch" yaml:"presentMatch"`
	// range_match block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#range_match ComputeUrlMap#range_match}
	RangeMatch *ComputeUrlMapPathMatcherRouteRulesMatchRulesHeaderMatchesRangeMatch `field:"optional" json:"rangeMatch" yaml:"rangeMatch"`
	// The value of the header must match the regular expression specified in regexMatch.
	//
	// For regular expression grammar, please see:
	// en.cppreference.com/w/cpp/regex/ecmascript  For matching against a port
	// specified in the HTTP request, use a headerMatch with headerName set to PORT and
	// a regular expression that satisfies the RFC2616 Host header's port specifier.
	// Only one of exactMatch, prefixMatch, suffixMatch, regexMatch, presentMatch or
	// rangeMatch must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#regex_match ComputeUrlMap#regex_match}
	RegexMatch *string `field:"optional" json:"regexMatch" yaml:"regexMatch"`
	// The value of the header must end with the contents of suffixMatch.
	//
	// Only one of
	// exactMatch, prefixMatch, suffixMatch, regexMatch, presentMatch or rangeMatch
	// must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#suffix_match ComputeUrlMap#suffix_match}
	SuffixMatch *string `field:"optional" json:"suffixMatch" yaml:"suffixMatch"`
}


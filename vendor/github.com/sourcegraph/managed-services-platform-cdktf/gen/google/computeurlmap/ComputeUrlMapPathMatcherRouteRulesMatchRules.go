package computeurlmap


type ComputeUrlMapPathMatcherRouteRulesMatchRules struct {
	// For satisfying the matchRule condition, the path of the request must exactly match the value specified in fullPathMatch after removing any query parameters and anchor that may be part of the original URL.
	//
	// FullPathMatch must be between 1
	// and 1024 characters. Only one of prefixMatch, fullPathMatch or regexMatch must
	// be specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#full_path_match ComputeUrlMap#full_path_match}
	FullPathMatch *string `field:"optional" json:"fullPathMatch" yaml:"fullPathMatch"`
	// header_matches block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#header_matches ComputeUrlMap#header_matches}
	HeaderMatches interface{} `field:"optional" json:"headerMatches" yaml:"headerMatches"`
	// Specifies that prefixMatch and fullPathMatch matches are case sensitive. Defaults to false.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#ignore_case ComputeUrlMap#ignore_case}
	IgnoreCase interface{} `field:"optional" json:"ignoreCase" yaml:"ignoreCase"`
	// metadata_filters block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#metadata_filters ComputeUrlMap#metadata_filters}
	MetadataFilters interface{} `field:"optional" json:"metadataFilters" yaml:"metadataFilters"`
	// For satisfying the matchRule condition, the path of the request must match the wildcard pattern specified in pathTemplateMatch after removing any query parameters and anchor that may be part of the original URL.
	//
	// pathTemplateMatch must be between 1 and 255 characters
	// (inclusive).  The pattern specified by pathTemplateMatch may
	// have at most 5 wildcard operators and at most 5 variable
	// captures in total.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#path_template_match ComputeUrlMap#path_template_match}
	PathTemplateMatch *string `field:"optional" json:"pathTemplateMatch" yaml:"pathTemplateMatch"`
	// For satisfying the matchRule condition, the request's path must begin with the specified prefixMatch.
	//
	// prefixMatch must begin with a /. The value must be
	// between 1 and 1024 characters. Only one of prefixMatch, fullPathMatch or
	// regexMatch must be specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#prefix_match ComputeUrlMap#prefix_match}
	PrefixMatch *string `field:"optional" json:"prefixMatch" yaml:"prefixMatch"`
	// query_parameter_matches block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#query_parameter_matches ComputeUrlMap#query_parameter_matches}
	QueryParameterMatches interface{} `field:"optional" json:"queryParameterMatches" yaml:"queryParameterMatches"`
	// For satisfying the matchRule condition, the path of the request must satisfy the regular expression specified in regexMatch after removing any query parameters and anchor supplied with the original URL.
	//
	// For regular expression grammar please
	// see en.cppreference.com/w/cpp/regex/ecmascript  Only one of prefixMatch,
	// fullPathMatch or regexMatch must be specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#regex_match ComputeUrlMap#regex_match}
	RegexMatch *string `field:"optional" json:"regexMatch" yaml:"regexMatch"`
}


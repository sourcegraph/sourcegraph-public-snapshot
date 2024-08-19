package computeurlmap


type ComputeUrlMapPathMatcherRouteRulesMatchRulesHeaderMatchesRangeMatch struct {
	// The end of the range (exclusive).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#range_end ComputeUrlMap#range_end}
	RangeEnd *float64 `field:"required" json:"rangeEnd" yaml:"rangeEnd"`
	// The start of the range (inclusive).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#range_start ComputeUrlMap#range_start}
	RangeStart *float64 `field:"required" json:"rangeStart" yaml:"rangeStart"`
}


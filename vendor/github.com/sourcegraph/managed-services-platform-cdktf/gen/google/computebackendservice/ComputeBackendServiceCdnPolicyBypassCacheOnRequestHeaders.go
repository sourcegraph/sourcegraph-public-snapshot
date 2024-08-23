package computebackendservice


type ComputeBackendServiceCdnPolicyBypassCacheOnRequestHeaders struct {
	// The header field name to match on when bypassing cache. Values are case-insensitive.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#header_name ComputeBackendService#header_name}
	HeaderName *string `field:"required" json:"headerName" yaml:"headerName"`
}


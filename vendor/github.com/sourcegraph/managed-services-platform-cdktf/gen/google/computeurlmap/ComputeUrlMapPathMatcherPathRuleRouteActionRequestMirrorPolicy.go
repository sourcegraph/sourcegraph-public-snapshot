package computeurlmap


type ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicy struct {
	// The BackendService resource being mirrored to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#backend_service ComputeUrlMap#backend_service}
	BackendService *string `field:"required" json:"backendService" yaml:"backendService"`
}


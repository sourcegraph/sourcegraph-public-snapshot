package computebackendservice


type ComputeBackendServiceCdnPolicyNegativeCachingPolicy struct {
	// The HTTP status code to define a TTL against.
	//
	// Only HTTP status codes 300, 301, 308, 404, 405, 410, 421, 451 and 501
	// can be specified as values, and you cannot specify a status code more than once.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#code ComputeBackendService#code}
	Code *float64 `field:"optional" json:"code" yaml:"code"`
	// The TTL (in seconds) for which to cache responses with the corresponding status code.
	//
	// The maximum allowed value is 1800s
	// (30 minutes), noting that infrequently accessed objects may be evicted from the cache before the defined TTL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#ttl ComputeBackendService#ttl}
	Ttl *float64 `field:"optional" json:"ttl" yaml:"ttl"`
}


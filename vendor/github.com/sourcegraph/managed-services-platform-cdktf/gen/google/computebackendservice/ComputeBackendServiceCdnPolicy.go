package computebackendservice


type ComputeBackendServiceCdnPolicy struct {
	// bypass_cache_on_request_headers block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#bypass_cache_on_request_headers ComputeBackendService#bypass_cache_on_request_headers}
	BypassCacheOnRequestHeaders interface{} `field:"optional" json:"bypassCacheOnRequestHeaders" yaml:"bypassCacheOnRequestHeaders"`
	// cache_key_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#cache_key_policy ComputeBackendService#cache_key_policy}
	CacheKeyPolicy *ComputeBackendServiceCdnPolicyCacheKeyPolicy `field:"optional" json:"cacheKeyPolicy" yaml:"cacheKeyPolicy"`
	// Specifies the cache setting for all responses from this backend.
	//
	// The possible values are: USE_ORIGIN_HEADERS, FORCE_CACHE_ALL and CACHE_ALL_STATIC Possible values: ["USE_ORIGIN_HEADERS", "FORCE_CACHE_ALL", "CACHE_ALL_STATIC"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#cache_mode ComputeBackendService#cache_mode}
	CacheMode *string `field:"optional" json:"cacheMode" yaml:"cacheMode"`
	// Specifies the maximum allowed TTL for cached content served by this origin.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#client_ttl ComputeBackendService#client_ttl}
	ClientTtl *float64 `field:"optional" json:"clientTtl" yaml:"clientTtl"`
	// Specifies the default TTL for cached content served by this origin for responses that do not have an existing valid TTL (max-age or s-max-age).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#default_ttl ComputeBackendService#default_ttl}
	DefaultTtl *float64 `field:"optional" json:"defaultTtl" yaml:"defaultTtl"`
	// Specifies the maximum allowed TTL for cached content served by this origin.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_ttl ComputeBackendService#max_ttl}
	MaxTtl *float64 `field:"optional" json:"maxTtl" yaml:"maxTtl"`
	// Negative caching allows per-status code TTLs to be set, in order to apply fine-grained caching for common errors or redirects.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#negative_caching ComputeBackendService#negative_caching}
	NegativeCaching interface{} `field:"optional" json:"negativeCaching" yaml:"negativeCaching"`
	// negative_caching_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#negative_caching_policy ComputeBackendService#negative_caching_policy}
	NegativeCachingPolicy interface{} `field:"optional" json:"negativeCachingPolicy" yaml:"negativeCachingPolicy"`
	// Serve existing content from the cache (if available) when revalidating content with the origin, or when an error is encountered when refreshing the cache.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#serve_while_stale ComputeBackendService#serve_while_stale}
	ServeWhileStale *float64 `field:"optional" json:"serveWhileStale" yaml:"serveWhileStale"`
	// Maximum number of seconds the response to a signed URL request will be considered fresh, defaults to 1hr (3600s).
	//
	// After this
	// time period, the response will be revalidated before
	// being served.
	//
	// When serving responses to signed URL requests, Cloud CDN will
	// internally behave as though all responses from this backend had a
	// "Cache-Control: public, max-age=[TTL]" header, regardless of any
	// existing Cache-Control header. The actual headers served in
	// responses will not be altered.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#signed_url_cache_max_age_sec ComputeBackendService#signed_url_cache_max_age_sec}
	SignedUrlCacheMaxAgeSec *float64 `field:"optional" json:"signedUrlCacheMaxAgeSec" yaml:"signedUrlCacheMaxAgeSec"`
}


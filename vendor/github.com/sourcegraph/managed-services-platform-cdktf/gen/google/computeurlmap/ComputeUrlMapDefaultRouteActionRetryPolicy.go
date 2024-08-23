package computeurlmap


type ComputeUrlMapDefaultRouteActionRetryPolicy struct {
	// Specifies the allowed number retries. This number must be > 0. If not specified, defaults to 1.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#num_retries ComputeUrlMap#num_retries}
	NumRetries *float64 `field:"optional" json:"numRetries" yaml:"numRetries"`
	// per_try_timeout block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#per_try_timeout ComputeUrlMap#per_try_timeout}
	PerTryTimeout *ComputeUrlMapDefaultRouteActionRetryPolicyPerTryTimeout `field:"optional" json:"perTryTimeout" yaml:"perTryTimeout"`
	// Specfies one or more conditions when this retry rule applies. Valid values are:.
	//
	// 5xx: Loadbalancer will attempt a retry if the backend service responds with any 5xx response code,
	// or if the backend service does not respond at all, example: disconnects, reset, read timeout,
	// connection failure, and refused streams.
	// gateway-error: Similar to 5xx, but only applies to response codes 502, 503 or 504.
	// connect-failure: Loadbalancer will retry on failures connecting to backend services,
	// for example due to connection timeouts.
	// retriable-4xx: Loadbalancer will retry for retriable 4xx response codes.
	// Currently the only retriable error supported is 409.
	// refused-stream:Loadbalancer will retry if the backend service resets the stream with a REFUSED_STREAM error code.
	// This reset type indicates that it is safe to retry.
	// cancelled: Loadbalancer will retry if the gRPC status code in the response header is set to cancelled
	// deadline-exceeded: Loadbalancer will retry if the gRPC status code in the response header is set to deadline-exceeded
	// resource-exhausted: Loadbalancer will retry if the gRPC status code in the response header is set to resource-exhausted
	// unavailable: Loadbalancer will retry if the gRPC status code in the response header is set to unavailable
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#retry_conditions ComputeUrlMap#retry_conditions}
	RetryConditions *[]*string `field:"optional" json:"retryConditions" yaml:"retryConditions"`
}


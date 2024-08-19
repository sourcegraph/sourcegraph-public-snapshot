package storagebucket


type StorageBucketCors struct {
	// The value, in seconds, to return in the Access-Control-Max-Age header used in preflight responses.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#max_age_seconds StorageBucket#max_age_seconds}
	MaxAgeSeconds *float64 `field:"optional" json:"maxAgeSeconds" yaml:"maxAgeSeconds"`
	// The list of HTTP methods on which to include CORS response headers, (GET, OPTIONS, POST, etc) Note: "*" is permitted in the list of methods, and means "any method".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#method StorageBucket#method}
	Method *[]*string `field:"optional" json:"method" yaml:"method"`
	// The list of Origins eligible to receive CORS response headers.
	//
	// Note: "*" is permitted in the list of origins, and means "any Origin".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#origin StorageBucket#origin}
	Origin *[]*string `field:"optional" json:"origin" yaml:"origin"`
	// The list of HTTP headers other than the simple response headers to give permission for the user-agent to share across domains.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#response_header StorageBucket#response_header}
	ResponseHeader *[]*string `field:"optional" json:"responseHeader" yaml:"responseHeader"`
}


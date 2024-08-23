package computebackendservice


type ComputeBackendServiceLocalityLbPoliciesCustomPolicy struct {
	// Identifies the custom policy.
	//
	// The value should match the type the custom implementation is registered
	// with on the gRPC clients. It should follow protocol buffer
	// message naming conventions and include the full path (e.g.
	// myorg.CustomLbPolicy). The maximum length is 256 characters.
	//
	// Note that specifying the same custom policy more than once for a
	// backend is not a valid configuration and will be rejected.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#name ComputeBackendService#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// An optional, arbitrary JSON object with configuration data, understood by a locally installed custom policy implementation.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#data ComputeBackendService#data}
	Data *string `field:"optional" json:"data" yaml:"data"`
}


package computebackendservice


type ComputeBackendServiceLocalityLbPoliciesPolicy struct {
	// The name of a locality load balancer policy to be used.
	//
	// The value
	// should be one of the predefined ones as supported by localityLbPolicy,
	// although at the moment only ROUND_ROBIN is supported.
	//
	// This field should only be populated when the customPolicy field is not
	// used.
	//
	// Note that specifying the same policy more than once for a backend is
	// not a valid configuration and will be rejected.
	//
	// The possible values are:
	//
	// 'ROUND_ROBIN': This is a simple policy in which each healthy backend
	//             is selected in round robin order.
	//
	// 'LEAST_REQUEST': An O(1) algorithm which selects two random healthy
	//               hosts and picks the host which has fewer active requests.
	//
	// 'RING_HASH': The ring/modulo hash load balancer implements consistent
	//           hashing to backends. The algorithm has the property that the
	//           addition/removal of a host from a set of N hosts only affects
	//           1/N of the requests.
	//
	// 'RANDOM': The load balancer selects a random healthy host.
	//
	// 'ORIGINAL_DESTINATION': Backend host is selected based on the client
	//                       connection metadata, i.e., connections are opened
	//                       to the same address as the destination address of
	//                       the incoming connection before the connection
	//                       was redirected to the load balancer.
	//
	// 'MAGLEV': used as a drop in replacement for the ring hash load balancer.
	//         Maglev is not as stable as ring hash but has faster table lookup
	//         build times and host selection times. For more information about
	//         Maglev, refer to https://ai.google/research/pubs/pub44824 Possible values: ["ROUND_ROBIN", "LEAST_REQUEST", "RING_HASH", "RANDOM", "ORIGINAL_DESTINATION", "MAGLEV"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#name ComputeBackendService#name}
	Name *string `field:"required" json:"name" yaml:"name"`
}


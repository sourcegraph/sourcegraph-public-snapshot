package computebackendservice


type ComputeBackendServiceBackend struct {
	// The fully-qualified URL of an Instance Group or Network Endpoint Group resource.
	//
	// In case of instance group this defines the list
	// of instances that serve traffic. Member virtual machine
	// instances from each instance group must live in the same zone as
	// the instance group itself. No two backends in a backend service
	// are allowed to use same Instance Group resource.
	//
	// For Network Endpoint Groups this defines list of endpoints. All
	// endpoints of Network Endpoint Group must be hosted on instances
	// located in the same zone as the Network Endpoint Group.
	//
	// Backend services cannot mix Instance Group and
	// Network Endpoint Group backends.
	//
	// Note that you must specify an Instance Group or Network Endpoint
	// Group resource using the fully-qualified URL, rather than a
	// partial URL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#group ComputeBackendService#group}
	Group *string `field:"required" json:"group" yaml:"group"`
	// Specifies the balancing mode for this backend.
	//
	// For global HTTP(S) or TCP/SSL load balancing, the default is
	// UTILIZATION. Valid values are UTILIZATION, RATE (for HTTP(S))
	// and CONNECTION (for TCP/SSL).
	//
	// See the [Backend Services Overview](https://cloud.google.com/load-balancing/docs/backend-service#balancing-mode)
	// for an explanation of load balancing modes. Default value: "UTILIZATION" Possible values: ["UTILIZATION", "RATE", "CONNECTION"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#balancing_mode ComputeBackendService#balancing_mode}
	BalancingMode *string `field:"optional" json:"balancingMode" yaml:"balancingMode"`
	// A multiplier applied to the group's maximum servicing capacity (based on UTILIZATION, RATE or CONNECTION).
	//
	// Default value is 1, which means the group will serve up to 100%
	// of its configured capacity (depending on balancingMode). A
	// setting of 0 means the group is completely drained, offering
	// 0% of its available Capacity. Valid range is [0.0,1.0].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#capacity_scaler ComputeBackendService#capacity_scaler}
	CapacityScaler *float64 `field:"optional" json:"capacityScaler" yaml:"capacityScaler"`
	// An optional description of this resource. Provide this property when you create the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#description ComputeBackendService#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// The max number of simultaneous connections for the group. Can be used with either CONNECTION or UTILIZATION balancing modes.
	//
	// For CONNECTION mode, either maxConnections or one
	// of maxConnectionsPerInstance or maxConnectionsPerEndpoint,
	// as appropriate for group type, must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_connections ComputeBackendService#max_connections}
	MaxConnections *float64 `field:"optional" json:"maxConnections" yaml:"maxConnections"`
	// The max number of simultaneous connections that a single backend network endpoint can handle.
	//
	// This is used to calculate the
	// capacity of the group. Can be used in either CONNECTION or
	// UTILIZATION balancing modes.
	//
	// For CONNECTION mode, either
	// maxConnections or maxConnectionsPerEndpoint must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_connections_per_endpoint ComputeBackendService#max_connections_per_endpoint}
	MaxConnectionsPerEndpoint *float64 `field:"optional" json:"maxConnectionsPerEndpoint" yaml:"maxConnectionsPerEndpoint"`
	// The max number of simultaneous connections that a single backend instance can handle.
	//
	// This is used to calculate the
	// capacity of the group. Can be used in either CONNECTION or
	// UTILIZATION balancing modes.
	//
	// For CONNECTION mode, either maxConnections or
	// maxConnectionsPerInstance must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_connections_per_instance ComputeBackendService#max_connections_per_instance}
	MaxConnectionsPerInstance *float64 `field:"optional" json:"maxConnectionsPerInstance" yaml:"maxConnectionsPerInstance"`
	// The max requests per second (RPS) of the group.
	//
	// Can be used with either RATE or UTILIZATION balancing modes,
	// but required if RATE mode. For RATE mode, either maxRate or one
	// of maxRatePerInstance or maxRatePerEndpoint, as appropriate for
	// group type, must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_rate ComputeBackendService#max_rate}
	MaxRate *float64 `field:"optional" json:"maxRate" yaml:"maxRate"`
	// The max requests per second (RPS) that a single backend network endpoint can handle.
	//
	// This is used to calculate the capacity of
	// the group. Can be used in either balancing mode. For RATE mode,
	// either maxRate or maxRatePerEndpoint must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_rate_per_endpoint ComputeBackendService#max_rate_per_endpoint}
	MaxRatePerEndpoint *float64 `field:"optional" json:"maxRatePerEndpoint" yaml:"maxRatePerEndpoint"`
	// The max requests per second (RPS) that a single backend instance can handle.
	//
	// This is used to calculate the capacity of
	// the group. Can be used in either balancing mode. For RATE mode,
	// either maxRate or maxRatePerInstance must be set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_rate_per_instance ComputeBackendService#max_rate_per_instance}
	MaxRatePerInstance *float64 `field:"optional" json:"maxRatePerInstance" yaml:"maxRatePerInstance"`
	// Used when balancingMode is UTILIZATION. This ratio defines the CPU utilization target for the group. Valid range is [0.0, 1.0].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_utilization ComputeBackendService#max_utilization}
	MaxUtilization *float64 `field:"optional" json:"maxUtilization" yaml:"maxUtilization"`
}


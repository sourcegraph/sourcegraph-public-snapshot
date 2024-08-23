package cloudrunv2service


type CloudRunV2ServiceTemplateContainersResources struct {
	// Determines whether CPU is only allocated during requests.
	//
	// True by default if the parent 'resources' field is not set. However, if
	// 'resources' is set, this field must be explicitly set to true to preserve the default behavior.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#cpu_idle CloudRunV2Service#cpu_idle}
	CpuIdle interface{} `field:"optional" json:"cpuIdle" yaml:"cpuIdle"`
	// Only memory and CPU are supported.
	//
	// Use key 'cpu' for CPU limit and 'memory' for memory limit. Note: The only supported values for CPU are '1', '2', '4', and '8'. Setting 4 CPU requires at least 2Gi of memory. The values of the map is string form of the 'quantity' k8s type: https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/api/resource/quantity.go
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#limits CloudRunV2Service#limits}
	Limits *map[string]*string `field:"optional" json:"limits" yaml:"limits"`
	// Determines whether CPU should be boosted on startup of a new container instance above the requested CPU threshold, this can help reduce cold-start latency.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#startup_cpu_boost CloudRunV2Service#startup_cpu_boost}
	StartupCpuBoost interface{} `field:"optional" json:"startupCpuBoost" yaml:"startupCpuBoost"`
}


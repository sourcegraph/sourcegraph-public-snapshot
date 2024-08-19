package computebackendservice


type ComputeBackendServiceOutlierDetection struct {
	// base_ejection_time block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#base_ejection_time ComputeBackendService#base_ejection_time}
	BaseEjectionTime *ComputeBackendServiceOutlierDetectionBaseEjectionTime `field:"optional" json:"baseEjectionTime" yaml:"baseEjectionTime"`
	// Number of errors before a host is ejected from the connection pool.
	//
	// When the
	// backend host is accessed over HTTP, a 5xx return code qualifies as an error.
	// Defaults to 5.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#consecutive_errors ComputeBackendService#consecutive_errors}
	ConsecutiveErrors *float64 `field:"optional" json:"consecutiveErrors" yaml:"consecutiveErrors"`
	// The number of consecutive gateway failures (502, 503, 504 status or connection errors that are mapped to one of those status codes) before a consecutive gateway failure ejection occurs.
	//
	// Defaults to 5.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#consecutive_gateway_failure ComputeBackendService#consecutive_gateway_failure}
	ConsecutiveGatewayFailure *float64 `field:"optional" json:"consecutiveGatewayFailure" yaml:"consecutiveGatewayFailure"`
	// The percentage chance that a host will be actually ejected when an outlier status is detected through consecutive 5xx.
	//
	// This setting can be used to disable
	// ejection or to ramp it up slowly. Defaults to 100.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#enforcing_consecutive_errors ComputeBackendService#enforcing_consecutive_errors}
	EnforcingConsecutiveErrors *float64 `field:"optional" json:"enforcingConsecutiveErrors" yaml:"enforcingConsecutiveErrors"`
	// The percentage chance that a host will be actually ejected when an outlier status is detected through consecutive gateway failures.
	//
	// This setting can be
	// used to disable ejection or to ramp it up slowly. Defaults to 0.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#enforcing_consecutive_gateway_failure ComputeBackendService#enforcing_consecutive_gateway_failure}
	EnforcingConsecutiveGatewayFailure *float64 `field:"optional" json:"enforcingConsecutiveGatewayFailure" yaml:"enforcingConsecutiveGatewayFailure"`
	// The percentage chance that a host will be actually ejected when an outlier status is detected through success rate statistics.
	//
	// This setting can be used to
	// disable ejection or to ramp it up slowly. Defaults to 100.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#enforcing_success_rate ComputeBackendService#enforcing_success_rate}
	EnforcingSuccessRate *float64 `field:"optional" json:"enforcingSuccessRate" yaml:"enforcingSuccessRate"`
	// interval block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#interval ComputeBackendService#interval}
	Interval *ComputeBackendServiceOutlierDetectionInterval `field:"optional" json:"interval" yaml:"interval"`
	// Maximum percentage of hosts in the load balancing pool for the backend service that can be ejected. Defaults to 10%.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#max_ejection_percent ComputeBackendService#max_ejection_percent}
	MaxEjectionPercent *float64 `field:"optional" json:"maxEjectionPercent" yaml:"maxEjectionPercent"`
	// The number of hosts in a cluster that must have enough request volume to detect success rate outliers.
	//
	// If the number of hosts is less than this setting, outlier
	// detection via success rate statistics is not performed for any host in the
	// cluster. Defaults to 5.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#success_rate_minimum_hosts ComputeBackendService#success_rate_minimum_hosts}
	SuccessRateMinimumHosts *float64 `field:"optional" json:"successRateMinimumHosts" yaml:"successRateMinimumHosts"`
	// The minimum number of total requests that must be collected in one interval (as defined by the interval duration above) to include this host in success rate based outlier detection.
	//
	// If the volume is lower than this setting, outlier
	// detection via success rate statistics is not performed for that host. Defaults
	// to 100.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#success_rate_request_volume ComputeBackendService#success_rate_request_volume}
	SuccessRateRequestVolume *float64 `field:"optional" json:"successRateRequestVolume" yaml:"successRateRequestVolume"`
	// This factor is used to determine the ejection threshold for success rate outlier ejection.
	//
	// The ejection threshold is the difference between the mean success
	// rate, and the product of this factor and the standard deviation of the mean
	// success rate: mean - (stdev * success_rate_stdev_factor). This factor is divided
	// by a thousand to get a double. That is, if the desired factor is 1.9, the
	// runtime value should be 1900. Defaults to 1900.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#success_rate_stdev_factor ComputeBackendService#success_rate_stdev_factor}
	SuccessRateStdevFactor *float64 `field:"optional" json:"successRateStdevFactor" yaml:"successRateStdevFactor"`
}


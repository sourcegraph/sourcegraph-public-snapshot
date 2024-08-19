package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigHttpCheckAuthInfo struct {
	// The password to authenticate.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#password MonitoringUptimeCheckConfig#password}
	Password *string `field:"required" json:"password" yaml:"password"`
	// The username to authenticate.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#username MonitoringUptimeCheckConfig#username}
	Username *string `field:"required" json:"username" yaml:"username"`
}


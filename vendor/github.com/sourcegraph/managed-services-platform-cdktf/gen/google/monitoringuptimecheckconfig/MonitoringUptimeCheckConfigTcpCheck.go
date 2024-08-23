package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigTcpCheck struct {
	// The port to the page to run the check against.
	//
	// Will be combined with host (specified within the 'monitored_resource') to construct the full URL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#port MonitoringUptimeCheckConfig#port}
	Port *float64 `field:"required" json:"port" yaml:"port"`
	// ping_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#ping_config MonitoringUptimeCheckConfig#ping_config}
	PingConfig *MonitoringUptimeCheckConfigTcpCheckPingConfig `field:"optional" json:"pingConfig" yaml:"pingConfig"`
}


package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigTcpCheckPingConfig struct {
	// Number of ICMP pings. A maximum of 3 ICMP pings is currently supported.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#pings_count MonitoringUptimeCheckConfig#pings_count}
	PingsCount *float64 `field:"required" json:"pingsCount" yaml:"pingsCount"`
}


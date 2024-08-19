package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigHttpCheckAcceptedResponseStatusCodes struct {
	// A class of status codes to accept. Possible values: ["STATUS_CLASS_1XX", "STATUS_CLASS_2XX", "STATUS_CLASS_3XX", "STATUS_CLASS_4XX", "STATUS_CLASS_5XX", "STATUS_CLASS_ANY"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#status_class MonitoringUptimeCheckConfig#status_class}
	StatusClass *string `field:"optional" json:"statusClass" yaml:"statusClass"`
	// A status code to accept.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#status_value MonitoringUptimeCheckConfig#status_value}
	StatusValue *float64 `field:"optional" json:"statusValue" yaml:"statusValue"`
}


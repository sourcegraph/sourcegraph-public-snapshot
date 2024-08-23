package monitoringalertpolicy


type MonitoringAlertPolicyAlertStrategyNotificationRateLimit struct {
	// Not more than one notification per period.
	//
	// A duration in seconds with up to nine fractional digits, terminated by 's'. Example "60.5s".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#period MonitoringAlertPolicy#period}
	Period *string `field:"optional" json:"period" yaml:"period"`
}


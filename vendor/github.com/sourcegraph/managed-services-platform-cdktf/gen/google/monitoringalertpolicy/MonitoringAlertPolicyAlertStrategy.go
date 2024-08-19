package monitoringalertpolicy


type MonitoringAlertPolicyAlertStrategy struct {
	// If an alert policy that was active has no data for this long, any open incidents will close.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#auto_close MonitoringAlertPolicy#auto_close}
	AutoClose *string `field:"optional" json:"autoClose" yaml:"autoClose"`
	// notification_channel_strategy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#notification_channel_strategy MonitoringAlertPolicy#notification_channel_strategy}
	NotificationChannelStrategy interface{} `field:"optional" json:"notificationChannelStrategy" yaml:"notificationChannelStrategy"`
	// notification_rate_limit block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#notification_rate_limit MonitoringAlertPolicy#notification_rate_limit}
	NotificationRateLimit *MonitoringAlertPolicyAlertStrategyNotificationRateLimit `field:"optional" json:"notificationRateLimit" yaml:"notificationRateLimit"`
}


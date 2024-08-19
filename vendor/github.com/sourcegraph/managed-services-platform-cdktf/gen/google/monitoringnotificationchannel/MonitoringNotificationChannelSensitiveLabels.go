package monitoringnotificationchannel


type MonitoringNotificationChannelSensitiveLabels struct {
	// An authorization token for a notification channel. Channel types that support this field include: slack.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_notification_channel#auth_token MonitoringNotificationChannel#auth_token}
	AuthToken *string `field:"optional" json:"authToken" yaml:"authToken"`
	// An password for a notification channel. Channel types that support this field include: webhook_basicauth.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_notification_channel#password MonitoringNotificationChannel#password}
	Password *string `field:"optional" json:"password" yaml:"password"`
	// An servicekey token for a notification channel. Channel types that support this field include: pagerduty.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_notification_channel#service_key MonitoringNotificationChannel#service_key}
	ServiceKey *string `field:"optional" json:"serviceKey" yaml:"serviceKey"`
}


package monitoringalertpolicy


type MonitoringAlertPolicyAlertStrategyNotificationChannelStrategy struct {
	// The notification channels that these settings apply to.
	//
	// Each of these
	// correspond to the name field in one of the NotificationChannel objects
	// referenced in the notification_channels field of this AlertPolicy. The format is
	// 'projects/[PROJECT_ID_OR_NUMBER]/notificationChannels/[CHANNEL_ID]'
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#notification_channel_names MonitoringAlertPolicy#notification_channel_names}
	NotificationChannelNames *[]*string `field:"optional" json:"notificationChannelNames" yaml:"notificationChannelNames"`
	// The frequency at which to send reminder notifications for open incidents.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#renotify_interval MonitoringAlertPolicy#renotify_interval}
	RenotifyInterval *string `field:"optional" json:"renotifyInterval" yaml:"renotifyInterval"`
}


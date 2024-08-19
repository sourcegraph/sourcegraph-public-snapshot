package monitoringalertpolicy


type MonitoringAlertPolicyTimeouts struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#create MonitoringAlertPolicy#create}.
	Create *string `field:"optional" json:"create" yaml:"create"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#delete MonitoringAlertPolicy#delete}.
	Delete *string `field:"optional" json:"delete" yaml:"delete"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#update MonitoringAlertPolicy#update}.
	Update *string `field:"optional" json:"update" yaml:"update"`
}


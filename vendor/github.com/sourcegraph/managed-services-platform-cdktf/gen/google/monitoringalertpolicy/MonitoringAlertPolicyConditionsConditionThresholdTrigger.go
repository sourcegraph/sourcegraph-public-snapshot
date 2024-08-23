package monitoringalertpolicy


type MonitoringAlertPolicyConditionsConditionThresholdTrigger struct {
	// The absolute number of time series that must fail the predicate for the condition to be triggered.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#count MonitoringAlertPolicy#count}
	Count *float64 `field:"optional" json:"count" yaml:"count"`
	// The percentage of time series that must fail the predicate for the condition to be triggered.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#percent MonitoringAlertPolicy#percent}
	Percent *float64 `field:"optional" json:"percent" yaml:"percent"`
}


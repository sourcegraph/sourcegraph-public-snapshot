package monitoringalertpolicy


type MonitoringAlertPolicyConditions struct {
	// A short name or phrase used to identify the condition in dashboards, notifications, and incidents.
	//
	// To avoid confusion, don't use the same
	// display name for multiple conditions in the same
	// policy.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#display_name MonitoringAlertPolicy#display_name}
	DisplayName *string `field:"required" json:"displayName" yaml:"displayName"`
	// condition_absent block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#condition_absent MonitoringAlertPolicy#condition_absent}
	ConditionAbsent *MonitoringAlertPolicyConditionsConditionAbsent `field:"optional" json:"conditionAbsent" yaml:"conditionAbsent"`
	// condition_matched_log block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#condition_matched_log MonitoringAlertPolicy#condition_matched_log}
	ConditionMatchedLog *MonitoringAlertPolicyConditionsConditionMatchedLog `field:"optional" json:"conditionMatchedLog" yaml:"conditionMatchedLog"`
	// condition_monitoring_query_language block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#condition_monitoring_query_language MonitoringAlertPolicy#condition_monitoring_query_language}
	ConditionMonitoringQueryLanguage *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage `field:"optional" json:"conditionMonitoringQueryLanguage" yaml:"conditionMonitoringQueryLanguage"`
	// condition_prometheus_query_language block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#condition_prometheus_query_language MonitoringAlertPolicy#condition_prometheus_query_language}
	ConditionPrometheusQueryLanguage *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage `field:"optional" json:"conditionPrometheusQueryLanguage" yaml:"conditionPrometheusQueryLanguage"`
	// condition_threshold block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#condition_threshold MonitoringAlertPolicy#condition_threshold}
	ConditionThreshold *MonitoringAlertPolicyConditionsConditionThreshold `field:"optional" json:"conditionThreshold" yaml:"conditionThreshold"`
}


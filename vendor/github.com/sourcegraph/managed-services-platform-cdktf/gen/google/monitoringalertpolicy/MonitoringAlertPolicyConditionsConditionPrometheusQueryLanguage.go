package monitoringalertpolicy


type MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage struct {
	// The PromQL expression to evaluate.
	//
	// Every evaluation cycle this
	// expression is evaluated at the current time, and all resultant time
	// series become pending/firing alerts. This field must not be empty.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#query MonitoringAlertPolicy#query}
	Query *string `field:"required" json:"query" yaml:"query"`
	// The alerting rule name of this alert in the corresponding Prometheus configuration file.
	//
	// Some external tools may require this field to be populated correctly
	// in order to refer to the original Prometheus configuration file.
	// The rule group name and the alert name are necessary to update the
	// relevant AlertPolicies in case the definition of the rule group changes
	// in the future.
	//
	// This field is optional. If this field is not empty, then it must be a
	// valid Prometheus label name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#alert_rule MonitoringAlertPolicy#alert_rule}
	AlertRule *string `field:"optional" json:"alertRule" yaml:"alertRule"`
	// Alerts are considered firing once their PromQL expression evaluated to be "true" for this long.
	//
	// Alerts whose PromQL expression was not
	// evaluated to be "true" for long enough are considered pending. The
	// default value is zero. Must be zero or positive.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#duration MonitoringAlertPolicy#duration}
	Duration *string `field:"optional" json:"duration" yaml:"duration"`
	// How often this rule should be evaluated.
	//
	// Must be a positive multiple
	// of 30 seconds or missing. The default value is 30 seconds. If this
	// PrometheusQueryLanguageCondition was generated from a Prometheus
	// alerting rule, then this value should be taken from the enclosing
	// rule group.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#evaluation_interval MonitoringAlertPolicy#evaluation_interval}
	EvaluationInterval *string `field:"optional" json:"evaluationInterval" yaml:"evaluationInterval"`
	// Labels to add to or overwrite in the PromQL query result. Label names must be valid.
	//
	// Label values can be templatized by using variables. The only available
	// variable names are the names of the labels in the PromQL result, including
	// "__name__" and "value". "labels" may be empty. This field is intended to be
	// used for organizing and identifying the AlertPolicy
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#labels MonitoringAlertPolicy#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// The rule group name of this alert in the corresponding Prometheus configuration file.
	//
	// Some external tools may require this field to be populated correctly
	// in order to refer to the original Prometheus configuration file.
	// The rule group name and the alert name are necessary to update the
	// relevant AlertPolicies in case the definition of the rule group changes
	// in the future. This field is optional.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#rule_group MonitoringAlertPolicy#rule_group}
	RuleGroup *string `field:"optional" json:"ruleGroup" yaml:"ruleGroup"`
}


package monitoringalertpolicy


type MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage struct {
	// The amount of time that a time series must violate the threshold to be considered failing.
	//
	// Currently, only values that are a
	// multiple of a minute--e.g., 0, 60, 120, or
	// 300 seconds--are supported. If an invalid
	// value is given, an error will be returned.
	// When choosing a duration, it is useful to
	// keep in mind the frequency of the underlying
	// time series data (which may also be affected
	// by any alignments specified in the
	// aggregations field); a good duration is long
	// enough so that a single outlier does not
	// generate spurious alerts, but short enough
	// that unhealthy states are detected and
	// alerted on quickly.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#duration MonitoringAlertPolicy#duration}
	Duration *string `field:"required" json:"duration" yaml:"duration"`
	// Monitoring Query Language query that outputs a boolean stream.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#query MonitoringAlertPolicy#query}
	Query *string `field:"required" json:"query" yaml:"query"`
	// A condition control that determines how metric-threshold conditions are evaluated when data stops arriving. Possible values: ["EVALUATION_MISSING_DATA_INACTIVE", "EVALUATION_MISSING_DATA_ACTIVE", "EVALUATION_MISSING_DATA_NO_OP"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#evaluation_missing_data MonitoringAlertPolicy#evaluation_missing_data}
	EvaluationMissingData *string `field:"optional" json:"evaluationMissingData" yaml:"evaluationMissingData"`
	// trigger block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_alert_policy#trigger MonitoringAlertPolicy#trigger}
	Trigger *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTrigger `field:"optional" json:"trigger" yaml:"trigger"`
}


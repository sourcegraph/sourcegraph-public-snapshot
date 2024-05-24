package alertpolicy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringnotificationchannel"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// newResponseCodeMetricCondition defines the alert condition for response code metrics
// Supports a single Code e.g. 404 or an entire Code Class e.g. 4xx
// Optionally when using a Code Class, codes to exclude can be defined
func newResponseCodeMetricCondition(config *Config) *monitoringalertpolicy.MonitoringAlertPolicyConditions {
	query := responseCodeMQLBuilder(config)

	if config.ResponseCodeMetric.DurationMinutes == nil {
		// default to 1 minute
		config.ResponseCodeMetric.DurationMinutes = pointers.Ptr(uint(1))
	}

	return &monitoringalertpolicy.MonitoringAlertPolicyConditions{
		DisplayName: pointers.Ptr(config.Name),
		ConditionMonitoringQueryLanguage: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage{
			Query:    pointers.Ptr(query),
			Duration: pointers.Stringf("%ds", *config.ResponseCodeMetric.DurationMinutes*60),
			Trigger: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTrigger{
				Count: pointers.Float64(1),
			},
		},
	}
}

// responseCodeMQLBuilder builds the MQL for a response code metric alert
func responseCodeMQLBuilder(config *Config) string {
	var builder strings.Builder

	builder.WriteString(`fetch cloud_run_revision
| metric 'run.googleapis.com/request_count'
| group_by 15s, [value_request_count_aggregate: aggregate(value.request_count)]
| every 15s
| {
`)
	if config.ResponseCodeMetric.CodeClass != nil {
		builder.WriteString("  group_by [metric.response_code, metric.response_code_class],\n")
	} else {
		builder.WriteString("  group_by [metric.response_code],\n")
	}
	builder.WriteString("  [response_code_count_aggregate: aggregate(value_request_count_aggregate)]\n")
	if config.ResponseCodeMetric.Code != nil {
		builder.WriteString(fmt.Sprintf("  | filter (metric.response_code = '%d')\n", *config.ResponseCodeMetric.Code))
	} else {
		builder.WriteString(fmt.Sprintf("  | filter (metric.response_code_class = '%s')\n", *config.ResponseCodeMetric.CodeClass))
	}
	if config.ResponseCodeMetric.ExcludeCodes != nil && len(config.ResponseCodeMetric.ExcludeCodes) > 0 {
		for _, code := range config.ResponseCodeMetric.ExcludeCodes {
			builder.WriteString(fmt.Sprintf("  | filter (metric.response_code != '%s')\n", code))
		}
	}
	builder.WriteString(`; group_by [],
  [value_request_count_aggregate_aggregate: aggregate(value_request_count_aggregate)]
}
| join
| value [response_code_ratio: val(0) / val(1)]
`)
	builder.WriteString(fmt.Sprintf("| condition gt(val(), %s)\n", strconv.FormatFloat(config.ResponseCodeMetric.Ratio, 'f', -1, 64)))
	return builder.String()
}

// notificationChannelIDs collects the IDs of the given notification channels.
// Returns nil if there are no channels.
func notificationChannelIDs(channels []monitoringnotificationchannel.MonitoringNotificationChannel) *[]*string {
	if len(channels) == 0 {
		return nil
	}
	var ids []*string
	for _, c := range channels {
		ids = append(ids, c.Id())
	}
	return &ids
}

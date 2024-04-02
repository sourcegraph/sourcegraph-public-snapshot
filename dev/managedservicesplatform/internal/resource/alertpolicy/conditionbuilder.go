package alertpolicy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type ResourceKind string

const (
	CloudRunService ResourceKind = "cloud-run-service"
	CloudRunJob     ResourceKind = "cloud-run-job"
	CloudRedis      ResourceKind = "cloud-redis"

	// CloudSQL represents a Cloud SQL instance.
	CloudSQL ResourceKind = "cloud-sql"
	// CloudSQLDatabase represents a database within a Cloud SQL instance.
	CloudSQLDatabase ResourceKind = "cloud-sql-database"

	URLUptime ResourceKind = "url-uptime"
)

type TriggerKind int

const (
	// TriggerKindAnyViolation is trigger { count: 1 } - any violation will
	// cause an alert to fire. This is the default.
	TriggerKindAnyViolation TriggerKind = iota
	// TriggerKindAllInViolation is trigger { percent: 100 } - all time series
	// must be in violation for alert to fire.
	TriggerKindAllInViolation
)

// ConditionBuilder are the options available to all GCP alert policy builder
// alerts, i.e. https://console.cloud.google.com/monitoring/alerting/policies/create
// alerts that don't use PromQL/MQL.
type ConditionBuilder struct {
	// ResourceKind identifies what is being monitored. Optional - used for
	// building required filters.
	ResourceKind ResourceKind
	// ResourceName is the identifier for the monitored resource of ResourceKind.
	// Only required if ResourceKind is provided - used for building required
	// filters.
	ResourceName string
	// Filters are additional custom filters.
	Filters map[string]string

	// Aggregations
	GroupByFields []string
	Aligner       Aligner
	Reducer       Reducer
	Period        string

	// Trigger is the strategy for determining if an alert should fire based
	// on the thresholds.
	Trigger TriggerKind
}

// buildGCPAlertBuilderFilters creates the Filter string for alert builder policies
// i.e. non-MQL and non-PromQL alerts
func (c ConditionBuilder) buildFilter() string {
	filters := make([]string, 0)
	for key, val := range c.Filters {
		filters = append(filters, fmt.Sprintf(`%s = "%s"`, key, val))
	}

	// Sort to ensure stable output for testing, because
	// config.ThresholdAggregation.Filters is a map.
	sort.Strings(filters)

	switch c.ResourceKind {
	case CloudRunService:
		filters = append(filters,
			`resource.type = "cloud_run_revision"`,
			fmt.Sprintf(`resource.labels.service_name = starts_with("%s")`, c.ResourceName),
		)
	case CloudRunJob:
		filters = append(filters,
			`resource.type = "cloud_run_job"`,
			fmt.Sprintf(`resource.labels.job_name = starts_with("%s")`, c.ResourceName),
		)
	case CloudRedis:
		filters = append(filters,
			`resource.type = "redis_instance"`,
			fmt.Sprintf(`resource.labels.instance_id = "%s"`, c.ResourceName),
		)
	case CloudSQL:
		filters = append(filters,
			`resource.type = "cloudsql_database"`,
			fmt.Sprintf(`resource.labels.database_id = "%s"`, c.ResourceName))
	case CloudSQLDatabase:
		filters = append(filters,
			`resource.type = "cloudsql_instance_database"`,
			fmt.Sprintf(`resource.labels.resource_id = "%s"`, c.ResourceName))
	case URLUptime:
		filters = append(filters,
			`resource.type = "uptime_url"`,
			fmt.Sprintf(`metric.labels.check_id = "%s"`, c.ResourceName),
		)
	}

	return strings.Join(filters, " AND ")
}

func (c ConditionBuilder) buildThresholdAggregations() []monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdAggregations {
	return []monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdAggregations{
		{
			AlignmentPeriod:    pointers.Ptr(c.Period),
			PerSeriesAligner:   pointers.NonZeroPtr(string(c.Aligner)),
			CrossSeriesReducer: pointers.NonZeroPtr(string(c.Reducer)),
			GroupByFields:      pointers.Ptr(pointers.Slice(c.GroupByFields)),
		},
	}
}

func (c ConditionBuilder) buildAbsentAggregations() []monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionAbsentAggregations {
	return []monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionAbsentAggregations{
		{
			AlignmentPeriod:    pointers.Ptr(c.Period),
			PerSeriesAligner:   pointers.NonZeroPtr(string(c.Aligner)),
			CrossSeriesReducer: pointers.NonZeroPtr(string(c.Reducer)),
			GroupByFields:      pointers.Ptr(pointers.Slice(c.GroupByFields)),
		},
	}
}

func (c ConditionBuilder) buildThresholdTrigger() *monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdTrigger {
	switch c.Trigger {
	case TriggerKindAllInViolation:
		return &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdTrigger{
			Percent: pointers.Float64(100),
		}
	case TriggerKindAnyViolation:
		fallthrough
	default:
		return &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdTrigger{
			Count: pointers.Float64(1),
		}
	}
}

func (c ConditionBuilder) buildAbsentTrigger() *monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionAbsentTrigger {
	switch c.Trigger {
	case TriggerKindAllInViolation:
		return &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionAbsentTrigger{
			Percent: pointers.Float64(100),
		}
	case TriggerKindAnyViolation:
		fallthrough
	default:
		return &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionAbsentTrigger{
			Count: pointers.Float64(1),
		}
	}
}
